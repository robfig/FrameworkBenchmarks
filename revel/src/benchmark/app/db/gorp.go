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

type GorpPlugin struct {
	db.DbPlugin
}

func (p GorpPlugin) OnAppStart() {
	p.DbPlugin.OnAppStart()
	Gorp = &gorp.DbMap{Db: db.Db, Dialect: gorp.MySQLDialect{}}
	Db = db.Db
	db.Db.SetMaxIdleConns(100)
}
