package matchmakingapi

import (
	"net/http"

	"github.com/nanagoboiler/internal/services/auth"
	"github.com/nanagoboiler/internal/services/matchmaking"
	"github.com/nanagoboiler/models"
)

func Que(s matchmaking.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() { _ = r.Body.Close() }()
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

		user, ok := r.Context().Value(auth.UserContextKey).(*models.User)
		if !ok || user == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		playerFromDB, err := s.GetPlayerByID(r.Context(), user.ID)
		if err != nil {
			http.Error(w, "failed to load player", http.StatusInternalServerError)
			return
		}

		player := &models.Player{
			Player_id:      playerFromDB.Player_id,
			Player_steamid: playerFromDB.Player_steamid,
		}

		if err := s.InQue(r.Context(), player); err != nil {
			http.Error(w, "failed to join queue", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
