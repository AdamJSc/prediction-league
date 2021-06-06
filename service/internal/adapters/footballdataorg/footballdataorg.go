package footballdataorg

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"prediction-league/service/internal/domain"
	"strings"
)

const baseURL = "https://api.football-data.org"

// Client defines our football-data.org API client
type Client struct {
	apiToken string
	tc       domain.TeamCollection
}

// RetrieveLatestStandingsBySeason implements this method on the clients.FootballDataSource interface
func (c *Client) RetrieveLatestStandingsBySeason(ctx context.Context, s domain.Season) (domain.Standings, error) {
	// TODO - football data source: abstract http call upstream
	var url = getFullURL(fmt.Sprintf("/v2/competitions/%s/standings?season=%d",
		s.ClientID.Value(),
		s.Active.From.Year()),
	)

	httpResponse, err := getResponse(c, ctx, url)
	if err != nil {
		return domain.Standings{}, err
	}

	body, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		return domain.Standings{}, err
	}

	var standingsResponse competitionStandingsGetResponse
	if err := json.Unmarshal(body, &standingsResponse); err != nil {
		return domain.Standings{}, err
	}

	ovSt, err := getOverallStandings(standingsResponse)
	if err != nil {
		return domain.Standings{}, fmt.Errorf("cannot get overall standings: %w", err)
	}

	standings := domain.Standings{
		SeasonID:    s.ID,
		RoundNumber: standingsResponse.Season.CurrentMatchday,
	}
	for _, tableElem := range ovSt.Table {
		ranking, err := tableElem.toRankingWithMeta(c.tc)
		if err != nil {
			return domain.Standings{}, err
		}
		standings.Rankings = append(standings.Rankings, ranking)
	}

	return standings, nil
}

// NewClient generates a new Client
func NewClient(apiToken string, tc domain.TeamCollection) (*Client, error) {
	if tc == nil {
		return nil, fmt.Errorf("team collection: %w", domain.ErrIsNil)
	}
	return &Client{
		apiToken: apiToken,
		tc:       tc,
	}, nil
}

// getFullURL appends the provided endpoint to the known base URL
func getFullURL(endpoint string) string {
	return fmt.Sprintf("%s/%s", baseURL, strings.Trim(endpoint, "/"))
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
