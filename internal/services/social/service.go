package social

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/nanagoboiler/internal/repository/redis"
	socialrepo "github.com/nanagoboiler/internal/repository/social"
	"github.com/nanagoboiler/internal/services/notifications"
	"github.com/nanagoboiler/models"
)

type socialService struct {
	notificationservice notifications.Service
	socialrepo          socialrepo.SocialRepository
	store               redis.Store
}

func NewsocialService(notificationservice notifications.Service, socialrepo socialrepo.SocialRepository, store redis.Store) Service {
	return &socialService{
		notificationservice: notificationservice,
		socialrepo:          socialrepo,
		store:               store,
	}
}

var (
	ErrInviteAlreadySent  = errors.New("invite already sent")
	ErrNotPartyLeader     = errors.New("sender is not the party leader")
	ErrUserAlreadyInParty = errors.New("recipient is already in a party")
)

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

func (s *socialService) BlockUser(ctx context.Context, req models.BlockRequest) error {

	if req.BlockerID == req.BlockedID {
		return errors.New("cannot block yourself")
	}

	err := s.socialrepo.BlockUser(ctx, req)
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

	Leader, err := s.socialrepo.CheckPartyLeader(
		ctx,
		partyInviteReq.PartyID,
	)
	if err != nil {
		return err
	}
	if Leader != partyInviteReq.SenderID {
		return ErrNotPartyLeader
	}

	// should prob make something to check if they are friends with party leader

	inviteKey := fmt.Sprintf(
		"party:invite:%s:%s",
		partyInviteReq.PartyID,
		partyInviteReq.RecipientID,
	)

	ok, err := s.store.AddNX(
		ctx,
		inviteKey,
		partyInviteReq.SenderID,
		6*time.Second,
	)

	if err != nil {
		return err
	}
	if !ok {
		return ErrInviteAlreadySent
	}

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
