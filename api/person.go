package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/codegangsta/martini"
	"github.com/hobeone/pointyhair/db"
	"github.com/martini-contrib/render"
)

type PeopleJSON struct {
	People []*personWithNotes `json:"people"`
}

type personWithNotes struct {
	db.Person
	NoteIds []int64 `json:"notes"`
}

func getPerson(rend render.Render, params martini.Params, dbh *db.DBHandle) {
	id, err := strconv.ParseInt(params["id"], 10, 64)
	if err != nil {
		rend.JSON(500, "Invalid id: "+err.Error())
		return
	}
	p, err := dbh.GetPersonById(id)
	if err != nil {
		rend.JSON(500, err.Error())
		return
	}
	rend.JSON(200, p)
}

func getPeople(rend render.Render, req *http.Request, dbh *db.DBHandle) {
	err := req.ParseForm()
	if err != nil {
		rend.JSON(500, err.Error())
		return
	}

	var people_json []*personWithNotes
	param_ids := req.Form["ids[]"]
	if len(param_ids) > 0 {
		people_ids, err := parseParamIds(param_ids)
		if err != nil {
			rend.JSON(500, err.Error())
			return
		}
		people_json = make([]*personWithNotes, len(people_ids))
		for i, pid := range people_ids {
			person, err := dbh.GetPersonById(pid)
			if err != nil {
				rend.JSON(404, fmt.Sprintf("Person with ID %d doesn't exist", pid))
				return
			}
			_, err = dbh.ORM.LoadRelated(person, "Notes")
			if err != nil {
				rend.JSON(500, err.Error())
				return
			}
			note_ids := make([]int64, len(person.Notes))
			for ni, n := range person.Notes {
				note_ids[ni] = n.Id
			}

			people_json[i] = &personWithNotes{*person, note_ids}
		}
	} else {
		people, err := dbh.GetPeopleById([]int64{})
		if err != nil {
			rend.JSON(500, err)
			return
		}

		people_json = make([]*personWithNotes, len(people))
		for i, p := range people {
			_, err = dbh.ORM.LoadRelated(p, "Notes")
			if err != nil {
				rend.JSON(500, err.Error())
				return
			}
			note_ids := make([]int64, len(p.Notes))
			for ni, n := range p.Notes {
				note_ids[ni] = n.Id
			}
			people_json[i] = &personWithNotes{*p, note_ids}
		}
	}
	rend.JSON(200, PeopleJSON{People: people_json})
}
