package util

import "strings"

func ShortenSummonerName(name string) string {
	// delete space
	name = strings.ReplaceAll(name, " ", "")
	// to lowercase
	name = strings.ToLower(name)
	return name
}

func LogisticNormalize(x float64, factor float64) float64 {
	return 1 / (1 + (x / factor))
}
