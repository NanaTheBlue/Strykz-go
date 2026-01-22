package social

import "errors"

var (
	ErrInviteAlreadySent  = errors.New("invite already sent")
	ErrNotPartyLeader     = errors.New("sender is not the party leader")
	ErrUserAlreadyInParty = errors.New("recipient is already in a party")
)
