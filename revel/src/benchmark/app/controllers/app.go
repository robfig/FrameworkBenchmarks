package controllers

import (
	"benchmark/app/db"
	"github.com/coocood/qbs"
	"github.com/robfig/revel"
	"math/rand"
	"runtime"
	"sort"
	"sync"
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
	WorldRowCount = 10000
)

func init() {
	revel.OnAppStart(func() {
		runtime.GOMAXPROCS(runtime.NumCPU())
		db.Init()
	})
}

type App struct {
	*revel.Controller
}

func (c App) Json() revel.Result {
	c.Response.ContentType = "application/json"
	return c.RenderJson(MessageStruct{"Hello, world"})
}

func (c App) Db(queries int) revel.Result {
	qbs, _ := qbs.GetQbs()
	defer qbs.Close()
	qbs.Log = true

	if queries <= 1 {
		rowNum := uint16(rand.Intn(WorldRowCount) + 1)
		var w World
		w.Id = rowNum
		qbs.Find(&w)
		return c.RenderJson(w)
	}

	ww := make([]World, queries)
	var wg sync.WaitGroup
	wg.Add(queries)
	for i := 0; i < queries; i++ {
		go func(i int) {
			rowNum := uint16(rand.Intn(WorldRowCount) + 1)
			ww[i].Id = rowNum
			qbs.Find(&ww[i])
			wg.Done()
		}(i)
	}
	wg.Wait()
	return c.RenderJson(ww)
}

func (c App) Fortune() revel.Result {
	qbs, _ := qbs.GetQbs()
	defer qbs.Close()
	qbs.Log = true

	var fortunes []*Fortune
	qbs.FindAll(&fortunes)
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
