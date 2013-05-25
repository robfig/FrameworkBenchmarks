package db

import (
	"github.com/eaigner/hood"
	"github.com/robfig/revel/modules/db/app"
)

// Not save for concurrent usage, use Hood.Copy() to get a local handle.
var Hood *hood.Hood

func Init() {
	db.DbPlugin{}.OnAppStart()
	Hood = hood.New(db.Db, hood.NewMysql())
	db.Db.SetMaxIdleConns(100)
}
