package models

import "time"

type ServerStatus string

const (
	ServerCreating ServerStatus = "CREATING"
	ServerReady    ServerStatus = "READY"
	ServerBusy     ServerStatus = "BUSY"
	ServerDead     ServerStatus = "DEAD"
)

type Gameserver struct {
	ID            string
	Region        string
	Status        ServerStatus
	LastHeartbeat time.Time
	CreatedAt     time.Time
}
