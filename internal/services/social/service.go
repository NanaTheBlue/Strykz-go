package social

import (
	"context"
	"errors"

	socialrepo "github.com/nanagoboiler/internal/repository/social"
	"github.com/nanagoboiler/internal/services/notifications"
	"github.com/nanagoboiler/models"
)

type socialService struct {
	notificationservice notifications.Service
	socialrepo          socialrepo.SocialRepository
}

func NewsocialService(notificationservice notifications.Service, socialrepo socialrepo.SocialRepository) Service {
	return &socialService{
		notificationservice: notificationservice,
		socialrepo:          socialrepo,
	}
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

func (s *socialService) RejectNotification(ctx context.Context, notif models.Notification) error {
	err := s.notificationservice.DeleteNotification(ctx, notif.ID)
	if err != nil {
		return err
	}
	return nil
}

func (s *socialService) CreateParty(ctx context.Context, userID string) (string, error) {
	partyID, err := s.CreateParty(ctx, userID)
	if err != nil {
		return "", err
	}

	return partyID, nil
}

func (s *socialService) PartyInvite(ctx context.Context, partyInviteReq models.PartyInviteRequest) error {

	return nil
}
