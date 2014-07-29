package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/codegangsta/martini"
	"github.com/golang/glog"
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
		db_todos := []*db.Todo{}
		_, err := dbh.ORM.QueryTable("todo").All(&db_todos)
		if err != nil {
			rend.JSON(500, err.Error())
			return
		}
		todos = make([]todoWithPersonIdJSON, len(db_todos))
		for i, todo := range db_todos {
			todos[i] = todoWithPersonIdJSON{todo, todo.Person.Id}
		}
	}
	rend.JSON(http.StatusOK, todos)
}

func getTodo(rend render.Render, req *http.Request, params martini.Params, dbh *db.DBHandle) {
	id, err := strconv.ParseInt(params["id"], 10, 64)
	if err != nil {
		rend.JSON(500, fmt.Sprintf("Invalid id %s: %s", params["id"], err.Error()))
		return
	}
	p := db.Todo{Id: id}
	err = dbh.ORM.Read(&p)

	if err != nil {
		if err == orm.ErrNoRows {
			rend.JSON(404, fmt.Sprintf("No Todo with id %d found.", id))
			return
		} else {
			rend.JSON(500, err.Error())
			return
		}
	}

	rend.JSON(200, todoWithPersonIdJSON{&p, p.Person.Id})
}

func createTodo(rend render.Render, req *http.Request, params martini.Params, dbh *db.DBHandle) {
	err := req.ParseForm()
	if err != nil {
		rend.JSON(500, err.Error())
		return
	}
	u := unmarshalTodoJSON{}
	glog.Info("Decoded json: %+v", u)
	err = json.NewDecoder(req.Body).Decode(&u)
	if err != nil {
		rend.JSON(500, err.Error())
		return
	}

	queryParams, _ := url.ParseQuery(req.URL.RawQuery)
	dbtodo := db.Todo{
		Subject:  u.Subject,
		Category: u.Category,
		Date:     u.Date,
	}

	if _, ok := queryParams["addToAll"]; ok {
		glog.Info(u)
		//do something here
		err = dbh.AddTodoToAllPeople(&dbtodo)
		if err != nil {
			rend.JSON(http.StatusInternalServerError, err.Error())
			return
		}
	} else {
		p := db.Person{Id: u.PersonId}
		err = dbh.ORM.Read(&p)
		if err != nil {
			rend.JSON(500, fmt.Sprintf("Unknown Person id: %d", u.PersonId))
			return
		}

		dbtodo.Person = &p
		_, err = dbh.ORM.Insert(&dbtodo)
		if err != nil {
			rend.JSON(500, err.Error())
			return
		}
		rend.JSON(200, todoWithPersonIdJSON{&dbtodo, p.Id})
	}
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
		if err == orm.ErrNoRows {
			rend.JSON(404, fmt.Sprintf("No Todo with id %d found.", todo_id))
			return
		} else {
			rend.JSON(500, err.Error())
			return
		}
	}
	if u.Todo.Subject != "" {
		dbtodo.Subject = u.Todo.Subject
	}
	if u.Todo.Category != "" {
		dbtodo.Category = u.Todo.Category
	}
	dbh.ORM.Update(&dbtodo)
	rend.JSON(200, todoWithPersonIdJSON{&dbtodo, dbtodo.Person.Id})
}

func deleteTodo(rend render.Render, params martini.Params, dbh *db.DBHandle) {
	todo_id, err := strconv.ParseInt(params["id"], 10, 64)
	if err != nil {
		rend.JSON(http.StatusBadRequest, err.Error())
		return
	}

	todo, err := dbh.GetTodoById(todo_id)
	if err != nil {
		rend.JSON(http.StatusNotFound, err.Error())
		return
	}

	err = dbh.RemoveTodo(todo)
	if err != nil {
		rend.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	rend.JSON(http.StatusNoContent, "")
}
