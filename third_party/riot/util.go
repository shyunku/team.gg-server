package riot

import "fmt"

func CreateUrl(region string, path string) string {
	return fmt.Sprintf("https://%s.api.riotgames.com%s?api_key=%s", region, path, apiKey)
}

func CreateUrlWithQuery(region string, path string, queries map[string]interface{}) string {
	query := "?"
	for key, value := range queries {
		query += key + "=" + value.(string) + "&"
	}
	return fmt.Sprintf("https://%s.api.riotgames.com%s?api_key=%s", region, path, apiKey)
}
