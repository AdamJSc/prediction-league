package app

import (
	coresql "github.com/LUSHDigital/core-sql"
	"github.com/gorilla/mux"
	"prediction-league/service/internal/clients"
	"prediction-league/service/internal/domain"
	"prediction-league/service/internal/views"
	"time"
)

type DependencyInjector interface {
	ConfigInjector
	MySQLInjector
	EmailClientInjector
	RouterInjector
	TemplateInjector
	DebugTimestampInjector
}

type ConfigInjector interface{ Config() domain.Config }
type MySQLInjector interface{ MySQL() coresql.Agent }
type EmailClientInjector interface{ EmailClient() clients.EmailClient }
type RouterInjector interface{ Router() *mux.Router }
type TemplateInjector interface{ Template() *views.Templates }
type DebugTimestampInjector interface{ DebugTimestamp() *time.Time }
