package matchmaking

import (
	"context"
	"encoding/json"
	"errors"

	"fmt"

	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	matchmakingrepo "github.com/nanagoboiler/internal/repository/matchmaking"
	orchestratorrepo "github.com/nanagoboiler/internal/repository/orchestrator"
	"github.com/nanagoboiler/internal/repository/redis"
	"github.com/nanagoboiler/models"
)

// will prob need to add a matchmaking repo
type matchmakingService struct {
	RedisRepo         redis.Store
	pool              *pgxpool.Pool
	matchmakingrepo   matchmakingrepo.MatchmakingRepository
	orchestratorrepo  orchestratorrepo.OrchestratoryRepository
	capacityRequester CapacityRequester
	notifier          Notifier
	serverSpeaker     ServerSpeaker
}

func NewMatchmakingService(redisRepo redis.Store, pool *pgxpool.Pool, matchmakingrepo matchmakingrepo.MatchmakingRepository, orchestratorrepo orchestratorrepo.OrchestratoryRepository, capacityRequester CapacityRequester, notifier Notifier, serverSpeaker ServerSpeaker) Service {
	return &matchmakingService{RedisRepo: redisRepo, pool: pool, matchmakingrepo: matchmakingrepo, orchestratorrepo: orchestratorrepo, capacityRequester: capacityRequester, notifier: notifier, serverSpeaker: serverSpeaker}
}

func (s *matchmakingService) InQue(ctx context.Context, player *models.Player) error {

	// Should Prob do some validation here to make sure none of these players are banned

	err := s.RedisRepo.Que(ctx, "1v1", "us", player)
	if err != nil {
		return err
	}

	return nil
}

func (s *matchmakingService) StartMatchMaking(ctx context.Context, mode string) {

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	s.QueReader(ctx, mode)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.QueReader(ctx, mode)
		}
	}

}

func (s *matchmakingService) QueReader(ctx context.Context, mode string) {
	regions := []string{"us"}

	for _, region := range regions {
		queueKey := fmt.Sprintf("queue:%s:%s", mode, region)

		matchCandidates, err := s.RedisRepo.DeQue(ctx, mode, region, 2)
		if err != nil {
			log.Printf("Error reading from queue %s: %v", queueKey, err)
			continue
		}

		if len(matchCandidates) < 2 {

			continue
		}

		go func() {
			s.CreateMatch(ctx, matchCandidates, region)
		}()

	}

}

func (s *matchmakingService) CreateMatch(ctx context.Context, matchCanidates []*models.Player, region string) error {

	var matchID string
	deadline := time.Now().Add(30 * time.Second)

	err := WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		repo := matchmakingrepo.NewMatchmakingRepository(tx)

		id, err := repo.CreateMatch(ctx, deadline, region)
		if err != nil {
			log.Printf("ERROR: repo.CreateMatch error: %v", err)
			return err
		}

		matchID = id

		err = repo.InsertPlayers(ctx, matchCanidates, matchID)
		if err != nil {
			log.Printf("ERROR: repo.InsertPlayers error: %v", err)
			return err
		}
		return nil
	})

	if err != nil {
		log.Printf("ERROR: Transaction result error: %v", err)
		return err
	}

	for _, p := range matchCanidates {

		payload := map[string]any{
			"match_id":   matchID,
			"region":     region,
			"expires_in": 30,
		}

		dataBytes, _ := json.Marshal(payload)

		notif := models.Notification{
			ID:          uuid.NewString(),
			SenderID:    "system",
			RecipientID: p.Player_id,
			Type:        models.NotificationType("MatchFound"),
			Data:        string(dataBytes),
			Status:      "unread",
			CreatedAt:   time.Now(),
		}

		notifid, err := s.notifier.CreateNoPublishNotification(ctx, notif)
		if err != nil {
			log.Printf("failed to notify %s: %v", p.Player_id, err)
		}
		if notifid == "" {
			log.Printf("ERROR: notifid NULL MatchMaking Service")
		}
	}

	return nil

}

