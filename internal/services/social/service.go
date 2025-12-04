package social

import (
	"context"

	socialrepo "github.com/nanagoboiler/internal/repository/social"
	"github.com/nanagoboiler/internal/services/notifications"
	"github.com/nanagoboiler/models"
)

type socialService struct {
	notificationservice notifications.Service
	socialrepo socialrepo.SocialRepository
}

func NewsocialService(notificationservice notifications.Service) Service {
	return &socialService{
		notificationservice: notificationservice,
	}
}

func (s *socialService) SendFriendRequest(ctx context.Context, notif models.Notification) error {
	err := s.notificationservice.SendNotification(ctx, notif)
	if err != nil {
		return err
	}

	return nil
}

func (s *socialService) AcceptNotification(ctx context.Context, notif models.Notification) error {

	if notif.Type == "FriendRequest" {
		// add friends
		err := s.socialrepo.AddFriend(ctx, notif)
		if err != nil {
			return err
		}
	} else if notif.Type == "PartyInvite" {-

		// Join Party

	}
	return nil
}

func (s *socialService) RejectNotification(ctx context.Context, notif models.Notification) error {
	err := s.notificationservice

	return nil
}
