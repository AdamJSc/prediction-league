package footballdataorg_test

import (
	"errors"
	"prediction-league/service/internal/adapters/footballdataorg"
	"prediction-league/service/internal/domain"
	"testing"
)

func TestNewClient(t *testing.T) {
	t.Run("passing invalid parameters must return expected error", func(t *testing.T) {
		tc := make(domain.TeamCollection)

		tt := []struct {
			tc      domain.TeamCollection
			wantErr error
		}{
			{nil, domain.ErrIsNil},
			{tc, nil},
		}
		for idx, tc := range tt {
			fdCl, gotErr := footballdataorg.NewClient("12345", tc.tc)
			if !errors.Is(gotErr, tc.wantErr) {
				t.Fatalf("tc #%d: want error %s (%T), got %s (%T)", idx, tc.wantErr, tc.wantErr, gotErr, gotErr)
			}
			if tc.wantErr == nil && fdCl == nil {
				t.Fatalf("tc #%d: want non-empty client, got nil", idx)
			}
		}
	})
}

// TODO - tests for Client.RetrieveLatestStandingsBySeason
