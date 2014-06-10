package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/codegangsta/martini"
	"github.com/golang/glog"
	"github.com/hobeone/pointyhair/db"
	"github.com/martini-contrib/render"
)

func createMartini(dbh *db.DBHandle) *martini.Martini {
	m := martini.New()
	m.Use(martini.Logger())
	m.Use(
		render.Renderer(
			render.Options{
				IndentJSON: true,
			},
		),
	)

	m.Use(func(w http.ResponseWriter, req *http.Request) {
		if origin := req.Header.Get("Origin"); origin != "" {
			w.Header().Add("Access-Control-Allow-Origin", origin)
		}
		w.Header().Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Add("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token")
		w.Header().Add("Access-Control-Allow-Credentials", "true")
	})
	m.Map(dbh)

	r := martini.NewRouter()
	r.Get("/api/1/people", getPeople)
	r.Get("/api/1/people/:id", getPerson)

	r.Get("/api/1/notes", getNotes)
	r.Options("/api/1/notes", send200)
	r.Post("/api/1/notes", createNote)

	r.Get("/api/1/notes/:id", getNote)
	r.Options("/api/1/notes/:id", send200)
	r.Put("/api/1/notes/:id", updateNote)

	r.Get("/api/1/todos", getTodos)
	r.Get("/api/1/todos/:id", getTodo)

	r.Post("/api/1/todos", createTodo)
	r.Options("/api/1/todos", send200)
	r.Put("/api/1/todos/:id", updateTodo)
	r.Options("/api/1/todos/:id", send200)

	m.Action(r.Handle)

	return m
}

func send200() int {
	return http.StatusOK
}

func RunWebUi(dbh *db.DBHandle) {
	m := createMartini(dbh)
	glog.Fatal(http.ListenAndServe(":3001", m))
}

func parseParamIds(str_ids []string) ([]int64, error) {
	if len(str_ids) == 0 {
		return nil, errors.New("No ids given")
	}
	int_ids := make([]int64, len(str_ids))
	for i, str_id := range str_ids {
		int_id, err := strconv.ParseInt(str_id, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("Error parsing feed id: %s", err)
		}
		int_ids[i] = int_id
	}
	return int_ids, nil
}
