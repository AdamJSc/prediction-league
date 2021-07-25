package domain_test

import (
	"context"
	"errors"
	"prediction-league/service/internal/domain"
	"testing"
	"time"

	gocmp "github.com/google/go-cmp/cmp"
	"gotest.tools/assert/cmp"
)

func TestNewTokenAgent(t *testing.T) {
	t.Run("passing invalid parameters must return expected error", func(t *testing.T) {
		cl := &mockClock{}
		l := &mockLogger{}

		tt := []struct {
			tr      domain.TokenRepository
			cl      domain.Clock
			l       domain.Logger
			wantErr error
		}{
			{nil, cl, l, domain.ErrIsNil},
			{tr, nil, l, domain.ErrIsNil},
			{tr, cl, nil, domain.ErrIsNil},
			{tr, cl, l, nil},
		}
		for idx, tc := range tt {
			agent, gotErr := domain.NewTokenAgent(tc.tr, tc.cl, tc.l)
			if !errors.Is(gotErr, tc.wantErr) {
				t.Fatalf("tc #%d: want error %s (%T), got %s (%T)", idx, tc.wantErr, tc.wantErr, gotErr, gotErr)
			}
			if tc.wantErr == nil && agent == nil {
				t.Fatalf("tc #%d: want non-empty agent, got nil", idx)
			}
		}
	})
}

func TestTokenAgent_GenerateToken(t *testing.T) {
	defer truncate(t)

	agent, err := domain.NewTokenAgent(tr, &mockClock{t: dt}, &mockLogger{})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("generate an auth token must succeed", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		expectedType := domain.TokenTypeAuth
		expectedValue := "Hello World"

		token, err := agent.GenerateToken(ctx, expectedType, expectedValue)
		if err != nil {
			t.Fatal(err)
		}

		if len(token.ID) != domain.TokenLength {
			expectedGot(t, domain.TokenLength, len(token.ID))
		}
		if token.Type != expectedType {
			expectedGot(t, expectedType, token.Type)
		}
		if token.Value != expectedValue {
			expectedGot(t, expectedValue, token.Value)
		}
		if !token.IssuedAt.Equal(dt) {
			expectedGot(t, dt, token.IssuedAt)
		}
		expectedExpires := dt.Add(domain.TokenValidityDuration[expectedType])
		if !token.ExpiresAt.Equal(expectedExpires) {
			expectedGot(t, expectedExpires, token.ExpiresAt)
		}

		// inserting same token a second time must fail
		err = tr.Insert(ctx, token)
		if !cmp.ErrorType(err, domain.DuplicateDBRecordError{})().Success() {
			expectedTypeOfGot(t, domain.DuplicateDBRecordError{}, err)
		}
	})

	t.Run("generate a short code reset token must succeed", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		expectedType := domain.TokenTypeMagicLogin
		expectedValue := "Hello World"

		token, err := agent.GenerateToken(ctx, expectedType, expectedValue)
		if err != nil {
			t.Fatal(err)
		}

		if len(token.ID) != domain.TokenLength {
			expectedGot(t, domain.TokenLength, len(token.ID))
		}
		if token.Type != expectedType {
			expectedGot(t, expectedType, token.Type)
		}
		if token.Value != expectedValue {
			expectedGot(t, expectedValue, token.Value)
		}
		if !token.IssuedAt.Equal(dt) {
			expectedGot(t, dt, token.IssuedAt)
		}
		expectedExpires := dt.Add(domain.TokenValidityDuration[expectedType])
		if !token.ExpiresAt.Equal(expectedExpires) {
			expectedGot(t, expectedExpires, token.ExpiresAt)
		}

		// inserting same token a second time must fail
		err = tr.Insert(ctx, token)
		if !cmp.ErrorType(err, domain.DuplicateDBRecordError{})().Success() {
			expectedTypeOfGot(t, domain.DuplicateDBRecordError{}, err)
		}
	})

	t.Run("generate a token of a non-existent token type must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		nonExistentTokenType := 123456

		_, err := agent.GenerateToken(ctx, nonExistentTokenType, "Hello World")
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
	})
}

