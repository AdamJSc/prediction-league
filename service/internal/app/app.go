package app

import (
	"prediction-league/service/internal/clients"
	"prediction-league/service/internal/domain"
	"prediction-league/service/internal/messages"
	"prediction-league/service/internal/views"
	"time"

	coresql "github.com/LUSHDigital/core-sql"
	"github.com/gorilla/mux"
)

type DependencyInjector interface {
	ConfigInjector
	MySQLInjector
	EmailClientInjector
	EmailQueueInjector
	RouterInjector
	TemplateInjector
	DebugTimestampInjector
}

type ConfigInjector interface{ Config() domain.Config }
type MySQLInjector interface{ MySQL() coresql.Agent }
type EmailClientInjector interface{ EmailClient() clients.EmailClient }
type EmailQueueInjector interface{ EmailQueue() chan messages.Email }
type RouterInjector interface{ Router() *mux.Router }
type TemplateInjector interface{ Template() *views.Templates }
type DebugTimestampInjector interface{ DebugTimestamp() *time.Time }
