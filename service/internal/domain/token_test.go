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
	t.Cleanup(truncate)

	agent, err := domain.NewTokenAgent(tr, &mockClock{t: testDate}, &mockLogger{})
	if err != nil {
		t.Fatal(err)
	}

	tt := []struct {
		name string
		typ  int
	}{
		{
			name: "generate an auth token must succeed",
			typ:  domain.TokenTypeAuth,
		},
		{
			name: "generate an entry registration token must succeed",
			typ:  domain.TokenTypeEntryRegistration,
		},
		{
			name: "generate a magic login token must succeed",
			typ:  domain.TokenTypeMagicLogin,
		},
		{
			name: "generate a prediction token must succeed",
			typ:  domain.TokenTypePrediction,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := testContextDefault(t)
			defer cancel()

			typ := tc.typ
			val := "Hello World"

			token, err := agent.GenerateToken(ctx, typ, val)
			if err != nil {
				t.Fatal(err)
			}

			if len(token.ID) != domain.TokenLength {
				t.Fatalf("want token length %d, got %d", domain.TokenLength, len(token.ID))
			}
			if token.Type != typ {
				t.Fatalf("want token type %d, got %d", typ, token.Type)
			}
			if token.Value != val {
				t.Fatalf("want token value %s, got %s", val, token.Value)
			}
			if !token.IssuedAt.Equal(testDate) {
				t.Fatalf("want token issued at %+v, got %+v", testDate, token.IssuedAt)
			}
			if token.RedeemedAt != nil {
				t.Fatalf("want nil token redeemed at, got %+v", token.RedeemedAt)
			}
			wantExp := testDate.Add(domain.TokenValidityDuration[typ])
			if !token.ExpiresAt.Equal(wantExp) {
				t.Fatalf("want token expires at %+v, got %+v", wantExp, token.ExpiresAt)
			}

			// inserting same token a second time must fail
			if gotErr := tr.Insert(ctx, token); !errors.As(gotErr, &domain.DuplicateDBRecordError{}) {
				t.Fatalf("want DuplicateDBRecordError, got %s (%T)", gotErr, gotErr)
			}
		})
	}

	t.Run("generate a token of a non-existent token type must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		nonExistentTokenType := 123456

		if _, gotErr := agent.GenerateToken(ctx, nonExistentTokenType, "Hello World"); !errors.As(gotErr, &domain.NotFoundError{}) {
			t.Fatalf("want NotFoundError, got %s (%T)", gotErr, gotErr)
		}
	})
}

func TestTokenAgent_GenerateExtendedToken(t *testing.T) {
	t.Cleanup(truncate)

	agent, err := domain.NewTokenAgent(tr, &mockClock{t: testDate}, &mockLogger{})
	if err != nil {
		t.Fatal(err)
	}

	tt := []struct {
		name string
		typ  int
	}{
		{
			name: "generate an auth token must succeed",
			typ:  domain.TokenTypeAuth,
		},
		{
			name: "generate an entry registration token must succeed",
			typ:  domain.TokenTypeEntryRegistration,
		},
		{
			name: "generate a magic login token must succeed",
			typ:  domain.TokenTypeMagicLogin,
		},
		{
			name: "generate a prediction token must succeed",
			typ:  domain.TokenTypePrediction,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := testContextDefault(t)
			ctx = domain.SetBasicAuthSuccessfulOnContext(ctx)
			defer cancel()

			typ := tc.typ
			val := "Hello World"

			token, err := agent.GenerateExtendedToken(ctx, typ, val)
			if err != nil {
				t.Fatal(err)
			}

			if len(token.ID) != domain.TokenLength {
				t.Fatalf("want token length %d, got %d", domain.TokenLength, len(token.ID))
			}
			if token.Type != typ {
				t.Fatalf("want token type %d, got %d", typ, token.Type)
			}
			if token.Value != val {
				t.Fatalf("want token value %s, got %s", val, token.Value)
			}
			if !token.IssuedAt.Equal(testDate) {
				t.Fatalf("want token issued at %+v, got %+v", testDate, token.IssuedAt)
			}
			if token.RedeemedAt != nil {
				t.Fatalf("want nil token redeemed at, got %+v", token.RedeemedAt)
			}
			wantExp := testDate.Add(6 * time.Hour)
			if !token.ExpiresAt.Equal(wantExp) {
				t.Fatalf("want token expires at %+v, got %+v", wantExp, token.ExpiresAt)
			}

			// inserting same token a second time must fail
			if gotErr := tr.Insert(ctx, token); !errors.As(gotErr, &domain.DuplicateDBRecordError{}) {
				t.Fatalf("want DuplicateDBRecordError, got %s (%T)", gotErr, gotErr)
			}
		})
	}

	t.Run("generate a token of a non-existent token type must fail", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		ctx = domain.SetBasicAuthSuccessfulOnContext(ctx)
		defer cancel()

		nonExistentTokenType := 123456

		if _, gotErr := agent.GenerateExtendedToken(ctx, nonExistentTokenType, "Hello World"); !errors.As(gotErr, &domain.NotFoundError{}) {
			t.Fatalf("want NotFoundError, got %s (%T)", gotErr, gotErr)
		}
	})
}

