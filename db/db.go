package db

import (
	"fmt"
	"sync"
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/davecgh/go-spew/spew"
	"github.com/golang/glog"
	_ "github.com/mattn/go-sqlite3"
)

type Todo struct {
	Id       int64     `json:"id"`
	Subject  string    `json:"subject"`
	Date     time.Time `json:"date"`
	Person   *Person   `orm:"rel(fk)"  json:"-"`
	Category string    `json:"category"`
	Notes    string    `orm:"type(text)"`
}

type RecurringTodo struct {
	Id         int `json:"id"`
	Recurrance string
	Subject    string
}

type DBHandle struct {
	ORM       orm.Ormer
	syncMutex sync.Mutex
}

func NewDBHandle(db_path string, verbose bool) (*DBHandle, error) {
	d := &DBHandle{}
	err, o := createAndOpenDB(db_path, verbose, false)
	if err != nil {
		return nil, err
	}
	d.ORM = o
	return d, nil
}

func NewMemoryDBHandle(db_path string, verbose bool) (*DBHandle, error) {
	d := &DBHandle{}
	err, o := createAndOpenDB(db_path, verbose, true)
	if err != nil {
		return nil, err
	}
	d.ORM = o
	return d, nil
}

func createAndOpenDB(db_path string, verbose bool, memory bool) (error, orm.Ormer) {

	mode := "rwc"
	if memory {
		mode = "memory"
	}
	db_path_ext := fmt.Sprintf("file:%s?mode=%s", db_path, mode)

	orm.RegisterDataBase("default", "sqlite3", db_path_ext)
	orm.Debug = verbose

	err := orm.RunSyncdb("default", false, verbose)
	if err != nil {
		return err, nil
	}

	return nil, orm.NewOrm()
}

func init() {
	orm.RegisterModel(new(Person))
	orm.RegisterModel(new(Note))
	orm.RegisterModel(new(Todo))
	orm.RegisterModel(new(RecurringTodo))
}

func Demo() {
	dbh, err := NewMemoryDBHandle("testing", true)
	if err != nil {
		glog.Fatalf("Error: %s", err)
	}

	dbh.ORM.Begin()
	p1 := Person{
		Name: "apw",
	}

	created, id, err := dbh.ORM.ReadOrCreate(&p1, "Name")
	if err != nil {
		glog.Fatal(err)
	}

	if created {
		fmt.Println("New Insert an object. Id:", id)
	} else {
		fmt.Println("Get an object. Id:", id)
	}
	spew.Dump(p1)

	n1 := Note{
		Person:   &p1,
		Text:     "testing\nfoo",
		Category: "test",
	}

	_, err = dbh.ORM.Insert(&n1)
	if err != nil {
		glog.Fatal(err)
	}

	spew.Dump(n1)
	dbh.ORM.Commit()

	fmt.Println("**********************************")

	p := Person{Id: p1.Id}
	err = dbh.ORM.Read(&p)
	if err != nil {
		glog.Fatal(err)
	}
	dbh.ORM.LoadRelated(&p, "Notes")

	spew.Dump(p.Notes)
	for _, n := range p.Notes {
		fmt.Printf("Note: %s - %s", p.Name, n.Text)
	}
}
