package footballdataorg_test

import (
	"errors"
	"net/http"
	"prediction-league/service/internal/adapters"
	"prediction-league/service/internal/adapters/footballdataorg"
	"prediction-league/service/internal/domain"
	"testing"
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
			fdCl, gotErr := footballdataorg.NewClient("12345", tc.tc, tc.hc)
			if !errors.Is(gotErr, tc.wantErr) {
				t.Fatalf("tc #%d: want error %s (%T), got %s (%T)", idx, tc.wantErr, tc.wantErr, gotErr, gotErr)
			}
			if tc.wantErr == nil && fdCl == nil {
				t.Fatalf("tc #%d: want non-empty client, got nil", idx)
			}
		}
	})
}

// TODO - football data source: tests for Client.RetrieveLatestStandingsBySeason

type doFunc func(*http.Request) (*http.Response, error)

type mockHTTPClient struct {
	doFunc
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.doFunc(req)
}
