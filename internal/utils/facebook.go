package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"
)

func PaginateRequest(url string) ([]byte, error) {
	var allData []map[string]interface{}

	for url != "" {
		resp, err := http.Get(url)
		if err != nil {
			log.Error().Err(err).Msg("Error making request to Facebook API")
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("failed to fetch data: status code %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Error().Err(err).Msg("Failed to read response body")
			return nil, err
		}

		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err != nil {
			log.Error().Err(err).Msg("Failed to parse JSON response")
			return nil, err
		}

		if data, ok := result["data"].([]interface{}); ok {
			for _, item := range data {
				if itemMap, ok := item.(map[string]interface{}); ok {
					allData = append(allData, itemMap)
				}
			}
		}

		if paging, exists := result["paging"].(map[string]interface{}); exists {
			if next, exists := paging["next"].(string); exists {
				url = next
			} else {
				url = ""
			}
		} else {
			url = ""
		}
	}

	responseJSON, err := json.Marshal(allData)
	if err != nil {
		return nil, err
	}

	return responseJSON, nil
}
