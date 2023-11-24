package riot

func CreateUrl(path string) string {
	return "https://kr.api.riotgames.com" + path + "?api_key=" + apiKey
}
