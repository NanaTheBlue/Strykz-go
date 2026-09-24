package models

import "time"

type ServerStatus string

const (
	ServerCreating ServerStatus = "CREATING"
	ServerReady    ServerStatus = "READY"
	ServerBusy     ServerStatus = "BUSY"
	ServerDead     ServerStatus = "DEAD"
)

type MatchStatus string

const (
	MatchAwaitingServer MatchStatus = "Awaiting_Server"
	MatchAccepted       MatchStatus = "accepted"
	MatchReady          MatchStatus = "ready"
	MatchFinished       MatchStatus = "finished"
	MatchCancelled      MatchStatus = "cancelled"
)

type Gameserver struct {
	ID            string
	IP            string
	Region        string
	Status        ServerStatus
	LastHeartbeat time.Time
	CreatedAt     time.Time
}
