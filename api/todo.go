package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/codegangsta/martini"
	"github.com/hobeone/pointyhair/db"
	"github.com/martini-contrib/render"
)

type todoWithPersonIdJSON struct {
	*db.Todo
	PersonId int64 `json:"person"`
}

type TodoJSON struct {
	Todo todoWithPersonIdJSON `json:"todo"`
}

type TodosJSON struct {
	Todos []todoWithPersonIdJSON `json:"todos"`
}

type unmarshalTodoJSON struct {
	Id       int       `json:"id"`
	Subject  string    `json:"subject"`
	Category string    `json:"category"`
	Date     time.Time `json:"date"`
	PersonId int64     `json:"person"`
}

type unmarshalTodoJSONContainer struct {
	Todo unmarshalTodoJSON `json:"todo"`
}

func getTodos(rend render.Render, req *http.Request, dbh *db.DBHandle) {
	err := req.ParseForm()
	if err != nil {
		rend.JSON(500, err.Error())
		return
	}
	var todos []todoWithPersonIdJSON
	param_ids := req.Form["ids[]"]
	if len(param_ids) > 0 {
		todo_ids, err := parseParamIds(param_ids)
		if err != nil {
			rend.JSON(500, err.Error())
			return
		}

		todos = make([]todoWithPersonIdJSON, len(todo_ids))
		for i, tid := range todo_ids {
			todo := db.Todo{Id: tid}
			err = dbh.ORM.Read(&todo)
			if err != nil {
				rend.JSON(404, err.Error())
				return
			}
			todos[i] = todoWithPersonIdJSON{&todo, todo.Person.Id}
		}
	} else {
		_, err := dbh.ORM.QueryTable("todo").All(&todos)
		if err != nil {
			rend.JSON(500, err.Error())
			return
		}
	}
	rend.JSON(http.StatusOK, TodosJSON{Todos: todos})
}

func getTodo(rend render.Render, req *http.Request, params martini.Params, dbh *db.DBHandle) {
	id, err := strconv.ParseInt(params["id"], 10, 64)
	if err != nil {
		rend.JSON(500, "Invalid id: "+err.Error())
		return
	}
	p := db.Todo{Id: id}
	err = dbh.ORM.Read(&p)
	if err != nil {
		rend.JSON(500, err.Error())
		return
	}

	rend.JSON(200, TodoJSON{Todo: todoWithPersonIdJSON{&p, p.Person.Id}})
}

func createTodo(rend render.Render, req *http.Request, params martini.Params, dbh *db.DBHandle) {
	err := req.ParseForm()
	if err != nil {
		rend.JSON(500, err.Error())
		return
	}
	u := unmarshalTodoJSONContainer{}
	err = json.NewDecoder(req.Body).Decode(&u)
	if err != nil {
		rend.JSON(500, err.Error())
		return
	}

	p := db.Person{Id: u.Todo.PersonId}
	err = dbh.ORM.Read(&p)
	if err != nil {
		rend.JSON(500, "Unknown Person ID")
		return
	}

	dbtodo := db.Todo{
		Subject:  u.Todo.Subject,
		Category: u.Todo.Category,
		Date:     u.Todo.Date,
		Person:   &p,
	}
	_, err = dbh.ORM.Insert(&dbtodo)
	if err != nil {
		rend.JSON(500, err.Error())
		return
	}
	rend.JSON(200, TodoJSON{Todo: todoWithPersonIdJSON{&dbtodo, p.Id}})
}

func updateTodo(rend render.Render, req *http.Request, params martini.Params, dbh *db.DBHandle) {
	todo_id, err := strconv.ParseInt(params["id"], 10, 64)
	if err != nil {
		rend.JSON(500, err.Error())
		return
	}

	err = req.ParseForm()
	if err != nil {
		rend.JSON(500, err.Error())
		return
	}

	u := unmarshalTodoJSONContainer{}
	err = json.NewDecoder(req.Body).Decode(&u)
	if err != nil {
		rend.JSON(500, err.Error())
		return
	}

	dbtodo := db.Todo{Id: todo_id}
	err = dbh.ORM.Read(&dbtodo)
	if err != nil {
		rend.JSON(404, err.Error())
		return
	}

	if u.Todo.Subject != "" {
		dbtodo.Subject = u.Todo.Subject
	}
	if u.Todo.Category != "" {
		dbtodo.Category = u.Todo.Category
	}
	dbh.ORM.Update(&dbtodo)
}
