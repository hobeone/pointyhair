package api

import (
	"fmt"
	"runtime/debug"
	"testing"

	"github.com/golang/glog"
	"github.com/hobeone/pointyhair/db"

	"github.com/codegangsta/martini"
)

func setupTest(t *testing.T) (*db.DBHandle, *martini.Martini) {
	dbh, err := db.NewMemoryDBHandle("testing", false)
	if err != nil {
		panic("Couldn't set up the test database")
	}
	m := createMartini(dbh)
	return dbh, m
}

func failOnError(t *testing.T, err error) {
	if err != nil {
		fmt.Println(string(debug.Stack()))
		t.Fatalf("Error: %s", err.Error())
	}
}

func loadFixtures(dbh *db.DBHandle) {
	people := map[string]db.Person{
		"test1": db.Person{Name: "test1"},
		"test2": db.Person{Name: "test2"},
		"test3": db.Person{Name: "test3"},
	}
	notes := map[string]db.Note{
		"test_feed1": db.Note{Text: "http://testfeed1/feed.atom"},
		"test_feed2": db.Note{Text: "http://testfeed2/feed.atom"},
		"test_feed3": db.Note{Text: "http://testfeed3/feed.atom"},
	}
	db_people := make([]*db.Person, len(people))
	i := 0
	for _, p := range people {
		err := dbh.CreatePerson(&p)
		if err != nil {
			glog.Fatal(err.Error())
		}
		db_people[i] = &p
		i++
	}

	for _, n := range notes {
		n.Person = db_people[0]
		err := dbh.CreateNote(&n)
		if err != nil {
			glog.Fatal(err.Error())
		}
	}
}
