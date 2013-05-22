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

const WorldRowCount = 10000

func init() {
	//revel.RegisterPlugin(db.GorpPlugin{})
	revel.OnAppStart(func() {
		runtime.GOMAXPROCS(runtime.NumCPU())
		db.GorpPlugin{}.OnAppStart()
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
	c.Response.ContentType = "application/json"
	return c.RenderJson(MessageStruct{"Hello, world"})
}

func (c App) Db(queries int) revel.Result {
	if queries <= 1 {
		rowNum := rand.Intn(WorldRowCount) + 1
		w, err := db.Gorp.Get(World{}, rowNum)
		if err != nil {
			revel.ERROR.Fatalf("Error scanning world row: %v", err)
		}
		return c.RenderJson(w.(*World))
	}

	ww := make([]*World, queries)
	var wg sync.WaitGroup
	wg.Add(queries)
	for i := 0; i < queries; i++ {
		go func(i int) {
			rowNum := rand.Intn(WorldRowCount) + 1
			result, err := db.Gorp.Get(World{}, rowNum)
			if err != nil {
				revel.ERROR.Fatalf("Error scanning world row: %v", err)
			}
			ww[i] = result.(*World)
			wg.Done()
		}(i)
	}
	wg.Wait()
	return c.RenderJson(ww)
}

func (c App) Update(queries int) revel.Result {
	rowNum := rand.Intn(WorldRowCount) + 1
	if queries <= 1 {
		var w World
		w, err := db.Gorp.Get(World{}, rowNum)
		w.RandomNumber = uint16(rand.Intn(WorldRowCount) + 1)
		_, err := db.Gorp.Update(&w)
		if err != nil {
			return revel.ERROR.Fatalf("Error updating row: %s", err)
		}
		return c.RenderJson(&w)
	}

	var (
		ww = make([]World, queries)
		wg sync.WaitGroup
	)
	wg.Add(queries)
	for i := 0; i < queries; i++ {
		go func(i int) {
			rowNum := rand.Intn(WorldRowCount) + 1
			result, err := db.Gorp.Get(World{}, rowNum)
			if err != nil {
				revel.ERROR.Fatalf("Error scanning world row: %v", err)
			}
			ww[i] = result.(*World)
			ww[i].RandomNumber = uint16(rand.Intn(WorldRowCount) + 1)
			_, err = db.Gorp.Update(&ww[i])
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
	results, err := db.Gorp.Select(Fortune{}, "SELECT * FROM Fortune")
	if err != nil {
		revel.ERROR.Fatalf("Error preparing statement: %v", err)
	}

	var fortunes Fortunes
	for _, r := range results {
		fortunes = append(fortunes, r.(*Fortune))
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
