package app

import (
	coresql "github.com/LUSHDigital/core-sql"
	"github.com/gorilla/mux"
)

type DependencyInjector interface {
	MySQLInjector
	RouterInjector
}

type MySQLInjector interface{ MySQL() coresql.Agent }
type RouterInjector interface{ Router() *mux.Router }
