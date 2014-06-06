package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hobeone/pointyhair/db"
)

const getNoteGoldenResponse = `{
  "notes": [
    {
      "id": 1,
      "date": "0001-01-01T00:00:00Z",
      "text": "http://testfeed1/feed.atom",
      "category": "",
      "person": 3
    },
    {
      "id": 2,
      "date": "0001-01-01T00:00:00Z",
      "text": "http://testfeed2/feed.atom",
      "category": "",
      "person": 3
    },
    {
      "id": 3,
      "date": "0001-01-01T00:00:00Z",
      "text": "http://testfeed3/feed.atom",
      "category": "",
      "person": 3
    }
  ]
}`

func TestGetNotes(t *testing.T) {
	dbh, m := setupTest(t)
	dbh.ORM.Begin()
	defer dbh.ORM.Rollback()
	loadFixtures(dbh)

	response := httptest.NewRecorder()

	req, _ := http.NewRequest("GET", "/api/1/notes", nil)
	m.ServeHTTP(response, req)

	if response.Code != 200 {
		fmt.Println(response.Body.String())
		t.Fatalf("Expected 200 response code, got %d", response.Code)
	}

	if response.Body.String() != getNoteGoldenResponse {
		fmt.Println(response.Body.String())
		t.Fatalf("Response doesn't match golden response")
	}
}

const getNoteWithIdsGoldenResponse = `{
  "notes": [
    {
      "id": 1,
      "date": "0001-01-01T00:00:00Z",
      "text": "http://testfeed1/feed.atom",
      "category": "",
      "person": 3
    },
    {
      "id": 2,
      "date": "0001-01-01T00:00:00Z",
      "text": "http://testfeed2/feed.atom",
      "category": "",
      "person": 3
    }
  ]
}`

func TestGetNotesWithIds(t *testing.T) {
	dbh, m := setupTest(t)

	dbh.ORM.Begin()
	defer dbh.ORM.Rollback()
	loadFixtures(dbh)

	response := httptest.NewRecorder()

	req, _ := http.NewRequest("GET", "/api/1/notes?ids[]=1&ids[]=2", nil)
	m.ServeHTTP(response, req)

	if response.Code != 200 {
		fmt.Println(response.Body.String())
		t.Fatalf("Expected 200 response code, got %d", response.Code)
	}

	if response.Body.String() != getNoteWithIdsGoldenResponse {
		fmt.Println(response.Body.String())
		t.Fatalf("Response doesn't match golden response")
	}
}

const getNoteWithIdGoldenResponse = `{
  "note": {
    "id": 1,
    "date": "0001-01-01T00:00:00Z",
    "text": "http://testfeed1/feed.atom",
    "category": "",
    "person": 3
  }
}`

func TestGetNotesById(t *testing.T) {
	dbh, m := setupTest(t)

	dbh.ORM.Begin()
	defer dbh.ORM.Rollback()
	loadFixtures(dbh)

	response := httptest.NewRecorder()

	req, _ := http.NewRequest("GET", "/api/1/notes/1", nil)
	m.ServeHTTP(response, req)

	if response.Code != 200 {
		fmt.Println(response.Body.String())
		t.Fatalf("Expected 200 response code, got %d", response.Code)
	}

	if response.Body.String() != getNoteWithIdGoldenResponse {
		fmt.Println(response.Body.String())
		t.Fatalf("Response doesn't match golden response")
	}
}

func TestCreateWithInvalidPersonId(t *testing.T) {
	dbh, m := setupTest(t)

	dbh.ORM.Begin()
	defer dbh.ORM.Rollback()
	loadFixtures(dbh)

	test_person_id := int64(1000)
	_, err := dbh.GetPersonById(test_person_id)
	if err == nil {
		t.Fatal("Person with ID 1000 shouldn't exist")
	}

	tdate := time.Now()
	n := NoteJSON{
		Note: noteWithPersonIdJSON{
			&db.Note{
				Text: "testtext",
				Date: tdate,
			},
			test_person_id,
		},
	}
	req_body, err := json.Marshal(n)
	failOnError(t, err)

	response := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/1/notes", bytes.NewReader(req_body))
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	m.ServeHTTP(response, req)

	if response.Code != 500 {
		fmt.Println(response.Body.String())
		t.Fatalf("Expected 500 response code, got %d", response.Code)
	}
}

func TestCreateNote(t *testing.T) {
	dbh, m := setupTest(t)

	dbh.ORM.Begin()
	defer dbh.ORM.Rollback()
	loadFixtures(dbh)

	tdate := time.Now()
	n := NoteJSON{
		Note: noteWithPersonIdJSON{
			&db.Note{
				Text: "testtext",
				Date: tdate,
			},
			1,
		},
	}
	req_body, err := json.Marshal(n)
	failOnError(t, err)

	response := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/1/notes", bytes.NewReader(req_body))
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	m.ServeHTTP(response, req)

	if response.Code != 200 {
		fmt.Println(response.Body.String())
		t.Fatalf("Expected 200 response code, got %d", response.Code)
	}

	u := unmarshalNoteJSONContainer{}
	err = json.NewDecoder(response.Body).Decode(&u)
	failOnError(t, err)
	if u.Note.Date != tdate {
		t.Fatalf("Note Date doesn't match test Date: %v != %v",
			u.Note.Date, tdate)
	}

	if u.Note.Text != "testtext" {
		t.Fatalf("Note Text doesn't match test text: %s != %s",
			u.Note.Text, "testtext")
	}
}

func TestUpdateNote(t *testing.T) {
	dbh, m := setupTest(t)

	dbh.ORM.Begin()
	defer dbh.ORM.Rollback()
	loadFixtures(dbh)

	dbnote, err := dbh.GetNoteById(1)
	failOnError(t, err)

	new_text_string := "New Text STRING"
	dbnote.Text = new_text_string
	n := NoteJSON{
		Note: noteWithPersonIdJSON{
			dbnote,
			1,
		},
	}
	req_body, err := json.Marshal(n)
	failOnError(t, err)

	response := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/1/notes/1", bytes.NewReader(req_body))
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	m.ServeHTTP(response, req)

	if response.Code != 200 {
		fmt.Println(response.Body.String())
		t.Fatalf("Expected 200 response code, got %d", response.Code)
	}

	dbnote, err = dbh.GetNoteById(1)
	failOnError(t, err)
	if dbnote.Text != new_text_string {
		t.Fatalf("text field wasn't updated")
	}
}

func TestUpdateNonExistingNote(t *testing.T) {
	dbh, m := setupTest(t)

	dbh.ORM.Begin()
	defer dbh.ORM.Rollback()
	loadFixtures(dbh)
	non_existing_id := int64(100)

	_, err := dbh.GetNoteById(non_existing_id)
	if err == nil {
		t.Fatalf("Didn't expect to find note with id 100")
	}

	n := NoteJSON{
		Note: noteWithPersonIdJSON{
			&db.Note{},
			1,
		},
	}
	req_body, err := json.Marshal(n)
	failOnError(t, err)

	response := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/1/notes/%d", non_existing_id), bytes.NewReader(req_body))
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	m.ServeHTTP(response, req)

	if response.Code != 404 {
		fmt.Println(response.Body.String())
		t.Fatalf("Expected 404 response code, got %d", response.Code)
	}
}
