package controllers

import (
	"benchmark/app/db"
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
	WorldRowCount      = 10000
	MaxConnectionCount = 100
)

func init() {
	revel.RegisterPlugin(db.JetPlugin{})
	revel.OnAppStart(func() {
		runtime.GOMAXPROCS(runtime.NumCPU())
		//db.Jet.Db.SetMaxIdleConns(MaxConnectionCount)
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
	const worldSelect = "SELECT id,randomNumber FROM World where id = "
	rowNum := rand.Intn(WorldRowCount) + 1
	if queries <= 1 {
		var w World
		err := db.Jet.Query(worldSelect, rowNum).Value(&w)
		if err != nil {
			revel.ERROR.Fatalf("Error scanning world row: %v", err)
		}
		return c.RenderJson(w)
	}

	ww := make([]World, queries)
	var wg sync.WaitGroup
	wg.Add(queries)
	for i := 0; i < queries; i++ {
		go func(i int) {
			err := db.Jet.Query(worldSelect, rowNum).Value(&ww[i])
			if err != nil {
				revel.ERROR.Fatalf("Error scanning world row: %v", err)
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	return c.RenderJson(ww)
}

func (c App) Fortune() revel.Result {
	const fortuneSelect = "SELECT id,message FROM Fortune"
	var fortunes Fortunes
	err := db.Jet.Query(fortuneSelect).Rows(&fortunes)
	if err != nil {
		revel.ERROR.Fatalf("Error preparing statement: %v", err)
	}
	fortunes = append(fortunes, &Fortune{Message: "Additional fortune added at request time."})
	sort.Sort(ByMessage{fortunes})
	return c.Render(fortunes)
}

type Fortunes []*Fortune

func (s Fortunes) Len() int      { return len(s) }
func (s Fortunes) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

type ByMessage struct{ Fortunes }

func (s ByMessage) Less(i, j int) bool { return s.Fortunes[i].Message < s.Fortunes[j].Message }
