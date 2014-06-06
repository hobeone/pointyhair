package main

import (
	"flag"

	"github.com/golang/glog"
	"github.com/hobeone/pointyhair/api"
	"github.com/hobeone/pointyhair/db"
)

func main() {
	dbh, err := db.NewDBHandle("file:test.sql", true)
	if err != nil {
		glog.Fatal(err)
	}
	flag.Set("logtostderr", "true")

	api.RunWebUi(dbh)
}
