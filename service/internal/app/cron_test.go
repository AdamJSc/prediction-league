package app

import (
	"bytes"
	"errors"
	"prediction-league/service/internal/adapters/footballdataorg"
	"prediction-league/service/internal/adapters/logger"
	"prediction-league/service/internal/domain"
	"testing"
	"time"
)

func TestNewCronFactory(t *testing.T) {
	t.Run("passing nil must return expected error", func(t *testing.T) {
		ea := &domain.EntryAgent{}
		sa := &domain.StandingsAgent{}
		sepa := &domain.ScoredEntryPredictionAgent{}
		ca := &domain.CommunicationsAgent{}
		sc := make(domain.SeasonCollection)
		tc := make(domain.TeamCollection)
		rlms := make(map[string]domain.Realm)
		cl := &domain.RealClock{}
		l := &logger.Logger{}
		fds, err := footballdataorg.NewClient("", tc)
		if err != nil {
			t.Fatal(err)
		}

		tt := []struct {
			ea   *domain.EntryAgent
			sa   *domain.StandingsAgent
			sepa *domain.ScoredEntryPredictionAgent
			ca   *domain.CommunicationsAgent
			sc   domain.SeasonCollection
			tc   domain.TeamCollection
			rlms map[string]domain.Realm
			cl   domain.Clock
			l    domain.Logger
		}{
			{nil, sa, sepa, ca, sc, tc, rlms, cl, l},
			{ea, nil, sepa, ca, sc, tc, rlms, cl, l},
			{ea, sa, nil, ca, sc, tc, rlms, cl, l},
			{ea, sa, sepa, nil, sc, tc, rlms, cl, l},
			{ea, sa, sepa, ca, nil, tc, rlms, cl, l},
			{ea, sa, sepa, ca, sc, nil, rlms, cl, l},
			{ea, sa, sepa, ca, sc, tc, nil, cl, l},
			{ea, sa, sepa, ca, sc, tc, rlms, nil, l},
			{ea, sa, sepa, ca, sc, tc, rlms, cl, nil},
		}

		for _, tc := range tt {
			_, gotErr := NewCronFactory(tc.ea, tc.sa, tc.sepa, tc.ca, tc.sc, tc.tc, tc.rlms, tc.cl, tc.l, fds)
			if !errors.Is(gotErr, domain.ErrIsNil) {
				t.Fatalf("want ErrIsNil, got %s (%T)", gotErr, gotErr)
			}
		}

		// allow nil fds
		if _, err := NewCronFactory(ea, sa, sepa, ca, sc, tc, rlms, cl, l, nil); err != nil {
			t.Fatal(err)
		}
	})
}

func TestCronFactory_Make(t *testing.T) {
	t.Run("must fail if no seasons", func(t *testing.T) {
		cf := &CronFactory{}
		wantErrMsg := "need at least one season for active realms"
		if _, gotErr := cf.Make(); gotErr == nil || gotErr.Error() != wantErrMsg {
			t.Fatalf("want err %s, got %+v", wantErrMsg, gotErr)
		}
	})

	t.Run("must fail if realm season id does not exist", func(t *testing.T) {
		rlms := map[string]domain.Realm{
			"realm-id": {SeasonID: "non-existent-season-id"},
		}

		sc := domain.SeasonCollection{"season-id": domain.Season{}}

		cf := &CronFactory{rlms: rlms, sc: sc}
		wantErrMsg := "cannot retrieve season by id 'non-existent-season-id': season id non-existent-season-id: not found"
		if _, gotErr := cf.Make(); gotErr == nil || gotErr.Error() != wantErrMsg {
			t.Fatalf("want err %s, got %+v", wantErrMsg, gotErr)
		}
	})

	t.Run("must fail if job cannot be generated", func(t *testing.T) {
		rlms := map[string]domain.Realm{
			"realm-id": {SeasonID: "season-id"},
		}

		sc := domain.SeasonCollection{"season-id": domain.Season{}}

		cf := &CronFactory{rlms: rlms, sc: sc}
		wantErrMsg := "cannot generate job configs: cannot generate new prediction window open job: cannot instantiate prediction window open worker: clock: is nil"
		if _, gotErr := cf.Make(); gotErr == nil || gotErr.Error() != wantErrMsg {
			t.Fatalf("want err %s, got %+v", wantErrMsg, gotErr)
		}
	})

	t.Run("must produce 4 cron entries when no football data source is provided", func(t *testing.T) {
		rlms := map[string]domain.Realm{
			"realm-1-id": {SeasonID: "season-1-id"},
			"realm-2-id": {SeasonID: "season-2-id"},
		}

		sc := domain.SeasonCollection{
			"season-1-id": domain.Season{},
			"season-2-id": domain.Season{},
		}

		cl := &domain.RealClock{}
		ea := &domain.EntryAgent{}
		ca := &domain.CommunicationsAgent{}

		buf := &bytes.Buffer{}
		loc, err := time.LoadLocation("Europe/London")
		if err != nil {
			t.Fatal(err)
		}
		dt := time.Date(2018, 5, 26, 14, 0, 0, 0, loc)
		l, err := logger.NewLogger(buf, &mockClock{t: dt})
		if err != nil {
			t.Fatal(err)
		}

		cf := &CronFactory{rlms: rlms, sc: sc, cl: cl, ea: ea, ca: ca, l: l}
		cr, err := cf.Make()
		if err != nil {
			t.Fatal(err)
		}

		// 2 jobs per season
		if len(cr.Entries()) != 4 {
			t.Fatalf("want 4 cron entries, got %d", len(cr.Entries()))
		}
	})

	t.Run("must produce 6 cron entries when football data source is provided", func(t *testing.T) {
		rlms := map[string]domain.Realm{
			"realm-1-id": {SeasonID: "season-1-id"},
			"realm-2-id": {SeasonID: "season-2-id"},
		}

		sc := domain.SeasonCollection{
			"season-1-id": domain.Season{},
			"season-2-id": domain.Season{},
		}

		tc := make(domain.TeamCollection)

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
		l, err := logger.NewLogger(buf, &mockClock{t: dt})
		if err != nil {
			t.Fatal(err)
		}

		fds, err := footballdataorg.NewClient("12345", tc)
		if err != nil {
			t.Fatal(err)
		}

		cf := &CronFactory{rlms: rlms, sc: sc, tc: tc, cl: cl, ea: ea, ca: ca, sa: sa, sepa: sepa, l: l, fds: fds}
		cr, err := cf.Make()
		if err != nil {
			t.Fatal(err)
		}

		// 3 jobs per season
		if len(cr.Entries()) != 6 {
			t.Fatalf("want 6 cron entries, got %d", len(cr.Entries()))
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
