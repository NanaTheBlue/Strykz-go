package social

func normalizePair(a, b string) (string, string) {
	if a > b {
		return b, a
	}
	return a, b
}
