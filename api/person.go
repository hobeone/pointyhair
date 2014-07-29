package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/astaxie/beego/orm"
	"github.com/codegangsta/martini"
	"github.com/golang/glog"
	"github.com/hobeone/pointyhair/db"
	"github.com/martini-contrib/render"
)

type PeopleJSON struct {
	People []*personWithRelations `json:"people"`
}

type PersonJSON struct {
	Person personWithRelations `json:"person"`
}

type personWithRelations struct {
	db.Person
	Notes []*db.Note `json:"notes"`
	Todos []*db.Todo `json:"todos"`
	//	NoteIds []int64 `json:"notes"`
	//	TodoIds []int64 `json:"todos"`
}

type unmarshalPersonJSON struct {
	Name string `json:"name"`
}

func getPerson(rend render.Render, params martini.Params, dbh *db.DBHandle) {
	id, err := strconv.ParseInt(params["id"], 10, 64)
	if err != nil {
		rend.JSON(500, "Invalid id: "+err.Error())
		return
	}
	p, err := dbh.GetPersonById(id)
	if err != nil {
		if err == orm.ErrNoRows {
			rend.JSON(404, fmt.Sprintf("No Person with id %d found.", id))
			return
		} else {
			rend.JSON(500, err.Error())
			return
		}
	}

	pn, err := newPersonWithRelations(p, dbh)
	if err != nil {
		rend.JSON(500, err)
		return
	}
	rend.JSON(200, pn)
}

func newPersonWithRelations(p *db.Person, dbh *db.DBHandle) (personWithRelations, error) {
	err := p.LoadRelated(dbh)
	if err != nil {
		return personWithRelations{}, err
	}
	note_ids := make([]int64, len(p.Notes))
	for ni, n := range p.Notes {
		note_ids[ni] = n.Id
	}
	todo_ids := make([]int64, len(p.Todos))
	for ti, t := range p.Todos {
		todo_ids[ti] = t.Id
	}

	pn := personWithRelations{
		*p,
		p.Notes,
		p.Todos,
	}
	return pn, nil
}

func getPeople(rend render.Render, req *http.Request, dbh *db.DBHandle) {
	err := req.ParseForm()
	if err != nil {
		rend.JSON(500, err.Error())
		return
	}

	var people_json []*personWithRelations
	param_ids := req.Form["ids[]"]
	if len(param_ids) > 0 {
		people_ids, err := parseParamIds(param_ids)
		if err != nil {
			rend.JSON(500, err.Error())
			return
		}
		people_json = make([]*personWithRelations, len(people_ids))
		for i, pid := range people_ids {
			person, err := dbh.GetPersonById(pid)
			if err != nil {
				if err == orm.ErrNoRows {
					rend.JSON(404, fmt.Sprintf("Person with ID %d doesn't exist", pid))
					return
				} else {
					rend.JSON(500, err.Error())
					return
				}
			}
			pn, err := newPersonWithRelations(person, dbh)
			if err != nil {
				rend.JSON(500, err)
				return
			}

			people_json[i] = &pn
		}
	} else {
		people, err := dbh.GetPeopleById([]int64{})
		if err != nil {
			rend.JSON(500, err)
			return
		}

		people_json = make([]*personWithRelations, len(people))
		for i, p := range people {
			pn, err := newPersonWithRelations(p, dbh)
			if err != nil {
				rend.JSON(500, err)
				return
			}
			people_json[i] = &pn
		}
	}
	rend.JSON(200, people_json)
}

func createPerson(rend render.Render, req *http.Request, params martini.Params, dbh *db.DBHandle) {
	err := req.ParseForm()
	if err != nil {
		rend.JSON(http.StatusBadRequest, err.Error())
		return
	}

	u := unmarshalPersonJSON{}
	glog.Info("Decoding person creation request")
	err = json.NewDecoder(req.Body).Decode(&u)
	if err != nil {
		rend.JSON(http.StatusBadRequest, err.Error())
		return
	}

	dbPerson := db.Person{
		Name: u.Name,
	}

	_, err = dbh.ORM.Insert(&dbPerson)
	if err != nil {
		rend.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	pn, err := newPersonWithRelations(&dbPerson, dbh)
	if err != nil {
		rend.JSON(500, err)
		return
	}
	rend.JSON(http.StatusOK, pn)
}
