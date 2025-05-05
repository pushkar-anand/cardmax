package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/pushkar-anand/build-with-go/logger"
	"github.com/pushkar-anand/cardmax/internal/cards"
	"github.com/pushkar-anand/cardmax/internal/db/models"
	"log/slog"
)

// PopulatePredefinedCards adds or updates predefined cards from the parsed JSON data to the database
func (d *DB) PopulatePredefinedCards(ctx context.Context, log *slog.Logger, cardList []*cards.Card) error {
	log.InfoContext(ctx, "populating predefined cards from static data", slog.Int("count", len(cardList)))

	// Start a transaction for bulk insert
	tx, err := d.Conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Create the queries instance with the transaction
	q := d.Queries.WithTx(tx)

	// If we return with an error, rollback the transaction
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				log.ErrorContext(ctx, "failed to rollback transaction", logger.Error(rbErr))
			}
		}
	}()

	for _, card := range cardList {
		var (
			dbCard *models.PredefinedCard
			err    error
		)

		// Check if the card already exists
		dbCard, err = q.GetPredefinedCardByKey(ctx, card.Key)
		if err != nil && errors.Is(err, sql.ErrNoRows) {

			// Card doesn't exist, create it
			var annualFeeWaiver *string
			if card.AnnualFeeWaiver != "" {
				annualFeeWaiver = &card.AnnualFeeWaiver
			}

			newCard, err := q.CreatePredefinedCard(ctx, models.CreatePredefinedCardParams{
				CardKey:           card.Key,
				Name:              card.Name,
				Issuer:            card.Issuer,
				CardType:          card.CardType,
				DefaultRewardRate: card.DefaultRewardRate,
				RewardType:        card.RewardType,
				PointValue:        card.PointValue,
				AnnualFee:         int64(card.AnnualFee),
				AnnualFeeWaiver:   annualFeeWaiver,
			})

			dbCard = newCard

			if err != nil {
				return fmt.Errorf("failed to create predefined card %s: %w", card.Key, err)
			}

			log.InfoContext(ctx, "created predefined card in database",
				slog.String("key", card.Key),
				slog.String("name", card.Name),
				slog.String("issuer", card.Issuer))
		}

		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("error checking if predefined card exists %s: %w", card.Key, err)
		}

		// Add reward rules for this card
		for _, rule := range card.RewardRules {
			_, err := q.CreatePredefinedRewardRule(ctx, models.CreatePredefinedRewardRuleParams{
				PredefinedCardID: dbCard.ID,
				Type:             rule.Type,
				EntityName:       rule.EntityName,
				RewardRate:       rule.RewardRate,
				RewardType:       rule.RewardType,
			})

			if err != nil {
				return fmt.Errorf("failed to create reward rule for card %s, entity %s: %w",
					card.Key, rule.EntityName, err)
			}
		}
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.InfoContext(ctx, "finished populating predefined cards from static data")
	return nil
}
