package mailgun_test

import (
	"errors"
	"prediction-league/service/internal/adapters/mailgun"
	"prediction-league/service/internal/domain"
	"testing"
)

func TestNewClient(t *testing.T) {
	if _, gotErr := mailgun.NewClient(""); !errors.Is(gotErr, domain.ErrIsEmpty) {
		t.Fatalf("want ErrIsEmpty, got %s (%T)", gotErr, gotErr)
	}
}