func (s *matchmakingService) ReconcileAwaitingMatches(ctx context.Context) {
	matches, err := s.matchmakingrepo.GetMatchesByStatus(ctx, models.AwaitingServer)
	if err != nil {
		log.Printf("ERROR: GetMatchesByStatus Error : %v", err)
	}

	for _, match := range matches {
		server, err := s.orchestratorrepo.AcquireReadyServer(ctx, match.Region)
		if err != nil {
			log.Printf("ERROR: AcquireReadyServer Error : %v", err)
			continue
		}

		if server == nil {
			continue
		}
		if err := s.finalizeMatch(ctx, match.ID, server.ID); err != nil {
			log.Printf("ERROR: finalizeWithServer failed for match %s: %v", match.ID, err)
			continue
		}
	}
}

func (s *matchmakingService) StartReconciler(ctx context.Context) {

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.ReconcileAwaitingMatches(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (s *matchmakingService) finalizeMatch(ctx context.Context, matchID string, serverID string) error {
	err := WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		repo := matchmakingrepo.NewMatchmakingRepository(tx)

		if err := repo.AssignServerToMatch(ctx, matchID, serverID); err != nil {
			return err
		}

		if err := repo.UpdateMatchStatus(ctx, matchID, "ready"); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	players, err := s.matchmakingrepo.GetMatchPlayers(ctx, matchID)
	if err != nil {
		return err
	}

	steamIDs := make([]string, len(players))
	for i, p := range players {
		steamIDs[i] = p.Player_steamid
	}

	if err := s.serverSpeaker.ReloadWhitelist(serverID, steamIDs); err != nil {
		return err
	}

	for _, p := range players {
		payload := map[string]any{
			"match_id":  matchID,
			"server_id": serverID,
			"type":      "match_ready",
		}
		dataBytes, _ := json.Marshal(payload)
		notif := models.Notification{
			ID:          uuid.NewString(),
			SenderID:    "system",
			RecipientID: p.Player_id,
			Type:        models.NotificationType("MatchReady"),
			Data:        string(dataBytes),
			Status:      "unread",
			CreatedAt:   time.Now(),
		}
		if _, err := s.notifier.CreateNoPublishNotification(ctx, notif); err != nil {
			log.Printf("failed to notify player %s: %v", p.Player_id, err)
		}
	}

	return nil
}

// as players join update status if kicked update status
func (s *matchmakingService) updatePlayerStatus(ctx context.Context, matchID string, player models.Player, status string) error {
	err := s.matchmakingrepo.UpdatePlayer(ctx, player, matchID, status)
	if err != nil {
		return err
	}
	return nil
}

func (s *matchmakingService) GetPlayerByID(ctx context.Context, userID string) (models.Player, error) {
	player, err := s.matchmakingrepo.GetPlayerByID(ctx, userID)
	if err != nil {
		return models.Player{}, err
	}
	return player, nil
}

func (s *matchmakingService) ConfirmMatch(ctx context.Context, player models.Player, matchID string, region string) error {
	var shouldFinalize bool

	err := WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		repo := matchmakingrepo.NewMatchmakingRepository(tx)

		match, err := repo.GetMatch(ctx, matchID)
		if err != nil {
			return err
		}

		if time.Now().After(*match.AcceptDeadline) {
			return errors.New("accept window expired")
		}

		if err := repo.UpdatePlayer(ctx, player, matchID, "accepted"); err != nil {
			return err
		}

		allAccepted, err := repo.AreAllPlayersAccepted(ctx, matchID)
		if err != nil {
			return err
		}

		if allAccepted {
			if err := repo.UpdateMatchStatus(ctx, matchID, "accepted"); err != nil {
				return err
			}

			shouldFinalize = true
		}

		return nil
	})

	if err != nil {
		return err
	}

	if shouldFinalize {
		server, err := s.orchestratorrepo.AcquireReadyServer(ctx, region)
		if err != nil {
			return err
		}
		if server == nil {
			s.capacityRequester.Request(region)

			return s.matchmakingrepo.UpdateMatchStatus(ctx, matchID, models.AwaitingServer)
		}

		return s.finalizeMatch(ctx, matchID, server.ID)
	}
	// should improve this just prototype rn
	return nil
}

func WithTx(
	ctx context.Context,
	pool *pgxpool.Pool,
	fn func(tx pgx.Tx) error,
) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
