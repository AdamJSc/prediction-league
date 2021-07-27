package domain_test

import (
	"prediction-league/service/internal/domain"
	"testing"

	gocmp "github.com/google/go-cmp/cmp"
)

func TestGetHomeURL(t *testing.T) {
	tt := []struct {
		name    string
		realm   *domain.Realm
		wantURL string
	}{
		{
			name:    "no realm",
			wantURL: "/",
		},
		{
			name:    "with realm",
			realm:   &domain.Realm{Origin: "http://localhost"},
			wantURL: "http://localhost",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			gotURL := domain.GetHomeURL(tc.realm)
			if diff := gocmp.Diff(tc.wantURL, gotURL); diff != "" {
				t.Fatalf("want url %s, got %s, diff: %s", tc.wantURL, gotURL, diff)
			}
		})
	}
}

func TestGetLeaderBoardURL(t *testing.T) {
	tt := []struct {
		name    string
		realm   *domain.Realm
		wantURL string
	}{
		{
			name:    "no realm",
			wantURL: "/leaderboard",
		},
		{
			name:    "with realm",
			realm:   &domain.Realm{Origin: "http://localhost"},
			wantURL: "http://localhost/leaderboard",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			gotURL := domain.GetLeaderBoardURL(tc.realm)
			if diff := gocmp.Diff(tc.wantURL, gotURL); diff != "" {
				t.Fatalf("want url %s, got %s, diff: %s", tc.wantURL, gotURL, diff)
			}
		})
	}
}

func TestGetJoinURL(t *testing.T) {
	tt := []struct {
		name    string
		realm   *domain.Realm
		wantURL string
	}{
		{
			name:    "no realm",
			wantURL: "/join",
		},
		{
			name:    "with realm",
			realm:   &domain.Realm{Origin: "http://localhost"},
			wantURL: "http://localhost/join",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			gotURL := domain.GetJoinURL(tc.realm)
			if diff := gocmp.Diff(tc.wantURL, gotURL); diff != "" {
				t.Fatalf("want url %s, got %s, diff: %s", tc.wantURL, gotURL, diff)
			}
		})
	}
}

func TestGetFAQURL(t *testing.T) {
	tt := []struct {
		name    string
		realm   *domain.Realm
		wantURL string
	}{
		{
			name:    "no realm",
			wantURL: "/faq",
		},
		{
			name:    "with realm",
			realm:   &domain.Realm{Origin: "http://localhost"},
			wantURL: "http://localhost/faq",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			gotURL := domain.GetFAQURL(tc.realm)
			if diff := gocmp.Diff(tc.wantURL, gotURL); diff != "" {
				t.Fatalf("want url %s, got %s, diff: %s", tc.wantURL, gotURL, diff)
			}
		})
	}
}

func TestGetLoginURL(t *testing.T) {
	tt := []struct {
		name    string
		realm   *domain.Realm
		wantURL string
	}{
		{
			name:    "no realm",
			wantURL: "/login",
		},
		{
			name:    "with realm",
			realm:   &domain.Realm{Origin: "http://localhost"},
			wantURL: "http://localhost/login",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			gotURL := domain.GetLoginURL(tc.realm)
			if diff := gocmp.Diff(tc.wantURL, gotURL); diff != "" {
				t.Fatalf("want url %s, got %s, diff: %s", tc.wantURL, gotURL, diff)
			}
		})
	}
}

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
			name:    "realm but no token",
			realm:   &domain.Realm{Origin: "http://localhost"},
			wantURL: "http://localhost/login",
		},
		{
			name:    "token but no realm",
			token:   &domain.Token{ID: "abc123"},
			wantURL: "/login/abc123",
		},
		{
			name:    "realm and token",
			realm:   &domain.Realm{Origin: "http://localhost"},
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

func TestGetPredictionURL(t *testing.T) {
	tt := []struct {
		name    string
		realm   *domain.Realm
		wantURL string
	}{
		{
			name:    "no realm",
			wantURL: "/prediction",
		},
		{
			name:    "with realm",
			realm:   &domain.Realm{Origin: "http://localhost"},
			wantURL: "http://localhost/prediction",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			gotURL := domain.GetPredictionURL(tc.realm)
			if diff := gocmp.Diff(tc.wantURL, gotURL); diff != "" {
				t.Fatalf("want url %s, got %s, diff: %s", tc.wantURL, gotURL, diff)
			}
		})
	}
}
