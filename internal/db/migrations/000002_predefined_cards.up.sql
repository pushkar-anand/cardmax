CREATE TABLE predefined_cards
(
    -- ID: Unique identifier for each predefined card.
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,

    -- Key: A unique identifier string for the card (e.g., 'HDFC-REGALIA-GOLD').
    card_key            TEXT NOT NULL UNIQUE,

    -- Name: The name of the credit card (e.g., 'HDFC Regalia Gold Credit Card').
    name                TEXT NOT NULL,

    -- Issuer: The bank or institution that issues the card.
    issuer              TEXT NOT NULL,

    -- CardType: The network/type (e.g., 'Visa', 'Mastercard', 'Amex').
    card_type           TEXT NOT NULL,

    -- DefaultRewardRate: The standard reward rate for the card.
    default_reward_rate REAL NOT NULL DEFAULT 0.0,

    -- RewardType: The type of rewards (e.g., 'Points', 'Cashback', 'Miles').
    reward_type         TEXT NOT NULL,

    -- PointValue: The monetary value per point/mile if applicable.
    point_value         REAL NOT NULL DEFAULT 0.0,

    -- AnnualFee: The annual fee for the card.
    annual_fee          INTEGER NOT NULL DEFAULT 0,

    -- AnnualFeeWaiver: Description of any conditions for waiving the annual fee.
    annual_fee_waiver   TEXT,

    -- Created at timestamp
    created_at          DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- Updated at timestamp
    updated_at          DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create rewards rules table for predefined cards
CREATE TABLE predefined_reward_rules
(
    -- ID: Unique identifier for each reward rule.
    id                INTEGER PRIMARY KEY AUTOINCREMENT,

    -- PredefinedCardID: Reference to the predefined card this rule belongs to.
    predefined_card_id INTEGER NOT NULL,

    -- Type: The type of rule (e.g., 'Category' or 'Merchant').
    type              TEXT NOT NULL,

    -- EntityName: The category or merchant name this rule applies to.
    entity_name       TEXT NOT NULL,

    -- RewardRate: The reward rate for this specific rule.
    reward_rate       REAL NOT NULL DEFAULT 0.0,

    -- RewardType: The type of reward (e.g., 'Points', 'Cashback').
    reward_type       TEXT NOT NULL,

    -- Created at timestamp
    created_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- Updated at timestamp
    updated_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- Foreign key reference to predefined_cards table
    FOREIGN KEY (predefined_card_id) REFERENCES predefined_cards (id) ON DELETE CASCADE
);

-- Create an index for faster lookups by card_key
CREATE INDEX idx_predefined_cards_card_key ON predefined_cards (card_key);

-- Create an index for faster lookups by issuer and card name
CREATE INDEX idx_predefined_cards_issuer_name ON predefined_cards (issuer, name);