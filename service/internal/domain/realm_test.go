package domain_test

import (
	"prediction-league/service/internal/domain"
	"testing"
)

func TestRealm_GetFullHomeURL(t *testing.T) {
	realm := domain.Realm{
		Site: domain.RealmSite{
			Origin: "http://localhost",
			Paths: domain.RealmSitePaths{
				Home: "/home",
			},
		},
	}

	want := "http://localhost/home"
	got := realm.GetFullHomeURL()

	cmpDiff(t, "full home url", want, got)
}

func TestRealm_GetFullLeaderboardURL(t *testing.T) {
	realm := domain.Realm{
		Site: domain.RealmSite{
			Origin: "http://localhost",
			Paths: domain.RealmSitePaths{
				Leaderboard: "/leaderboard",
			},
		},
	}

	want := "http://localhost/leaderboard"
	got := realm.GetFullLeaderboardURL()

	cmpDiff(t, "full leaderboard url", want, got)
}

func TestRealm_GetFullMyTableURL(t *testing.T) {
	realm := domain.Realm{
		Site: domain.RealmSite{
			Origin: "http://localhost",
			Paths: domain.RealmSitePaths{
				MyTable: "/prediction",
			},
		},
	}

	want := "http://localhost/prediction"
	got := realm.GetFullMyTableURL()

	cmpDiff(t, "full leaderboard url", want, got)
}

func TestRealm_GetMagicLoginURL(t *testing.T) {
	realm := domain.Realm{Site: domain.RealmSite{
		Origin: "http://localhost",
		Paths: domain.RealmSitePaths{
			Login: "/login",
		},
	}}

	tt := []struct {
		name    string
		token   *domain.Token
		wantURL string
	}{
		{
			name:    "no token",
			wantURL: "http://localhost/login",
		},
		{
			name:    "valid token",
			token:   &domain.Token{ID: "abc123"},
			wantURL: "http://localhost/login/abc123",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			gotURL := realm.GetMagicLoginURL(tc.token)
			cmpDiff(t, "magic login url", tc.wantURL, gotURL)
		})
	}
}

func TestRealmCollection_GetByName(t *testing.T) {
	collection := domain.RealmCollection{
		domain.Realm{Config: domain.RealmConfig{Name: "realm_1"}},
		domain.Realm{Config: domain.RealmConfig{Name: "realm_2"}},
		domain.Realm{Config: domain.RealmConfig{Name: "realm_3"}},
	}

	t.Run("retrieving an existing realm by id must succeed", func(t *testing.T) {
		id := "realm_2"
		r, err := collection.GetByName(id)
		if err != nil {
			t.Fatal(err)
		}

		if r.Config.Name != id {
			expectedGot(t, id, r.Config.Name)
		}
	})

	t.Run("retrieving a non-existing realm by id must fail", func(t *testing.T) {
		id := "not_existent_realm_id"
		if _, err := collection.GetByName(id); err == nil {
			expectedNonEmpty(t, "realm collection getbyid error")
		}
	})
}
