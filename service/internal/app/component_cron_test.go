package app

import (
	"bytes"
	"errors"
	"net/http"
	"prediction-league/service/internal/adapters/footballdataorg"
	"prediction-league/service/internal/adapters/logger"
	"prediction-league/service/internal/domain"
	"testing"
	"time"
)

func TestNewCronHandler(t *testing.T) {
	ea := &domain.EntryAgent{}
	sa := &domain.StandingsAgent{}
	sepa := &domain.ScoredEntryPredictionAgent{}
	ca := &domain.CommunicationsAgent{}
	mwsa := &domain.MatchWeekSubmissionAgent{}
	mwra := &domain.MatchWeekResultAgent{}
	sc := make(domain.SeasonCollection)
	tc := make(domain.TeamCollection)
	rlms := make(domain.RealmCollection, 0)
	cl := &domain.RealClock{}
	l := &mockLogger{}
	fds := &footballdataorg.Client{}

	tt := []struct {
		name    string
		ea      *domain.EntryAgent
		sa      *domain.StandingsAgent
		sepa    *domain.ScoredEntryPredictionAgent
		ca      *domain.CommunicationsAgent
		mwsa    *domain.MatchWeekSubmissionAgent
		mwra    *domain.MatchWeekResultAgent
		sc      domain.SeasonCollection
		tc      domain.TeamCollection
		rlms    domain.RealmCollection
		cl      domain.Clock
		l       domain.Logger
		fds     domain.FootballDataSource
		wantErr error
	}{
		{"missing entry agent", nil, sa, sepa, ca, mwsa, mwra, sc, tc, rlms, cl, l, fds, domain.ErrIsNil},
		{"missing standings agent", ea, nil, sepa, ca, mwsa, mwra, sc, tc, rlms, cl, l, fds, domain.ErrIsNil},
		{"missing scored entry predictions agent", ea, sa, nil, ca, mwsa, mwra, sc, tc, rlms, cl, l, fds, domain.ErrIsNil},
		{"missing comms agent", ea, sa, sepa, nil, mwsa, mwra, sc, tc, rlms, cl, l, fds, domain.ErrIsNil},
		{"missing match week submission agent", ea, sa, sepa, ca, nil, mwra, sc, tc, rlms, cl, l, fds, domain.ErrIsNil},
		{"missing match week result agent", ea, sa, sepa, ca, mwsa, nil, sc, tc, rlms, cl, l, fds, domain.ErrIsNil},
		{"missing season collection", ea, sa, sepa, ca, mwsa, mwra, nil, tc, rlms, cl, l, fds, domain.ErrIsNil},
		{"missing team collection", ea, sa, sepa, ca, mwsa, mwra, sc, nil, rlms, cl, l, fds, domain.ErrIsNil},
		{"missing realm collection", ea, sa, sepa, ca, mwsa, mwra, sc, tc, nil, cl, l, fds, domain.ErrIsNil},
		{"missing clock", ea, sa, sepa, ca, mwsa, mwra, sc, tc, rlms, nil, l, fds, domain.ErrIsNil},
		{"missing logger", ea, sa, sepa, ca, mwsa, mwra, sc, tc, rlms, cl, nil, fds, domain.ErrIsNil},
		{"missing football client", ea, sa, sepa, ca, mwsa, mwra, sc, tc, rlms, cl, l, nil, domain.ErrIsNil},
		{"no missing dependencies", ea, sa, sepa, ca, mwsa, mwra, sc, tc, rlms, cl, l, fds, nil},
	}

	for idx, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			cnt := &container{
				entryAgent:        tc.ea,
				standingsAgent:    tc.sa,
				sepAgent:          tc.sepa,
				commsAgent:        tc.ca,
				mwSubmissionAgent: tc.mwsa,
				mwResultAgent:     tc.mwra,
				seasons:           tc.sc,
				teams:             tc.tc,
				realms:            tc.rlms,
				clock:             tc.cl,
				logger:            tc.l,
				ftblDataSrc:       tc.fds,
			}
			cr, gotErr := NewCronHandler(cnt)
			if !errors.Is(gotErr, tc.wantErr) {
				t.Fatalf("tc #%d: want error %s (%T), got %s (%T)", idx, tc.wantErr, tc.wantErr, gotErr, gotErr)
			}
			if tc.wantErr == nil && cr == nil {
				t.Fatalf("tc #%d: want non-empty cron handler, got nil", idx)
			}
		})
	}
}

