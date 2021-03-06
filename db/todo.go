package db

import "time"

// Proably a better way of dealing with this.
type Todo struct {
	Id       int64     `json:"id"`
	Date     time.Time `json:"date"`
	Person   *Person   `orm:"rel(fk)"  json:"-"`
	Text     string    `orm:"type(text)" json:"text"`
	Category string    `json:"category"`
}

func (dbh *DBHandle) GetTodoById(id int64) (*Todo, error) {
	t := Todo{Id: id}
	err := dbh.ORM.Read(&t)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (dbh *DBHandle) GetTodos() ([]*Todo, error) {
	var todos []*Todo
	_, err := dbh.ORM.QueryTable("todo").All(&todos)
	return todos, err
}

func (dbh *DBHandle) GetTodosByIds(ids []int64) ([]*Todo, error) {
	var todos []*Todo
	_, err := dbh.ORM.QueryTable("todo").Filter("id__in", ids).All(&todos)
	return todos, err
}

func (dbh *DBHandle) CreateTodo(t *Todo) error {
	if _, err := dbh.ORM.Insert(t); err != nil {
		return err
	}
	return nil
}

func (dbh *DBHandle) UpdateTodo(t *Todo) error {
	if _, err := dbh.ORM.Update(&t); err != nil {
		return err
	}
	return nil
}

func (dbh *DBHandle) RemoveTodo(t *Todo) error {
	if _, err := dbh.ORM.Delete(t); err != nil {
		return err
	}
	return nil
}

func (dbh *DBHandle) AddTodoToAllPeople(t *Todo) error {
	var people []*Person
	_, err := dbh.ORM.QueryTable("person").All(&people)
	if err != nil {
		return err
	}
	for _, p := range people {
		t.Person = p
		err = dbh.CreateTodo(t)
		if err != nil {
			return err
		}
	}
	return nil
}
