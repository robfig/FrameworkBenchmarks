package db

import (
	"github.com/eaigner/jet"
	"github.com/robfig/revel"
	"os"
)

var (
	Jet    jet.Db
	Driver string
	Spec   string
)

type JetPlugin struct {
	revel.EmptyPlugin
}

func (p JetPlugin) OnAppStart() {
	// Read configuration.
	var found bool
	if Driver, found = revel.Config.String("db.driver"); !found {
		revel.ERROR.Fatal("No db.driver found.")
	}
	if Spec, found = revel.Config.String("db.spec"); !found {
		revel.ERROR.Fatal("No db.spec found.")
	}

	// Open a connection.
	var err error
	Jet, err = jet.Open(Driver, Spec)
	if err != nil {
		revel.ERROR.Fatal(err)
	}
	Jet.SetLogger(jet.NewLogger(os.Stdout))
}

type Transactional struct {
	*revel.Controller
	Txn jet.Tx
}

func (c Transactional) begin() revel.Result {
	var err error
	if c.Txn, err = Jet.Begin(); err != nil {
		panic(err)
	}
	return nil
}

func (c Transactional) commit() revel.Result {
	if err := c.Txn.Commit(); err != nil {
		panic(err)
	}
	c.Txn = nil
	return nil
}

func (c Transactional) rollback() revel.Result {
	if c.Txn == nil {
		return nil
	}
	if err := c.Txn.Rollback(); err != nil {
		panic(err)
	}
	return nil
}

func init() {
	revel.InterceptMethod(Transactional.begin, revel.BEFORE)
	revel.InterceptMethod(Transactional.commit, revel.AFTER)
	revel.InterceptMethod(Transactional.rollback, revel.PANIC)
}
