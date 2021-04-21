package app

import (
	"prediction-league/service/internal/domain"
	"time"

	"github.com/gorilla/mux"
)

type DependencyInjector interface {
	ConfigInjector
	EmailClientInjector
	EmailQueueInjector
	RouterInjector
	TemplateInjector
	DebugTimestampInjector
	StandingsRepoInjector
	EntryRepoInjector
	EntryPredictionRepoInjector
	ScoredEntryPredictionRepoInjector
	TokenRepoInjector
}

type ConfigInjector interface{ Config() domain.Config }
type EmailClientInjector interface{ EmailClient() domain.EmailClient }
type EmailQueueInjector interface{ EmailQueue() chan domain.Email }
type RouterInjector interface{ Router() *mux.Router }
type TemplateInjector interface{ Template() *domain.Templates }
type DebugTimestampInjector interface{ DebugTimestamp() *time.Time }
type StandingsRepoInjector interface {
	StandingsRepo() domain.StandingsRepository
}
type EntryRepoInjector interface {
	EntryRepo() domain.EntryRepository
}
type EntryPredictionRepoInjector interface {
	EntryPredictionRepo() domain.EntryPredictionRepository
}
type ScoredEntryPredictionRepoInjector interface {
	ScoredEntryPredictionRepo() domain.ScoredEntryPredictionRepository
}
type TokenRepoInjector interface {
	TokenRepo() domain.TokenRepository
}
