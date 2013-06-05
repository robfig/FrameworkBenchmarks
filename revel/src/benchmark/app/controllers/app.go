package controllers

import (
	"benchmark/app/db"
	"github.com/robfig/revel"
	"math/rand"
	"runtime"
	"sort"
)

type MessageStruct struct {
	Message string `json:"message"`
}

type World struct {
	Id           uint16 `json:"id"`
	RandomNumber uint16 `json:"randomNumber"`
}

type Fortune struct {
	Id      uint16 `json:"id"`
	Message string `json:"message"`
}

const (
	WorldRowCount      = 10000
	MaxConnectionCount = 256
)

func init() {
	revel.Filters = []revel.Filter{
		revel.RouterFilter,
		revel.ParamsFilter,
		revel.ActionInvoker,
	}
	revel.OnAppStart(func() {
		runtime.GOMAXPROCS(runtime.NumCPU())
		db.Init()
		db.Db.SetMaxIdleConns(MaxConnectionCount)
		db.Gorp.AddTable(World{}).
			SetKeys(true, "Id")
		db.Gorp.AddTable(Fortune{}).
			SetKeys(true, "Id")
	})
}

type App struct {
	*revel.Controller
}

func (c App) Json() revel.Result {
	return c.RenderJson(MessageStruct{"Hello, world"})
}

func (c App) Plaintext() revel.Result {
	return c.RenderText("Hello, World!")
}

func (c App) Db(queries int) revel.Result {
	if queries <= 1 {
		w, err := db.Gorp.Get(World{}, rand.Intn(WorldRowCount)+1)
		if err != nil {
			revel.ERROR.Fatalf("Error scanning world row: %v", err)
		}
		return c.RenderJson(w.(*World))
	}

	ww := make([]*World, queries)
	for i := 0; i < queries; i++ {
		result, err := db.Gorp.Get(World{}, rand.Intn(WorldRowCount)+1)
		if err != nil {
			revel.ERROR.Fatalf("Error scanning world row: %v", err)
		}
		ww[i] = result.(*World)
	}
	return c.RenderJson(ww)
}

func (c App) Update(queries int) revel.Result {
	if queries <= 1 {
		result, err := db.Gorp.Get(World{}, rand.Intn(WorldRowCount)+1)
		w := result.(*World)
		w.RandomNumber = uint16(rand.Intn(WorldRowCount) + 1)
		_, err = db.Gorp.Update(w)
		if err != nil {
			revel.ERROR.Fatalf("Error updating row: %s", err)
		}
		return c.RenderJson(w)
	}

	ww := make([]*World, queries)
	for i := 0; i < queries; i++ {
		result, err := db.Gorp.Get(World{}, rand.Intn(WorldRowCount)+1)
		if err != nil {
			revel.ERROR.Fatalf("Error scanning world row: %v", err)
		}
		ww[i] = result.(*World)
		ww[i].RandomNumber = uint16(rand.Intn(WorldRowCount) + 1)
		_, err = db.Gorp.Update(ww[i])
		if err != nil {
			revel.ERROR.Fatalf("Error scanning world row: %v", err)
		}
	}
	return c.RenderJson(ww)
}

func (c App) Fortune() revel.Result {
	var fortunes Fortunes
	_, err := db.Gorp.Select(&fortunes, "SELECT * FROM Fortune")
	if err != nil {
		revel.ERROR.Fatalf("Error selecting fortunes: %v", err)
	}
	fortunes = append(fortunes,
		&Fortune{Message: "Additional fortune added at request time."})

	sort.Sort(ByMessage{fortunes})
	return c.Render(fortunes)
}

type Fortunes []*Fortune

func (s Fortunes) Len() int      { return len(s) }
func (s Fortunes) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

type ByMessage struct{ Fortunes }

func (s ByMessage) Less(i, j int) bool { return s.Fortunes[i].Message < s.Fortunes[j].Message }
