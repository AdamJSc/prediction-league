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
	t.Run("passing invalid parameters must return expected error", func(t *testing.T) {
		ea := &domain.EntryAgent{}
		sa := &domain.StandingsAgent{}
		sepa := &domain.ScoredEntryPredictionAgent{}
		ca := &domain.CommunicationsAgent{}
		sc := make(domain.SeasonCollection)
		tc := make(domain.TeamCollection)
		rlms := make(domain.RealmCollection)
		cl := &domain.RealClock{}
		l := &mockLogger{}
		fds := &footballdataorg.Client{}

		tt := []struct {
			ea      *domain.EntryAgent
			sa      *domain.StandingsAgent
			sepa    *domain.ScoredEntryPredictionAgent
			ca      *domain.CommunicationsAgent
			sc      domain.SeasonCollection
			tc      domain.TeamCollection
			rlms    domain.RealmCollection
			cl      domain.Clock
			l       domain.Logger
			fds     domain.FootballDataSource
			wantErr error
		}{
			{nil, sa, sepa, ca, sc, tc, rlms, cl, l, fds, domain.ErrIsNil},
			{ea, nil, sepa, ca, sc, tc, rlms, cl, l, fds, domain.ErrIsNil},
			{ea, sa, nil, ca, sc, tc, rlms, cl, l, fds, domain.ErrIsNil},
			{ea, sa, sepa, nil, sc, tc, rlms, cl, l, fds, domain.ErrIsNil},
			{ea, sa, sepa, ca, nil, tc, rlms, cl, l, fds, domain.ErrIsNil},
			{ea, sa, sepa, ca, sc, nil, rlms, cl, l, fds, domain.ErrIsNil},
			{ea, sa, sepa, ca, sc, tc, nil, cl, l, fds, domain.ErrIsNil},
			{ea, sa, sepa, ca, sc, tc, rlms, nil, l, fds, domain.ErrIsNil},
			{ea, sa, sepa, ca, sc, tc, rlms, cl, nil, fds, domain.ErrIsNil},
			{ea, sa, sepa, ca, sc, tc, rlms, cl, l, nil, domain.ErrIsNil},
			{ea, sa, sepa, ca, sc, tc, rlms, cl, l, fds, nil},
		}

		for idx, tc := range tt {
			cnt := &container{
				entryAgent:     tc.ea,
				standingsAgent: tc.sa,
				sepAgent:       tc.sepa,
				commsAgent:     tc.ca,
				seasons:        tc.sc,
				teams:          tc.tc,
				realms:         tc.rlms,
				clock:          tc.cl,
				logger:         tc.l,
				ftblDataSrc:    tc.fds,
			}
			cr, gotErr := NewCronHandler(cnt)
			if !errors.Is(gotErr, tc.wantErr) {
				t.Fatalf("tc #%d: want error %s (%T), got %s (%T)", idx, tc.wantErr, tc.wantErr, gotErr, gotErr)
			}
			if tc.wantErr == nil && cr == nil {
				t.Fatalf("tc #%d: want non-empty cron handler, got nil", idx)
			}
		}
	})
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
			"realm-id": {SeasonID: "non-existent-season-id"},
		}

		sc := domain.SeasonCollection{"season-id": domain.Season{}}

		cf := &CronHandler{rc: rc, sc: sc}
		wantErrMsg := "cannot retrieve season by id 'non-existent-season-id': season id non-existent-season-id: not found"
		if _, gotErr := cf.generateCron(); gotErr == nil || gotErr.Error() != wantErrMsg {
			t.Fatalf("want err %s, got %+v", wantErrMsg, gotErr)
		}
	})

	t.Run("must produce 6 cron entries when football data source is provided", func(t *testing.T) {
		rc := domain.RealmCollection{
			"realm-1-id": {SeasonID: "season-1-id"},
			"realm-2-id": {SeasonID: "season-2-id"},
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

		cf := &CronHandler{rc: rc, sc: sc, tc: tc, cl: cl, ea: ea, ca: ca, sa: sa, sepa: sepa, l: l, fds: fds}
		cr, err := cf.generateCron()
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