func TestTokenAgent_RetrieveTokenByID(t *testing.T) {
	t.Cleanup(truncate)

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
	t.Cleanup(truncate)

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

	agent, err := domain.NewTokenAgent(tr, &mockClock{t: testDate}, &mockLogger{})
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
		wantTkn.RedeemedAt = &testDate
		gotTkn := getTokenByID(token.ID)

		if diff := gocmp.Diff(wantTkn, gotTkn); diff != "" {
			t.Fatalf("want token %+v, got %+v, diff: %s", wantTkn, gotTkn, diff)
		}
	})

	t.Run("redeem an existing token that has already been redeemed must fail", func(t *testing.T) {
		token := generateTestToken("tkn-2")
		token.RedeemedAt = &testDate
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
	t.Cleanup(truncate)

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
	t.Cleanup(truncate)

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

func TestTokenAgent_DeleteInFlightTokens(t *testing.T) {
	t.Cleanup(truncate)

	agent, err := domain.NewTokenAgent(tr, &mockClock{}, &mockLogger{})
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now().Truncate(time.Second)
	redeemed := testDate

	// tkn1 represents our target token
	tkn1 := &domain.Token{
		ID:        "tkn-1",
		Type:      1,
		Value:     "abcdef",
		IssuedAt:  now,
		ExpiresAt: now,
	}

	// tkn2 represents a token that has been redeemed
	tkn2 := &domain.Token{
		ID:         "tkn-2",
		Type:       1,
		Value:      "abcdef",
		IssuedAt:   now,
		RedeemedAt: &redeemed,
		ExpiresAt:  now,
	}

	// tkn3 represents a token with an alt value
	tkn3 := &domain.Token{
		ID:        "tkn-3",
		Type:      1,
		Value:     "nnnnnnnnnnnn",
		IssuedAt:  now,
		ExpiresAt: now,
	}

	// tkn4 represents a token with an alt type
	tkn4 := &domain.Token{
		ID:        "tkn-4",
		Type:      1234,
		Value:     "abcdef",
		IssuedAt:  now,
		ExpiresAt: now,
	}

	var tokens = []*domain.Token{tkn1, tkn2, tkn3, tkn4}

	for _, token := range tokens {
		if err := tr.Insert(context.Background(), token); err != nil {
			t.Fatal(err)
		}
	}

	t.Run("delete in flight tokens must produce the expected results", func(t *testing.T) {
		ctx, cancel := testContextDefault(t)
		defer cancel()

		gotCnt, err := agent.DeleteInFlightTokens(ctx, 1, "abcdef")
		if err != nil {
			t.Fatal(err)
		}

		wantCnt := int64(1)
		if gotCnt != wantCnt {
			t.Fatalf("want %d deleted tokens, got %d", wantCnt, gotCnt)
		}

		// token 1 must have been deleted
		if err := tr.ExistsByID(ctx, tkn1.ID); !errors.As(err, &domain.MissingDBRecordError{}) {
			t.Fatalf("want token %s to have been deleted, but got err %s (%T)", tkn1.ID, err, err)
		}

		// token 2 must not have been deleted
		if err := tr.ExistsByID(ctx, tkn2.ID); err != nil {
			t.Fatal(err)
		}

		// token 3 must not have been deleted
		if err := tr.ExistsByID(ctx, tkn3.ID); err != nil {
			t.Fatal(err)
		}

		// token 4 must not have been deleted
		if err := tr.ExistsByID(ctx, tkn4.ID); err != nil {
			t.Fatal(err)
		}
	})
}

func TestTokenAgent_IsTokenValid(t *testing.T) {
	tkn := &domain.Token{
		ID:        "tkn-id",
		Type:      domain.TokenTypeAuth,
		Value:     "abcdef",
		ExpiresAt: testDate,
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
			now:     testDate,
			wantRes: true,
		},
		{
			name:    "token with expiry that occurs after now must be valid",
			typ:     domain.TokenTypeAuth,
			val:     "abcdef",
			now:     testDate.Add(-time.Nanosecond),
			wantRes: true,
		},
		{
			name:       "token with alt type must not be valid",
			typ:        domain.TokenTypeEntryRegistration,
			val:        "abcdef",
			now:        testDate,
			wantRes:    false,
			wantLogMsg: "token id 'tkn-id': token type 0 is not 1",
		},
		{
			name:       "token with alt value must not be valid",
			typ:        domain.TokenTypeAuth,
			val:        "ghijkl",
			now:        testDate,
			wantRes:    false,
			wantLogMsg: "token id 'tkn-id': token value 'abcdef' is not 'ghijkl'",
		},
		{
			name:       "token that has expired must not be valid",
			typ:        domain.TokenTypeAuth,
			val:        "abcdef",
			now:        testDate.Add(time.Second),
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
