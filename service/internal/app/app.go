package app

import (
	coresql "github.com/LUSHDigital/core-sql"
	"github.com/gorilla/mux"
	"prediction-league/service/internal/domain"
)

type DependencyInjector interface {
	ConfigInjector
	MySQLInjector
	RouterInjector
}

type ConfigInjector interface{ Config() domain.Config }
type MySQLInjector interface{ MySQL() coresql.Agent }
type RouterInjector interface{ Router() *mux.Router }