func TestCronFactory_GenerateCron(t *testing.T) {
	t.Run("must fail if no seasons", func(t *testing.T) {
		cf := &CronHandler{}
		wantErrMsg := "need at least one season for active realms"
		if _, gotErr := cf.generateCron(); gotErr == nil || gotErr.Error() != wantErrMsg {
			t.Fatalf("want err %s, got %+v", wantErrMsg, gotErr)
		}
	})

	t.Run("must fail if realm season id does not exist", func(t *testing.T) {
		rc := domain.RealmCollection{
			domain.Realm{Config: domain.RealmConfig{
				SeasonID: "non-existent-season-id",
			}},
		}

		sc := domain.SeasonCollection{"season-id": domain.Season{}}

		cf := &CronHandler{realmCollection: rc, seasonCollection: sc}
		wantErrMsg := "cannot retrieve season by id 'non-existent-season-id': season id non-existent-season-id: not found"
		if _, gotErr := cf.generateCron(); gotErr == nil || gotErr.Error() != wantErrMsg {
			t.Fatalf("want err %s, got %+v", wantErrMsg, gotErr)
		}
	})

	t.Run("must produce 6 cron entries when football data source is provided", func(t *testing.T) {
		rc := domain.RealmCollection{
			domain.Realm{Config: domain.RealmConfig{
				SeasonID: "season-1-id",
			}},
			domain.Realm{Config: domain.RealmConfig{
				SeasonID: "season-2-id",
			}},
		}

		sc := domain.SeasonCollection{
			"season-1-id": domain.Season{},
			"season-2-id": domain.Season{},
		}

		tc := make(domain.TeamCollection)
		hc := &mockHTTPClient{}

		cl := &domain.RealClock{}
		ea := &domain.EntryAgent{}
		ca := &domain.CommunicationsAgent{}
		sa := &domain.StandingsAgent{}
		sepa := &domain.ScoredEntryPredictionAgent{}
		mwsa := &domain.MatchWeekSubmissionAgent{}
		mwra := &domain.MatchWeekResultAgent{}

		buf := &bytes.Buffer{}
		loc, err := time.LoadLocation("Europe/London")
		if err != nil {
			t.Fatal(err)
		}
		dt := time.Date(2018, 5, 26, 14, 0, 0, 0, loc)
		l, err := logger.NewLogger("DEBUG", buf, &mockClock{t: dt})
		if err != nil {
			t.Fatal(err)
		}

		fds, err := footballdataorg.NewClient("12345", tc, hc)
		if err != nil {
			t.Fatal(err)
		}

		handler := &CronHandler{
			realmCollection:            rc,
			seasonCollection:           sc,
			teamCollection:             tc,
			clock:                      cl,
			entryAgent:                 ea,
			commsAgent:                 ca,
			mwSubmissionAgent:          mwsa,
			mwResultAgent:              mwra,
			standingsAgent:             sa,
			scoredEntryPredictionAgent: sepa,
			logger:                     l,
			footballClient:             fds,
		}

		cr, err := handler.generateCron()
		if err != nil {
			t.Fatal(err)
		}

		// 1 job per season
		if len(cr.Entries()) != 2 {
			t.Fatalf("want 2 cron entries, got %d", len(cr.Entries()))
		}
	})
}

type mockClock struct {
	t time.Time
	domain.Clock
}

func (m *mockClock) Now() time.Time {
	return m.t
}

type doFunc func(*http.Request) (*http.Response, error)

type mockHTTPClient struct {
	doFunc
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.doFunc(req)
}
