package matchmakingapi

import (
	"net/http"

	"github.com/nanagoboiler/internal/services/matchmaking"
)

func Que(s matchmaking.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}
