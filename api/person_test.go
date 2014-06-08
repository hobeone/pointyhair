package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

const getPersonByIdGoldenResponse = `{
  "person": {
    "id": 3,
    "name": "test3",
    "notes": [
      1,
      2,
      3
    ]
  }
}`

func TestGetPersonById(t *testing.T) {
	dbh, m := setupTest(t)

	dbh.ORM.Begin()
	defer dbh.ORM.Rollback()
	loadFixtures(dbh)

	response := httptest.NewRecorder()

	req, _ := http.NewRequest("GET", "/api/1/people/3", nil)
	m.ServeHTTP(response, req)

	if response.Code != 200 {
		fmt.Println(response.Body.String())
		t.Fatalf("Expected 200 response code, got %d", response.Code)
	}

	if response.Body.String() != getPersonByIdGoldenResponse {
		fmt.Println(response.Body.String())
		t.Fatalf("Response doesn't match golden response")
	}

	// Non existing id
	response = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/1/people/100", nil)
	m.ServeHTTP(response, req)
	if response.Code != http.StatusNotFound {
		fmt.Println(response.Body.String())
		t.Fatalf("Expected %d response code, got %d", http.StatusNotFound, response.Code)
	}
}

const getPeopleByIdGoldenResponse = `{
  "people": [
    {
      "id": 2,
      "name": "test2",
      "notes": []
    },
    {
      "id": 3,
      "name": "test3",
      "notes": [
        1,
        2,
        3
      ]
    }
  ]
}`

func TestGetPeopleWithIds(t *testing.T) {
	dbh, m := setupTest(t)

	dbh.ORM.Begin()
	defer dbh.ORM.Rollback()
	loadFixtures(dbh)

	response := httptest.NewRecorder()

	req, _ := http.NewRequest("GET", "/api/1/people?ids[]=2&ids[]=3", nil)
	m.ServeHTTP(response, req)
	if response.Code != 200 {
		fmt.Println(response.Body.String())
		t.Fatalf("Expected 200 response code, got %d", response.Code)
	}

	if response.Body.String() != getPeopleByIdGoldenResponse {
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
