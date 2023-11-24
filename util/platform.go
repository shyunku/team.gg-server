package util

import "strings"

func ShortenSummonerName(name string) string {
	// delete space
	name = strings.ReplaceAll(name, " ", "")
	// to lowercase
	name = strings.ToLower(name)
	return name
}
