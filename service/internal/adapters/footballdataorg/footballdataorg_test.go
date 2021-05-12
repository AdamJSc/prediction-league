package footballdataorg_test

import (
	"errors"
	"prediction-league/service/internal/adapters/footballdataorg"
	"prediction-league/service/internal/domain"
	"testing"
)

func TestNewClient(t *testing.T) {
	t.Run("passing nil must return expected error", func(t *testing.T) {
		_, gotErr := footballdataorg.NewClient("12345", nil)
		if !errors.Is(gotErr, domain.ErrIsNil) {
			t.Fatalf("want ErrIsNil, got %s (%T)", gotErr, gotErr)
		}
	})
}

// TODO - tests for Client.RetrieveLatestStandingsBySeason
