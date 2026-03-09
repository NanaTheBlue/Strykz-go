package social

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nanagoboiler/internal/repository/redis"
	socialrepo "github.com/nanagoboiler/internal/repository/social"
	"github.com/nanagoboiler/internal/services/notifications"
	"github.com/nanagoboiler/models"
)

type socialService struct {
	notificationservice notifications.Service
	socialrepo          socialrepo.SocialRepository
	store               redis.Store
	pool                *pgxpool.Pool
}

func NewsocialService(notificationservice notifications.Service, pool *pgxpool.Pool, socialrepo socialrepo.SocialRepository, store redis.Store) Service {
	return &socialService{
		notificationservice: notificationservice,
		socialrepo:          socialrepo,
		store:               store,
	}
}

func (s *socialService) ReportUser(ctx context.Context, reportreq models.ReportRequestInput) error {

	err := s.socialrepo.AddReport(ctx, reportreq)
	if err != nil {
		return fmt.Errorf("failed to add report: %w", err)
	}

	return nil
}

func (s *socialService) SendFriendRequest(ctx context.Context, friendreq models.FriendRequestInput) error {

	err := s.socialrepo.CreateFriendRequest(ctx, friendreq)
	if err != nil {
		return err
	}

	notif := models.Notification{
		SenderID:    friendreq.SenderID,
		RecipientID: friendreq.RecipientID,
		Type:        models.FriendRequest,
		Data:        "",
		Status:      "Pending",
	}

	err = s.notificationservice.PublishNotification(ctx, notif)
	if err != nil {
		return err
	}

	return nil
}

func (s *socialService) BlockUser(ctx context.Context, blocker string, blocked string) error {
	if blocker == blocked {
		return errors.New("cannot block yourself")
	}

	err := s.socialrepo.BlockUser(ctx, blocker, blocked)
	if err != nil {
		return err
	}

	return nil
}

func (s *socialService) AcceptNotification(ctx context.Context, notif models.Notification) error {

	switch notif.Type {
	case models.FriendRequest:
		err := s.socialrepo.AddFriend(ctx, notif.SenderID, notif.RecipientID)
		if err != nil {
			return err
		}
		err = s.notificationservice.CreateAndPublishNotification(ctx, notif)
		if err != nil {
			return err
		}
	case models.PartyInvite:

	default:

	}
	return nil
}

func (s *socialService) AcceptPartyInvite(ctx context.Context, partyinvitereq models.PartyInviteRequest) error {
	//This is Wrong
	err := s.socialrepo.AddPartyMember(ctx, partyinvitereq)
	if err != nil {
		return err
	}
	return nil
}

func (s *socialService) RejectNotification(ctx context.Context, notifID string) error {
	err := s.notificationservice.DeleteNotification(ctx, notifID)
	if err != nil {
		return err
	}
	return nil
}

func (s *socialService) CreateParty(ctx context.Context, userID string) (string, error) {
	partyID, err := s.socialrepo.CreateParty(ctx, userID)
	if err != nil {
		return "", err
	}

	return partyID, nil
}

func (s *socialService) PartyInvite(ctx context.Context, partyInviteReq models.PartyInviteRequest) error {

	err := WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		repo := socialrepo.NewSocialRepository(tx)
		Leader, err := repo.CheckPartyLeader(
			ctx,
			partyInviteReq.PartyID,
		)
		if err != nil {
			return err
		}
		if Leader != partyInviteReq.SenderID {
			return ErrNotPartyLeader
		}

		blocked, err := repo.IsMutuallyBlocked(ctx, partyInviteReq.SenderID, partyInviteReq.RecipientID)
		if blocked {
			return errors.New("cannot send invite")
		}
		if err != nil {
			return err
		}

		friends, err := repo.IsFriends(ctx, partyInviteReq.SenderID, partyInviteReq.RecipientID)
		if err != nil {
			return err
		}

		if !friends {
			return errors.New("cannot send party invite to user you not friends with")
		}

		exists, err := repo.AddPartyInvite(ctx, partyInviteReq)
		if err != nil {
			return err
		}
		if exists {
			return errors.New("party invite already exists")
		}

		return nil
	})

	//

	data, err := json.Marshal(map[string]any{
		"party_id":  partyInviteReq.PartyID,
		"sender_id": partyInviteReq.SenderID,
	})
	if err != nil {
		return err
	}

	notif := models.Notification{
		SenderID:    partyInviteReq.SenderID,
		Type:        "party_invite",
		RecipientID: partyInviteReq.RecipientID,
		Data:        string(data),
		CreatedAt:   time.Now(),
	}

	err = s.notificationservice.PublishNotification(ctx, notif)
	if err != nil {
		return err
	}

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
