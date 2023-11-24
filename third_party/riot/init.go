package riot

import "os"

var apiKey string

func Init() {
	apiKey = os.Getenv("RIOT_API_KEY")
}
