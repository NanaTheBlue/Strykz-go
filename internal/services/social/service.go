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

func NewsocialService(notificationservice notifications.Service) Service {
	return &socialService{
		notificationservice: notificationservice,
	}
}

func (s *socialService) SendFriendRequest(ctx context.Context, notif models.Notification) error {
	userID, friendID := normalizePair(notif.SenderID, notif.RecipientID)

	err := s.socialrepo.CreateFriendRequest(ctx, userID, friendID)
	if err != nil {
		return err
	}

	err = s.notificationservice.SendNotification(ctx, notif)
	if err != nil {
		return err
	}

	return nil
}

func (s *socialService) BlockUser(ctx context.Context, req models.BlockRequest) error {

	//make this a transaction

	if req.BlockerID == req.BlockedID {
		return errors.New("cannot block yourself")
	}
	exists, err := s.socialrepo.IsBlocked(ctx, req.BlockerID, req.BlockedID)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	err = s.socialrepo.BlockUser(ctx, req)
	if err != nil {
		return err
	}

	notif := models.Notification{
		SenderID:    req.BlockerID,
		RecipientID: req.BlockedID,
		Type:        models.BlockNotification,
		Data:        "",
		Status:      "UnRead", // will make this a type later
	}
	err = s.notificationservice.SendNotification(ctx, notif)
	if err != nil {
		return err
	}

	err = s.socialrepo.RemoveFriend(ctx, req.BlockerID, req.BlockedID)
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
		err = s.socialrepo.DeleteFriendRequest(ctx, notif.SenderID, notif.RecipientID)
		if err != nil {
			return err
		}
		err = s.notificationservice.SendNotification(ctx, notif)
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
