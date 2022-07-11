package domain_test

import (
	"prediction-league/service/internal/domain"
	"testing"
)

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
