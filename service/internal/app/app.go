package app

import (
	coresql "github.com/LUSHDigital/core-sql"
	"github.com/gorilla/mux"
	"prediction-league/service/internal/domain"
	"prediction-league/service/internal/views"
)

type DependencyInjector interface {
	ConfigInjector
	MySQLInjector
	RouterInjector
	TemplateInjector
}

type ConfigInjector interface{ Config() domain.Config }
type MySQLInjector interface{ MySQL() coresql.Agent }
type RouterInjector interface{ Router() *mux.Router }
type TemplateInjector interface{ Template() *views.Templates }
