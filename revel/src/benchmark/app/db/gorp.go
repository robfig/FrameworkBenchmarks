package db

import (
	"database/sql"
	"github.com/coopernurse/gorp"
	"github.com/robfig/revel/modules/db/app"
)

var (
	Db   *sql.DB
	Gorp *gorp.DbMap
)

func Init() {
	db.Init()
	Gorp = &gorp.DbMap{Db: db.Db, Dialect: gorp.MySQLDialect{}}
	Db = db.Db
}
