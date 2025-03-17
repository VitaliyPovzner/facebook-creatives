package service

import (
	"encoding/json"
	"fmt"
	"os"

	"facebook-creatives/internal/utils"
	"github.com/rs/zerolog/log"
	"time"
	"errors"
)

type FacebookService struct {
	APIVersion  string
	AccessToken string
}

type AdAccount struct {
	ID                     string `json:"account_id"`
	Name                   string `json:"name"`
	TimezoneOffsetHoursUTC int    `json:"timezone_offset_hours_utc"`
	TimezoneName           string `json:"timezone_name"`
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

func (s *FacebookService) FetchCreativeData() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for {
		log.Info().Msg("Fetching Ad Accounts...")

		accounts, err := s.GetAdAccounts()
		if err != nil {
			log.Error().Err(err).Msg("Failed to fetch ad accounts")
		}

		for _, account := range accounts {
			log.Info().Msgf("Fetching Ad Insights for Account: %s (%s)", account.Name, account.ID)

			insights, err := s.FetchAdInsights(account)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to fetch insights for account %s", account.Name)
				continue
			}

			log.Info().Msgf("Fetched %d insights for account %s", len(insights), account.Name)
		}

		<-ticker.C
	}
}


func (s *FacebookService) FetchAdInsights(account AdAccount) ([]AdInsight, error) {
	adAccountId := account.ID
	log.Info().Msgf("Fetching insights for Ad Account: %s (%s)", account.Name, adAccountId)

	// Step 1: Create Async Job for Insights
	jobID, err := s.createAdInsightsJob(adAccountId)
	if err != nil {
		return nil, err
	}

	// Step 2: Poll until job is completed
	if err := s.waitForJobCompletion(jobID, account.Name); err != nil {
		return nil, err
	}

	// Step 3: Fetch insights using the job ID
	insights, err := s.fetchAdInsightsResults(jobID)
	if err != nil {
		return nil, err
	}

	// Step 4: Add additional account data to results
	for i := range insights {
		insights[i].TimezoneOffsetHoursUTC = account.TimezoneOffsetHoursUTC
		insights[i].AccountID = account.ID
	}

	return insights, nil
}

func (s *FacebookService) createAdInsightsJob(adAccountId string) (string, error) {
	url := fmt.Sprintf("https://graph.facebook.com/%s/act_%s/insights?access_token=%s",
		s.APIVersion, adAccountId, s.AccessToken)

	data := map[string]string{
		"level":                         "ad",
		"limit":                         "300",
		"fields":                        "ad_id,account_name,outbound_clicks,spend,cost_per_inline_link_click,cost_per_unique_outbound_click,cost_per_unique_inline_link_click,cost_per_unique_click,campaign_id,adset_id,impressions,actions,ad_name",
		"use_unified_attribution_setting": "true",
		"date_preset":                   "yesterday",
		"breakdowns":                     "hourly_stats_aggregated_by_advertiser_time_zone",
	}

	responseData, err := utils.PostRequest(url, data)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create ad insights job")
		return "", err
	}

	var response struct {
		ReportRunID string `json:"report_run_id"`
	}

	if err := json.Unmarshal(responseData, &response); err != nil {
		log.Error().Err(err).Msg("Failed to parse job creation response")
		return "", err
	}

	log.Info().Msgf("Created Async Job ID: %s for account %s", response.ReportRunID, adAccountId)
	return response.ReportRunID, nil
}

func (s *FacebookService) waitForJobCompletion(jobID, accountName string) error {
	for {
		time.Sleep(3 * time.Second)

		url := fmt.Sprintf("https://graph.facebook.com/%s/%s?access_token=%s",
			s.APIVersion, jobID, s.AccessToken)

		responseData, err := utils.GetRequest(url)
		if err != nil {
			log.Error().Err(err).Msgf("Error checking job status for account %s", accountName)
			return err
		}

		var statusResponse struct {
			AsyncStatus string `json:"async_status"`
		}

		if err := json.Unmarshal(responseData, &statusResponse); err != nil {
			log.Error().Err(err).Msg("Failed to parse job status response")
			return err
		}

		log.Info().Msgf("Facebook responded with Job Status: %s for ad account %s", statusResponse.AsyncStatus, accountName)

		if statusResponse.AsyncStatus == "Job Completed" {
			return nil
		} else if statusResponse.AsyncStatus == "Job Failed" {
			return errors.New("facebook job failed")
		}
	}
}

func (s *FacebookService) fetchAdInsightsResults(jobID string) ([]AdInsight, error) {
	url := fmt.Sprintf("https://graph.facebook.com/%s/%s/insights?access_token=%s",
		s.APIVersion, jobID, s.AccessToken)

	responseData, err := utils.PaginateRequest(url)
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch ad insights results")
		return nil, err
	}

	var insights []AdInsight
	if err := json.Unmarshal(responseData, &insights); err != nil {
		log.Error().Err(err).Msg("Failed to parse ad insights response")
		return nil, err
	}

	return insights, nil
}

type AdInsight struct {
	AdID                      string           `json:"ad_id"`
	AccountName               string           `json:"account_name"`
	OutboundClicks            []ActionMetric   `json:"outbound_clicks,omitempty"`
	Spend                     string           `json:"spend"`
	CostPerInlineLinkClick    string           `json:"cost_per_inline_link_click,omitempty"`
	CostPerUniqueOutboundClick string           `json:"cost_per_unique_outbound_click,omitempty"`
	CostPerUniqueInlineLinkClick string        `json:"cost_per_unique_inline_link_click,omitempty"`
	CostPerUniqueClick        string           `json:"cost_per_unique_click,omitempty"`
	CampaignID                string           `json:"campaign_id"`
	AdsetID                   string           `json:"adset_id"`
	Impressions               string           `json:"impressions"` // Facebook returns it as a string
	Actions                   []ActionMetric   `json:"actions,omitempty"`
	AdName                    string           `json:"ad_name"`
	DateStart                 string           `json:"date_start"`
	DateStop                  string           `json:"date_stop"`
	HourlyStats               string           `json:"hourly_stats_aggregated_by_advertiser_time_zone"`
	AccountID                 string           `json:"account_id"`
		TimezoneOffsetHoursUTC   int    `json:"timezone_offset_hours_utc"`
}

type ActionMetric struct {
	ActionType string `json:"action_type"`
	Value      string `json:"value"`
}

