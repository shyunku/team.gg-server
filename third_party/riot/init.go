package riot

import "os"

var (
	apiKey   string
	ApiCalls int
)

func Init() {
	apiKey = os.Getenv("RIOT_API_KEY")
	ApiCalls = 0
}

func incrementApiCalls() {
	ApiCalls++
}
