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
	WorldSelect        = `SELECT id, randomNumber FROM World WHERE id = ?`
	WorldUpdate        = `UPDATE World SET randomNumber = ? WHERE id = ?`
	FortuneSelect      = `SELECT id, message FROM Fortune`
	WorldRowCount      = 10000
	MaxConnectionCount = 260
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
		db.Jet.SetMaxIdleConns(MaxConnectionCount)
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
		var w World
		err := db.Jet.Query(WorldSelect, rand.Intn(WorldRowCount)+1).Rows(&w)
		if err != nil {
			revel.ERROR.Fatalf("Db/WorldSelect error: %v", err)
		}
		return c.RenderJson(w)
	}

	ww := make([]World, queries)
	var wg sync.WaitGroup
	wg.Add(queries)
	acquire(queries)
	for i := 0; i < queries; i++ {
		go func(i int) {
			err := db.Jet.Query(WorldSelect, rand.Intn(WorldRowCount)+1).Rows(&ww[i])
			if err != nil {
				revel.ERROR.Fatalf("Db/WorldSelect2 error: %v", err)
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	release(queries)
	return c.RenderJson(ww)
}

func (c App) Update(queries int) revel.Result {
	if queries <= 1 {
		var w World
		err := db.Jet.Query(WorldSelect, rand.Intn(WorldRowCount)+1).Rows(&w)
		if err != nil {
			revel.ERROR.Fatalf("Update/WorldSelect error: %v", err)
		}
		w.RandomNumber = uint16(rand.Intn(WorldRowCount) + 1)
		if err = db.Jet.Query(WorldUpdate, w.RandomNumber, w.Id).Run(); err != nil {
			revel.ERROR.Fatalf("Update/WorldUpdate error: %v", err)
		}
		return c.RenderJson(&w)
	}

	var (
		ww = make([]World, queries)
		wg sync.WaitGroup
	)
	wg.Add(queries)
	acquire(queries)
	for i := 0; i < queries; i++ {
		go func(i int) {
			err := db.Jet.Query(WorldSelect, rand.Intn(WorldRowCount)+1).Rows(&ww[i])
			if err != nil {
				revel.ERROR.Fatalf("Error scanning world row: %v", err)
			}
			ww[i].RandomNumber = uint16(rand.Intn(WorldRowCount) + 1)
			if err = db.Jet.Query(WorldUpdate, ww[i].RandomNumber, ww[i].Id).Run(); err != nil {
				revel.ERROR.Fatalf("Update/WorldUpdate2 error: %v", err)
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	release(queries)
	return c.RenderJson(ww)
}

func (c App) Fortune() revel.Result {
	var fortunes Fortunes
	err := db.Jet.Query(FortuneSelect).Rows(&fortunes)
	if err != nil {
		revel.ERROR.Fatalf("Fortune/FortuneSelect error: %v", err)
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

// Utility for limiting in-flight queries to MaxConnectionLimit

var (
	conns   int
	connMu  = &sync.Mutex{}
	waiting []*Wait
)

type Wait struct {
	n    int
	cond *sync.Cond
}

func acquire(n int) {
	connMu.Lock()
	var w *Wait
	if conns+n > MaxConnectionCount {
		w = &Wait{n, sync.NewCond(connMu)}
		waiting = append(waiting, w)
		w.cond.Wait()
	} else {
		conns += n
	}
	connMu.Unlock()
}

func release(n int) {
	var w *Wait
	connMu.Lock()
	conns -= n
	if len(waiting) > 0 && waiting[0].n <= MaxConnectionCount-conns {
		w, waiting = waiting[0], waiting[1:]
		conns += w.n
	}
	connMu.Unlock()
	if w != nil {
		w.cond.Signal()
	}
}
