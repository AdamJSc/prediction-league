package footballdataorg

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"prediction-league/service/internal/adapters"
	"prediction-league/service/internal/domain"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestNewClient(t *testing.T) {
	t.Run("passing invalid parameters must return expected error", func(t *testing.T) {
		tc := make(domain.TeamCollection)
		hc := &mockHTTPClient{}

		tt := []struct {
			tc      domain.TeamCollection
			hc      adapters.HTTPClient
			wantErr error
		}{
			{nil, hc, domain.ErrIsNil},
			{tc, nil, domain.ErrIsNil},
			{tc, hc, nil},
		}
		for idx, tc := range tt {
			fdCl, gotErr := NewClient("12345", tc.tc, tc.hc)
			if !errors.Is(gotErr, tc.wantErr) {
				t.Fatalf("tc #%d: want error %s (%T), got %s (%T)", idx, tc.wantErr, tc.wantErr, gotErr, gotErr)
			}
			if tc.wantErr == nil && fdCl == nil {
				t.Fatalf("tc #%d: want non-empty client, got nil", idx)
			}
		}
	})
}

func TestClient_RetrieveLatestStandingsBySeason(t *testing.T) {
	loc, err := time.LoadLocation("Europe/London")
	if err != nil {
		t.Fatal(err)
	}

	dt := time.Date(2018, 5, 26, 14, 0, 0, 0, loc)
	apiToken := "my-token"
	s := domain.Season{
		ID:       "season-id",
		ClientID: domain.SeasonIdentifier{SeasonID: "season-client-id"},
		Live:     domain.TimeFrame{From: dt},
	}

	t.Run("happy path must produce the expected standings", func(t *testing.T) {
		hc := &mockHTTPClient{func(req *http.Request) (*http.Response, error) {
			wantURL := "https://api.football-data.org/v2/competitions/season-client-id/standings?season=2018"
			gotURL := req.URL.String()
			diff := cmp.Diff(wantURL, gotURL)
			if diff != "" {
				t.Fatalf("want request url '%s', got '%s', diff: %s", wantURL, gotURL, diff)
			}

			wantHdr := apiToken
			gotHdr := req.Header.Get("X-Auth-Token")
			diff = cmp.Diff(wantHdr, gotHdr)
			if diff != "" {
				t.Fatalf("want token heander '%s', got '%s', diff: %s", wantHdr, gotHdr, diff)
			}

			body := `{
				"season": {
					"currentMatchday": 123
				},
				"standings": [
					{
						"type": "TOTAL",
						"table": [
							{
								"position": 1,
								"team": {
									"id": 1111,
									"name": "AFC Bournemouth"
								},
								"playedGames": 1,
								"goalsFor": 2,
								"goalsAgainst": 3,
								"goalDifference": 4,
								"points": 5
							},
							{
								"position": 2,
								"team": {
									"id": 2222,
									"name": "Poole Town"
								},
								"playedGames": 6,
								"goalsFor": 7,
								"goalsAgainst": 8,
								"goalDifference": 9,
								"points": 10
							},
							{
								"position": 3,
								"team": {
									"id": 3333,
									"name": "Redbridge Rovers"
								},
								"playedGames": 11,
								"goalsFor": 12,
								"goalsAgainst": 13,
								"goalDifference": 14,
								"points": 15
							}
						]
					}
				]
			}`

			resp := &http.Response{Body: ioutil.NopCloser(bytes.NewReader([]byte(body)))}

			return resp, nil
		}}

		tc := domain.TeamCollection{
			"aaa": domain.Team{
				ID:       "AFCB",
				ClientID: domain.TeamIdentifier{TeamID: 1111},
			},
			"bbb": domain.Team{
				ID:       "PTFC",
				ClientID: domain.TeamIdentifier{TeamID: 2222},
			},
			"ccc": domain.Team{
				ID:       "RRFC",
				ClientID: domain.TeamIdentifier{TeamID: 3333},
			},
		}

		wantSt := domain.Standings{
			SeasonID:    "season-id",
			RoundNumber: 123,
			Rankings: []domain.RankingWithMeta{
				{
					Ranking: domain.Ranking{
						ID:       "AFCB",
						Position: 1,
					},
					MetaData: map[string]int{
						domain.MetaKeyPlayedGames: 1,
					},
				},
				{
					Ranking: domain.Ranking{
						ID:       "PTFC",
						Position: 2,
					},
					MetaData: map[string]int{
						domain.MetaKeyPlayedGames: 6,
					},
				},
				{
					Ranking: domain.Ranking{
						ID:       "RRFC",
						Position: 3,
					},
					MetaData: map[string]int{
						domain.MetaKeyPlayedGames: 11,
					},
				},
			},
		}

		cl := Client{apiToken, tc, hc}
		gotSt, err := cl.RetrieveLatestStandingsBySeason(context.Background(), s)
		if err != nil {
			t.Fatal(err)
		}
		diff := cmp.Diff(wantSt, gotSt)
		if diff != "" {
			t.Fatalf("want standings %+v, got %+v, diff: %s", wantSt, gotSt, diff)
		}
	})

	t.Run("failed call to http client must return expected error", func(t *testing.T) {
		hc := &mockHTTPClient{doFunc: func(req *http.Request) (*http.Response, error) {
			return nil, errors.New("sad times :'(")
		}}

		cl := &Client{hc: hc}

		wantErrMsg := "cannot get retrieve standings response: sad times :'("
		_, gotErr := cl.RetrieveLatestStandingsBySeason(context.Background(), s)
		if gotErr == nil || gotErr.Error() != wantErrMsg {
			t.Fatalf("want error msg %s, got %+v (%T)", wantErrMsg, gotErr, gotErr)
		}
	})

	t.Run("failure to read response body must return expected error", func(t *testing.T) {
		hc := &mockHTTPClient{doFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{Body: &mockReader{readFunc: func(p []byte) (int, error) {
				return 0, errors.New("sad times :'(")
			}}}, nil
		}}

		cl := &Client{hc: hc}

		wantErrMsg := "cannot read retrieve standings response body: sad times :'("
		_, gotErr := cl.RetrieveLatestStandingsBySeason(context.Background(), s)
		if gotErr == nil || gotErr.Error() != wantErrMsg {
			t.Fatalf("want error msg %s, got %+v (%T)", wantErrMsg, gotErr, gotErr)
		}
	})

	t.Run("failure to unmarshal response body must return expected error", func(t *testing.T) {
		hc := &mockHTTPClient{doFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{Body: ioutil.NopCloser(bytes.NewBuffer([]byte(
				`{"not_valid_json"}"`,
			)))}, nil
		}}

		cl := &Client{hc: hc}

		wantErrMsg := "cannot unmarshal retrieve standings response: invalid character '}' after object key"
		_, gotErr := cl.RetrieveLatestStandingsBySeason(context.Background(), s)
		if gotErr == nil || gotErr.Error() != wantErrMsg {
			t.Fatalf("want error msg %s, got %+v (%T)", wantErrMsg, gotErr, gotErr)
		}
	})

	t.Run("failure to obtain overall standings must return expected error", func(t *testing.T) {
		hc := &mockHTTPClient{doFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{Body: ioutil.NopCloser(bytes.NewBuffer([]byte(
				`{
						"season": {
							"currentMatchday": 123
						},
						"standings": [
							{
								"type": "NOT_TOTAL",
								"table": []
							}
						]
					}`,
			)))}, nil
		}}

		cl := &Client{hc: hc}

		wantErrMsg := "cannot get overall standings: cannot find standings with type of total"
		_, gotErr := cl.RetrieveLatestStandingsBySeason(context.Background(), s)
		if gotErr == nil || gotErr.Error() != wantErrMsg {
			t.Fatalf("want error msg %s, got %+v (%T)", wantErrMsg, gotErr, gotErr)
		}
	})

	t.Run("failure to convert table element must return expected error", func(t *testing.T) {
		hc := &mockHTTPClient{doFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{Body: ioutil.NopCloser(bytes.NewBuffer([]byte(
				`{
						"season": {
							"currentMatchday": 123
						},
						"standings": [
							{
								"type": "TOTAL",
								"table": [
									{
										"team": {
											"id": 99999
										}
									}
								]
							}
						]
					}`,
			)))}, nil
		}}

		cl := &Client{hc: hc}

		wantErrMsg := "cannot convert table elem to ranking with meta: team client resource id 99999: not found"
		_, gotErr := cl.RetrieveLatestStandingsBySeason(context.Background(), s)
		if gotErr == nil || gotErr.Error() != wantErrMsg {
			t.Fatalf("want error msg %s, got %+v (%T)", wantErrMsg, gotErr, gotErr)
		}
	})
}

type doFunc func(*http.Request) (*http.Response, error)

type mockHTTPClient struct {
	doFunc
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.doFunc(req)
}

type readFunc func(p []byte) (int, error)
type closeFunc func() error

type mockReader struct {
	readFunc
	closeFunc
}

func (m *mockReader) Read(p []byte) (int, error) {
	return m.readFunc(p)
}

func (m *mockReader) Close() error {
	return m.closeFunc()
}
