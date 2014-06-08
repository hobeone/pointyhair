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

const getTodoGoldenResponse = `{
  "todo": {
    "id": 1,
    "subject": "test todo1",
    "date": "0001-01-01T00:00:00Z",
    "category": "",
    "Notes": "test todo1 notes",
    "person": 3
  }
}`

func TestGetTodo(t *testing.T) {
	dbh, m := setupTest(t)
	dbh.ORM.Begin()
	defer dbh.ORM.Rollback()
	loadFixtures(dbh)

	response := httptest.NewRecorder()

	req, _ := http.NewRequest("GET", "/api/1/todos/1", nil)
	m.ServeHTTP(response, req)

	if response.Code != http.StatusOK {
		fmt.Println(response.Body.String())
		t.Fatalf("Expected %d response code, got %d", http.StatusOK, response.Code)
	}

	if response.Body.String() != getTodoGoldenResponse {
		fmt.Println(response.Body.String())
		t.Fatalf("Response doesn't match golden response")
	}

	// Non existing id
	response = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/1/todos/100", nil)
	m.ServeHTTP(response, req)
	if response.Code != http.StatusNotFound {
		fmt.Println(response.Body.String())
		t.Fatalf("Expected %d response code, got %d", http.StatusNotFound, response.Code)
	}

}

const getTodosGoldenResponse = `{
  "todos": [
    {
      "id": 1,
      "subject": "test todo1",
      "date": "0001-01-01T00:00:00Z",
      "category": "",
      "Notes": "test todo1 notes",
      "person": 3
    },
    {
      "id": 2,
      "subject": "test todo2",
      "date": "0001-01-01T00:00:00Z",
      "category": "",
      "Notes": "test todo2 notes",
      "person": 3
    },
    {
      "id": 3,
      "subject": "test todo3",
      "date": "0001-01-01T00:00:00Z",
      "category": "",
      "Notes": "test todo3 notes",
      "person": 3
    }
  ]
}`

const getTodosByIdGoldenResponse = `{
  "todos": [
    {
      "id": 1,
      "subject": "test todo1",
      "date": "0001-01-01T00:00:00Z",
      "category": "",
      "Notes": "test todo1 notes",
      "person": 3
    },
    {
      "id": 2,
      "subject": "test todo2",
      "date": "0001-01-01T00:00:00Z",
      "category": "",
      "Notes": "test todo2 notes",
      "person": 3
    }
  ]
}`

func TestGetTodos(t *testing.T) {
	dbh, m := setupTest(t)

	dbh.ORM.Begin()
	defer dbh.ORM.Rollback()
	loadFixtures(dbh)

	//ALL
	response := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/1/todos", nil)
	m.ServeHTTP(response, req)
	if response.Code != http.StatusOK {
		fmt.Println(response.Body.String())
		t.Fatalf("Expected %d response code, got %d", http.StatusOK, response.Code)
	}

	if response.Body.String() != getTodosGoldenResponse {
		fmt.Println(response.Body.String())
		t.Fatalf("Response doesn't match golden response")
	}

	// By ID
	response = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/1/todos?ids[]=1&ids[]=2", nil)
	m.ServeHTTP(response, req)
	if response.Code != http.StatusOK {
		fmt.Println(response.Body.String())
		t.Fatalf("Expected %d response code, got %d", http.StatusOK, response.Code)
	}

	if response.Body.String() != getTodosByIdGoldenResponse {
		fmt.Println(response.Body.String())
		t.Fatalf("Response doesn't match golden response")
	}

	// Non existing id
	response = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/1/people?ids[]=100", nil)
	m.ServeHTTP(response, req)
	if response.Code != http.StatusNotFound {
		fmt.Println(response.Body.String())
		t.Fatalf("Expected %d response code, got %d", http.StatusNotFound, response.Code)
	}
}

func TestCreateTodo(t *testing.T) {
	dbh, m := setupTest(t)
	dbh.ORM.Begin()
	defer dbh.ORM.Rollback()
	loadFixtures(dbh)

	tdate := time.Now()
	n := TodoJSON{
		Todo: todoWithPersonIdJSON{
			&db.Todo{
				Subject: "testtext",
				Date:    tdate,
			},
			1,
		},
	}
	req_body, err := json.Marshal(n)
	failOnError(t, err)

	response := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/1/todos", bytes.NewReader(req_body))
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	m.ServeHTTP(response, req)

	if response.Code != http.StatusOK {
		fmt.Println(response.Body.String())
		t.Fatalf("Expected %d response code, got %d", http.StatusOK, response.Code)
	}

	resp_todo := unmarshalTodoJSONContainer{}
	err = json.NewDecoder(response.Body).Decode(&resp_todo)
	failOnError(t, err)
	if resp_todo.Todo.Date != tdate {
		t.Fatalf("Todo Date doesn't match set date: %v != %v", resp_todo.Todo.Date,
			tdate)
	}
}

func TestUpdateTodo(t *testing.T) {
	dbh, m := setupTest(t)
	dbh.ORM.Begin()
	defer dbh.ORM.Rollback()
	loadFixtures(dbh)

	db_todo, err := dbh.GetTodoById(int64(1))
	failOnError(t, err)

	test_new_text := db_todo.Subject + "new text"
	db_todo.Subject = test_new_text

	n := TodoJSON{
		Todo: todoWithPersonIdJSON{
			db_todo,
			1,
		},
	}
	req_body, err := json.Marshal(n)
	failOnError(t, err)

	response := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/1/todos/%d", db_todo.Id), bytes.NewReader(req_body))
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	m.ServeHTTP(response, req)

	if response.Code != http.StatusOK {
		fmt.Println(response.Body.String())
		t.Fatalf("Expected %d response code, got %d", http.StatusOK, response.Code)
	}

	resp_todo := unmarshalTodoJSONContainer{}
	err = json.NewDecoder(response.Body).Decode(&resp_todo)
	if err != nil {
		t.Fatalf("Error decoding response: %v", response.Body)
	}

	if resp_todo.Todo.Subject != test_new_text {
		t.Fatalf("Todo Subject doesn't match set subject: %v != %v",
			resp_todo.Todo.Subject,
			test_new_text)
	}

	// Update a non existing Todo
	n = TodoJSON{
		Todo: todoWithPersonIdJSON{
			&db.Todo{Id: 1000},
			1,
		},
	}
	req_body, err = json.Marshal(n)
	failOnError(t, err)

	response = httptest.NewRecorder()
	req, _ = http.NewRequest("PUT", "/api/1/todos/1000", bytes.NewReader(req_body))
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	m.ServeHTTP(response, req)
	if response.Code != http.StatusNotFound {
		fmt.Println(response.Body.String())
		t.Fatalf("Expected %d response code, got %d", http.StatusNotFound, response.Code)
	}

}
