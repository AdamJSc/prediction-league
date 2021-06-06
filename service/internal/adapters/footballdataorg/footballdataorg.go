package footballdataorg

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"prediction-league/service/internal/adapters"
	"prediction-league/service/internal/domain"
)

const baseURL = "https://api.football-data.org"

// Client defines our football-data.org API client
type Client struct {
	apiToken string
	tc       domain.TeamCollection
	hc       adapters.HTTPClient
}

// RetrieveLatestStandingsBySeason implements this method on the clients.FootballDataSource interface
func (c *Client) RetrieveLatestStandingsBySeason(ctx context.Context, s domain.Season) (domain.Standings, error) {
	req, err := c.prepareRetrieveStandingsRequest(ctx, s.ClientID.Value(), s.Active.From.Year())
	if err != nil {
		return domain.Standings{}, fmt.Errorf("cannot prepare retrieve standings request: %w", err)
	}

	resp, err := c.hc.Do(req)
	if err != nil {
		return domain.Standings{}, fmt.Errorf("cannot get retrieve standings response: %w", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return domain.Standings{}, fmt.Errorf("cannot read retrieve standings response body: %w", err)
	}

	var stndResp competitionStandingsGetResponse
	if err := json.Unmarshal(body, &stndResp); err != nil {
		return domain.Standings{}, fmt.Errorf("cannot unmarshal retrieve standings response: %w", err)
	}

	ovSt, err := getOverallStandings(stndResp)
	if err != nil {
		return domain.Standings{}, fmt.Errorf("cannot get overall standings: %w", err)
	}

	standings := domain.Standings{
		SeasonID:    s.ID,
		RoundNumber: stndResp.Season.CurrentMatchday,
	}
	for _, tableElem := range ovSt.Table {
		ranking, err := tableElem.toRankingWithMeta(c.tc)
		if err != nil {
			return domain.Standings{}, fmt.Errorf("cannot convert table elem to ranking with meta: %w", err)
		}
		standings.Rankings = append(standings.Rankings, ranking)
	}

	return standings, nil
}

// NewClient generates a new Client
func NewClient(apiToken string, tc domain.TeamCollection, hc adapters.HTTPClient) (*Client, error) {
	if tc == nil {
		return nil, fmt.Errorf("team collection: %w", domain.ErrIsNil)
	}
	if hc == nil {
		return nil, fmt.Errorf("http client: %w", domain.ErrIsNil)
	}
	return &Client{apiToken, tc, hc}, nil
}

func (c *Client) prepareRetrieveStandingsRequest(ctx context.Context, compID string, year int) (*http.Request, error) {
	url := fmt.Sprintf("%s/v2/competitions/%s/standings?season=%d", baseURL, compID, year)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot generate request: url '%s': %w", url, err)
	}

	req.Header.Add("X-Auth-Token", c.apiToken)

	return req, nil
}

// competitionStandingsGetResponse defines the expected payload structure of the request to retrieve the current standings
type competitionStandingsGetResponse struct {
	Season struct {
		CurrentMatchday int `json:"currentMatchday"`
	} `json:"season"`
	Standings []competitionStandings `json:"standings"`
}

// competitionStandings defines the expected payload structure of a standings object on the response
type competitionStandings struct {
	Type  string      `json:"type"`
	Table []tableElem `json:"table"`
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
func (t *tableElem) toRankingWithMeta(tc domain.TeamCollection) (domain.RankingWithMeta, error) {
	r := domain.NewRankingWithMeta()

	team, err := tc.GetByResourceID(domain.TeamIdentifier{TeamID: t.Team.ID})
	if err != nil {
		return domain.RankingWithMeta{}, err
	}

	r.ID = team.ID
	r.Position = t.Position
	r.MetaData[domain.MetaKeyPlayedGames] = t.PlayedGames
	r.MetaData[domain.MetaKeyPoints] = t.Points
	r.MetaData[domain.MetaKeyGoalsFor] = t.GoalsFor
	r.MetaData[domain.MetaKeyGoalsAgainst] = t.GoalsAgainst
	r.MetaData[domain.MetaKeyGoalDifference] = t.GoalDifference

	return r, nil
}

// getOverallStandings returns the standings with a type value of TOTAL from the provided response
func getOverallStandings(resp competitionStandingsGetResponse) (*competitionStandings, error) {
	for _, s := range resp.Standings {
		if s.Type == "TOTAL" {
			return &s, nil
		}
	}
	return nil, errors.New("cannot find standings with type of total")
}
