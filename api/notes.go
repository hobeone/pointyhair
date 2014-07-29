package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/codegangsta/martini"
	"github.com/hobeone/pointyhair/db"
	"github.com/martini-contrib/render"
)

// To Convert From/To Ember Data format
type noteWithPersonIdJSON struct {
	*db.Note
	PersonId int64 `json:"person"`
}

type NoteJSON struct {
	Note noteWithPersonIdJSON `json:"note"`
}

type NotesJSON struct {
	Notes []noteWithPersonIdJSON `json:"notes"`
}

type unmarshalNoteJSON struct {
	Id       int       `json:"id"`
	Text     string    `json:"text"`
	Category string    `json:"category"`
	Date     time.Time `json:"date"`
	PersonId int64     `json:"person"`
}

type unmarshalNoteJSONContainer struct {
	Note unmarshalNoteJSON `json:"note"`
}

func createNote(rend render.Render, req *http.Request, params martini.Params, dbh *db.DBHandle) {
	err := req.ParseForm()
	if err != nil {
		rend.JSON(http.StatusBadRequest, err.Error())
		return
	}
	u := unmarshalNoteJSON{}
	err = json.NewDecoder(req.Body).Decode(&u)
	if err != nil {
		rend.JSON(http.StatusBadRequest, err.Error())
		return
	}

	p := db.Person{Id: u.PersonId}
	err = dbh.ORM.Read(&p)
	if err != nil {
		rend.JSON(http.StatusInternalServerError, fmt.Sprintf("Unknown Person ID: %d", u.PersonId))
		return
	}

	dbnote := db.Note{
		Text:     u.Text,
		Category: u.Category,
		Date:     u.Date,
		Person:   &p,
	}
	err = dbh.CreateNote(&dbnote)
	if err != nil {
		rend.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	rend.JSON(200, noteWithPersonIdJSON{&dbnote, p.Id})
}

func deleteNote(rend render.Render, params martini.Params, dbh *db.DBHandle) {
	note_id, err := strconv.ParseInt(params["id"], 10, 64)
	if err != nil {
		rend.JSON(http.StatusBadRequest, err.Error())
		return
	}

	note, err := dbh.GetNoteById(note_id)
	if err != nil {
		rend.JSON(http.StatusNotFound, err.Error())
		return
	}

	err = dbh.RemoveNote(note)
	if err != nil {
		rend.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	rend.JSON(http.StatusNoContent, "")
}

func updateNote(rend render.Render, req *http.Request, params martini.Params, dbh *db.DBHandle) {
	note_id, err := strconv.ParseInt(params["id"], 10, 64)
	if err != nil {
		rend.JSON(500, err.Error())
		return
	}

	err = req.ParseForm()
	if err != nil {
		rend.JSON(500, err.Error())
		return
	}

	u := unmarshalNoteJSONContainer{}
	err = json.NewDecoder(req.Body).Decode(&u)
	if err != nil {
		rend.JSON(500, err.Error())
		return
	}

	dbnote := db.Note{Id: note_id}
	err = dbh.ORM.Read(&dbnote)
	if err != nil {
		rend.JSON(404, err.Error())
		return
	}

	if u.Note.Text != "" {
		dbnote.Text = u.Note.Text
	}
	if u.Note.Category != "" {
		dbnote.Category = u.Note.Category
	}
	dbh.ORM.Update(&dbnote)
}

func getNote(rend render.Render, req *http.Request, params martini.Params, dbh *db.DBHandle) {
	id, err := strconv.ParseInt(params["id"], 10, 64)
	if err != nil {
		rend.JSON(500, "Invalid id: "+err.Error())
		return
	}
	n, err := dbh.GetNoteById(id)
	if err != nil {
		rend.JSON(500, err.Error())
		return
	}

	rend.JSON(200, noteWithPersonIdJSON{n, n.Person.Id})
}

func getNotes(rend render.Render, req *http.Request, dbh *db.DBHandle) {
	err := req.ParseForm()
	if err != nil {
		rend.JSON(500, err.Error())
		return
	}

	var notes []noteWithPersonIdJSON
	param_ids := req.Form["ids[]"]
	if len(param_ids) > 0 {
		note_ids, err := parseParamIds(param_ids)
		if err != nil {
			rend.JSON(500, err.Error())
			return
		}

		notes = make([]noteWithPersonIdJSON, len(note_ids))
		for i, nid := range note_ids {
			note, err := dbh.GetNoteById(nid)
			if err != nil {
				rend.JSON(404, err.Error())
				return
			}
			notes[i] = noteWithPersonIdJSON{note, note.Person.Id}
		}
	} else {
		dbnotes, err := dbh.GetNotesById([]int64{})
		notes = make([]noteWithPersonIdJSON, len(dbnotes))
		for i, n := range dbnotes {
			notes[i] = noteWithPersonIdJSON{n, n.Person.Id}
		}
		if err != nil {
			rend.JSON(500, err.Error())
			return
		}
	}
	rend.JSON(http.StatusOK, notes)
}
