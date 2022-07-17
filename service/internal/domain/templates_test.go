package domain_test

import (
	"prediction-league/service/internal/domain"
	"testing"

	gocmp "github.com/google/go-cmp/cmp"
)

func TestGetMagicLoginURL(t *testing.T) {
	tt := []struct {
		name    string
		realm   *domain.Realm
		token   *domain.Token
		wantURL string
	}{
		{
			name:    "no realm or token",
			wantURL: "/login",
		},
		{
			name: "realm but no token",
			realm: &domain.Realm{Site: domain.RealmSite{
				Origin: "http://localhost",
			}},
			wantURL: "http://localhost/login",
		},
		{
			name:    "token but no realm",
			token:   &domain.Token{ID: "abc123"},
			wantURL: "/login/abc123",
		},
		{
			name: "realm and token",
			realm: &domain.Realm{Site: domain.RealmSite{
				Origin: "http://localhost",
			}},
			token:   &domain.Token{ID: "abc123"},
			wantURL: "http://localhost/login/abc123",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			gotURL := domain.GetMagicLoginURL(tc.realm, tc.token)
			if diff := gocmp.Diff(tc.wantURL, gotURL); diff != "" {
				t.Fatalf("want url %s, got %s, diff: %s", tc.wantURL, gotURL, diff)
			}
		})
	}
}
