package querepo

type QueRepository interface {
	CheckBan(player string) (bool, error)
}
