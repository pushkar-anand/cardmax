package cards

import (
	"github.com/pushkar-anand/build-with-go/http/response"
	"github.com/pushkar-anand/cardmax/internal/cards"
	"log/slog"
	"net/http"
)

// GetAllHandler returns all predefined cards
func GetAllHandler(
	log *slog.Logger,
	jw *response.JSONWriter,
) http.HandlerFunc {
	type Response struct {
		Cards []cards.Card `json:"cards"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		//jw.Ok(r.Context(), w, cards.GetAll())
	}
}

// GetByKeyHandler returns a specific predefined card by key
func GetByKeyHandler(
	log *slog.Logger,
	jw *response.JSONWriter,
) http.HandlerFunc {
	type Response struct {
		cards.Card
	}

	return func(w http.ResponseWriter, r *http.Request) {
		//vars := mux.Vars(r)
		//key := vars["key"]

		//card, found := cards.GetByKey(key)
		//if !found {
		//	jw.WriteProblem(r.Context(), r, w, response.NewProblem().WithStatus(http.StatusNotFound).WithDetail("card not found").Build())
		//	return
		//}

		//jw.Ok(r.Context(), w, card)
	}
}
