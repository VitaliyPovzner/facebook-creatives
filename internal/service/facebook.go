package service

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	"facebook-creatives/internal/utils"
)


type FacebookService struct {
	APIVersion  string
	AccessToken string
}

type AdAccount struct {
	ID                    string `json:"account_id"`
	Name                  string `json:"name"`
	TimezoneOffsetHoursUTC int    `json:"timezone_offset_hours_utc"`
	TimezoneName          string `json:"timezone_name"`
}



func NewFacebookService(accessToken string) *FacebookService {
	apiVersion := os.Getenv("FB_API_VERSION")
	if apiVersion == "" {
		apiVersion = "v18.0" 
	}
	return &FacebookService{
		APIVersion:  apiVersion,
		AccessToken: accessToken,
	}
}

func (s *FacebookService) GetAdAccounts() ([]AdAccount, error) {
	url := fmt.Sprintf(
		"https://graph.facebook.com/%s/me/adaccounts?fields=account_id,name,timezone_offset_hours_utc,timezone_name&access_token=%s",
		s.APIVersion, s.AccessToken,
	)

	responseData, err := utils.PaginateRequest(url)
	if err != nil {
		log.Error().Err(err).Msg("Error fetching ad accounts")
		return nil, err
	}

	var accounts []AdAccount
	log.Info().Msgf("Raw Response: %s", string(responseData))
	if err := json.Unmarshal(responseData, &accounts); err != nil {
		log.Error().Err(err).Msg("Failed to parse ad accounts response")
		return nil, err
	}

	return accounts, nil
}
