package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

func PaginateRequest(url string) ([]byte, error) {
	allData := make([]map[string]interface{}, 0, 1000)
	startTime := time.Now()

	for url != "" {
		resp, err := http.Get(url)
		if err != nil {
			log.Error().Err(err).Msg("Error making request to Facebook API")
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to fetch data: status code %d", resp.StatusCode)
		}

		decoder := json.NewDecoder(resp.Body)
		var result map[string]interface{}
		if err := decoder.Decode(&result); err != nil {
			resp.Body.Close()
			log.Error().Err(err).Msg("Failed to decode JSON response")
			return nil, err
		}
		resp.Body.Close()

		// Process the "data" items
		if data, ok := result["data"].([]interface{}); ok {
			for _, item := range data {
				if itemMap, ok := item.(map[string]interface{}); ok {
					allData = append(allData, itemMap)
				}
			}
		}

		// Update the URL for the next page, if available
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

	duration := time.Since(startTime)
	log.Printf("PaginateRequest completed, time taken: %v", duration)
	return responseJSON, nil
}