func TestTokenAgent_RetrieveTokenByID(t *testing.T) {
	defer truncate(t)

	agent, err := domain.NewTokenAgent(tr, &mockClock{}, &mockLogger{})
	if err != nil {
		t.Fatal(err)
	}

	token := generateTestToken("token_id")
	if err := tr.Insert(context.Background(), token); err != nil {
		t.Fatal(err)
	}

	t.Run("retrieve an existing token must succeed", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		retrievedToken, err := agent.RetrieveTokenByID(ctx, token.ID)
		if err != nil {
			t.Fatal(err)
		}

		if retrievedToken.ID != token.ID {
			expectedGot(t, token.ID, retrievedToken.ID)
		}
		if retrievedToken.Type != token.Type {
			expectedGot(t, token.Type, retrievedToken.Type)
		}
		if retrievedToken.Value != token.Value {
			expectedGot(t, token.Value, retrievedToken.Value)
		}
		if !retrievedToken.IssuedAt.Equal(token.IssuedAt) {
			expectedGot(t, token.IssuedAt, retrievedToken.IssuedAt)
		}
		if !retrievedToken.ExpiresAt.Equal(token.ExpiresAt) {
			expectedGot(t, token.ExpiresAt, retrievedToken.ExpiresAt)
		}
	})

	t.Run("retrieve a non-existent token must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		_, err := agent.RetrieveTokenByID(ctx, "non_existent_id")
		if !cmp.ErrorType(err, domain.NotFoundError{})().Success() {
			expectedTypeOfGot(t, domain.NotFoundError{}, err)
		}
	})
}

func TestTokenAgent_RedeemToken(t *testing.T) {
	defer truncate(t)

	ctx := context.Background()

	var getTokenByID = func(id string) domain.Token {
		res, err := tr.Select(ctx, map[string]interface{}{
			"id": id,
		}, false)
		if err != nil {
			t.Fatalf("cannot get token by id: %s", err.Error())
		}

		if len(res) != 1 {
			t.Fatalf("got more than 1 token by id: num of results %d", len(res))
		}

		return res[0]
	}

	agent, err := domain.NewTokenAgent(tr, &mockClock{t: dt}, &mockLogger{})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("redeem an existing token that has not been redeemed must succeed", func(t *testing.T) {
		token := generateTestToken("tkn-1")
		if err := tr.Insert(ctx, token); err != nil {
			t.Fatal(err)
		}

		if err := agent.RedeemToken(context.Background(), *token); err != nil {
			t.Fatal(err)
		}

		// token must be redeemed
		wantTkn := *token
		wantTkn.RedeemedAt = &dt
		gotTkn := getTokenByID(token.ID)

		if diff := gocmp.Diff(wantTkn, gotTkn); diff != "" {
			t.Fatalf("want token %+v, got %+v, diff: %s", wantTkn, gotTkn, diff)
		}
	})

	t.Run("redeem an existing token that has already been redeemed must fail", func(t *testing.T) {
		token := generateTestToken("tkn-2")
		token.RedeemedAt = &dt
		if err := tr.Insert(ctx, token); err != nil {
			t.Fatal(err)
		}

		if gotErr := agent.RedeemToken(context.Background(), *token); !errors.As(gotErr, &domain.ConflictError{}) {
			t.Fatalf("want conflict error, got %s (%T)", gotErr, gotErr)
		}
	})

	t.Run("redeem a non-existent token must fail", func(t *testing.T) {
		token := generateTestToken("non-existent-id")
		if gotErr := agent.RedeemToken(context.Background(), *token); !errors.As(gotErr, &domain.NotFoundError{}) {
			t.Fatalf("want not found error, got %s (%T)", gotErr, gotErr)
		}
	})
}

func TestTokenAgent_DeleteToken(t *testing.T) {
	defer truncate(t)

	agent, err := domain.NewTokenAgent(tr, &mockClock{}, &mockLogger{})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("delete an existing token must succeed", func(t *testing.T) {
		tkn := generateTestToken("tkn-1")
		if err := tr.Insert(context.Background(), tkn); err != nil {
			t.Fatal(err)
		}

		ctx, cancel := testContextDefault(t)
		defer cancel()

		if err := agent.DeleteToken(ctx, *tkn); err != nil {
			t.Fatal(err)
		}

		// token must have been deleted
		if gotErr := tr.ExistsByID(ctx, tkn.ID); !errors.As(gotErr, &domain.MissingDBRecordError{}) {
			t.Fatalf("want MissingDBRecordError, got %s (%T)", gotErr, gotErr)
		}
	})

	t.Run("delete a non-existent token must succeed", func(t *testing.T) {
		tkn := generateTestToken("tkn-2")
		if err := tr.Insert(context.Background(), tkn); err != nil {
			t.Fatal(err)
		}

		ctx, cancel := testContextDefault(t)
		defer cancel()

		altTkn := domain.Token{ID: "non-existent-token"}

		if gotErr := agent.DeleteToken(ctx, altTkn); !errors.As(gotErr, &domain.NotFoundError{}) {
			t.Fatalf("want NotFoundError, got %s (%T)", gotErr, gotErr)
		}
	})
}

