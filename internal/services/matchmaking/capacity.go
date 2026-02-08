package matchmaking

type CapacityRequester interface {
	Request(region string)
}
