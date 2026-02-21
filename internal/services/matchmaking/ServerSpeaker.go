package matchmaking

type ServerSpeaker interface {
	ReloadWhitelist(serverID string, steamIDs []string) error
}