func TestTokenAgent_DeleteTokensExpiredAfter(t *testing.T) {
	defer truncate(t)

	agent, err := domain.NewTokenAgent(tr, &mockClock{}, &mockLogger{})
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now().Truncate(time.Second)

	// token 1 represents a token that expires 1 second in the past
	token1 := generateTestToken("tkn-1")
	token1.ExpiresAt = now.Add(-time.Second)

	// token 2 represents a token that expires now
	token2 := generateTestToken("tkn-2")
	token2.ExpiresAt = now

	// token 3 represents a token that expires 1 second in the future
	token3 := generateTestToken("tkn-3")
	token3.ExpiresAt = now.Add(time.Second)

	var tokens = []*domain.Token{
		token1,
		token2,
		token3,
	}

	for _, token := range tokens {
		if err := tr.Insert(context.Background(), token); err != nil {
			t.Fatal(err)
		}
	}

	t.Run("delete tokens expired after valid timestamp must succeed", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		gotCnt, err := agent.DeleteTokensExpiredAfter(ctx, now)
		if err != nil {
			t.Fatal(err)
		}

		wantCnt := int64(2)
		if gotCnt != wantCnt {
			t.Fatalf("want %d deleted tokens, got %d", wantCnt, gotCnt)
		}

		// token 1 must have been deleted
		if err := tr.ExistsByID(ctx, token1.ID); !errors.As(err, &domain.MissingDBRecordError{}) {
			t.Fatalf("want token %s to have been deleted, but got err %s (%T)", token1.ID, err, err)
		}

		// token 2 must have been deleted
		if err := tr.ExistsByID(ctx, token2.ID); !errors.As(err, &domain.MissingDBRecordError{}) {
			t.Fatalf("want token %s to have been deleted, but got err %s (%T)", token2.ID, err, err)
		}

		// token 3 must not have been deleted
		if err := tr.ExistsByID(ctx, token3.ID); err != nil {
			t.Fatal(err)
		}
	})
}

func TestTokenAgent_IsTokenValid(t *testing.T) {
	tkn := &domain.Token{
		ID:        "tkn-id",
		Type:      domain.TokenTypeAuth,
		Value:     "abcdef",
		ExpiresAt: dt,
	}

	tt := []struct {
		name       string
		typ        int
		val        string
		now        time.Time
		wantRes    bool
		wantLogMsg string
	}{
		{
			name:    "token with expiry that matches now must be valid",
			typ:     domain.TokenTypeAuth,
			val:     "abcdef",
			now:     dt,
			wantRes: true,
		},
		{
			name:    "token with expiry that occurs after now must be valid",
			typ:     domain.TokenTypeAuth,
			val:     "abcdef",
			now:     dt.Add(-time.Nanosecond),
			wantRes: true,
		},
		{
			name:       "token with alt type must not be valid",
			typ:        domain.TokenTypeEntryRegistration,
			val:        "abcdef",
			now:        dt,
			wantRes:    false,
			wantLogMsg: "token id 'tkn-id': token type 0 is not 1",
		},
		{
			name:       "token with alt value must not be valid",
			typ:        domain.TokenTypeAuth,
			val:        "ghijkl",
			now:        dt,
			wantRes:    false,
			wantLogMsg: "token id 'tkn-id': token value 'abcdef' is not 'ghijkl'",
		},
		{
			name:       "token that has expired must not be valid",
			typ:        domain.TokenTypeAuth,
			val:        "abcdef",
			now:        dt.Add(time.Second),
			wantRes:    false,
			wantLogMsg: "token id 'tkn-id': expired",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			l := newMockLogger()
			cl := &mockClock{t: tc.now}

			ta, err := domain.NewTokenAgent(tr, cl, l)
			if err != nil {
				t.Fatal(err)
			}

			if gotRes := ta.IsTokenValid(tkn, tc.typ, tc.val); gotRes != tc.wantRes {
				t.Fatalf("want result %t, got %t", tc.wantRes, gotRes)
			}

			gotLogMsg := l.buf.String()
			if diff := gocmp.Diff(tc.wantLogMsg, gotLogMsg); diff != "" {
				t.Fatalf("want log msg '%s', got '%s', diff: %s", tc.wantLogMsg, gotLogMsg, diff)
			}
		})
	}
}

func generateTestToken(id string) *domain.Token {
	// arbitrary timestamp that isn't the current moment
	ts := time.Now().Truncate(time.Second).Add(-24 * time.Hour)

	return &domain.Token{
		ID:        id,
		Type:      123,
		Value:     "Hello World",
		IssuedAt:  ts,
		ExpiresAt: ts.Add(time.Minute),
	}
}
