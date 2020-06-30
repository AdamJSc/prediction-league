package football_data_org

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"prediction-league/service/internal/datastore"
	"prediction-league/service/internal/models"
	"strings"
)

const (
	MetaKeyPlayedGames    = "playedGames"
	MetaKeyPoints         = "points"
	MetaKeyGoalsFor       = "goalsFor"
	MetaKeyGoalsAgainst   = "goalsAgainst"
	MetaKeyGoalDifference = "goalDifference"
)

const baseURL = "https://api.football-data.org"

// Client defines our football-data.org API client
type Client struct {
	baseURL  string
	apiToken string
}

// RetrieveLatestStandingsBySeason implements this method on the clients.FootballDataSource interface
func (c *Client) RetrieveLatestStandingsBySeason(ctx context.Context, s *models.Season) (*models.Standings, error) {
	var url = getClientFullURL(c, fmt.Sprintf("/v2/competitions/%s/standings?season=%d&standingType=TOTAL",
		s.ClientID.Value(),
		s.Active.From.Year()),
	)

	httpResponse, err := getResponse(c, ctx, url)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		return nil, err
	}

	var standingsResponse competitionStandingsGetResponse
	if err := json.Unmarshal(body, &standingsResponse); err != nil {
		return nil, err
	}

	if len(standingsResponse.Standings) != 1 {
		return nil, fmt.Errorf("expected standings length of 1, got %d", len(standingsResponse.Standings))
	}

	standings := models.Standings{
		SeasonID:    s.ID,
		RoundNumber: standingsResponse.Season.CurrentMatchday,
	}
	for _, tableElem := range standingsResponse.Standings[0].Table {
		ranking, err := tableElem.toRankingWithMeta()
		if err != nil {
			return nil, err
		}
		standings.Rankings = append(standings.Rankings, ranking)
	}

	return &standings, nil
}

// NewClient generates a new Client
func NewClient(apiToken string) *Client {
	return &Client{
		apiToken: apiToken,
		baseURL:  baseURL,
	}
}

// getClientFullURL appends the provided endpoint to the provided Client's base URL
func getClientFullURL(c *Client, endpoint string) string {
	return fmt.Sprintf("%s/%s", c.baseURL, strings.Trim(endpoint, "/"))
}

// getResponse performs a GET request to the provided URL and returns a response object
func getResponse(c *Client, ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("X-Auth-Token", c.apiToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// competitionStandingsGetResponse defines the expected payload structure of the request to retrieve the current standings
type competitionStandingsGetResponse struct {
	Season struct {
		CurrentMatchday int `json:"currentMatchday"`
	} `json:"season"`
	Standings []struct {
		Table []tableElem `json:"table"`
	} `json:"standings"`
}

// tableElem defines the nested payload structure within the response that retrieves the current standings
type tableElem struct {
	Position int `json:"position"`
	Team     struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"team"`
	PlayedGames    int `json:"playedGames"`
	GoalsFor       int `json:"goalsFor"`
	GoalsAgainst   int `json:"goalsAgainst"`
	GoalDifference int `json:"goalDifference"`
	Points         int `json:"points"`
}

// toRankingWithMeta transforms a tableElem object to a more abstracted RankingWithMeta object
func (t *tableElem) toRankingWithMeta() (models.RankingWithMeta, error) {
	r := models.NewRankingWithMeta()

	team, err := datastore.Teams.GetByResourceID(models.TeamIdentifier{TeamID: t.Team.ID})
	if err != nil {
		return models.RankingWithMeta{}, err
	}

	r.ID = team.ID
	r.Position = t.Position
	r.MetaData[MetaKeyPlayedGames] = t.PlayedGames
	r.MetaData[MetaKeyPoints] = t.Points
	r.MetaData[MetaKeyGoalsFor] = t.GoalsFor
	r.MetaData[MetaKeyGoalsAgainst] = t.GoalsAgainst
	r.MetaData[MetaKeyGoalDifference] = t.GoalDifference

	return r, nil
}
