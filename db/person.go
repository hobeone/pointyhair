package db

type Person struct {
	Id    int64   `json:"id"`
	Name  string  `orm:"size(255);unique" json:"name"`
	Notes []*Note `orm:"reverse(many)" json:"-"`
	Todos []*Todo `orm:"reverse(many)" json:"-"`
}

func (p *Person) LoadRelated(dbh *DBHandle) error {
	if _, err := dbh.ORM.LoadRelated(p, "Notes"); err != nil {
		return err
	}

	if _, err := dbh.ORM.LoadRelated(p, "Todos"); err != nil {
		return err
	}

	return nil
}

// Returns all people if ids arguement is empty
func (dbh *DBHandle) GetPeopleById(ids []int64) ([]*Person, error) {
	var p []*Person
	_, err := dbh.ORM.QueryTable("person").All(&p)
	return p, err
}

func (dbh *DBHandle) GetPersonById(id int64) (*Person, error) {
	p := Person{Id: id}
	err := dbh.ORM.Read(&p)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (dbh *DBHandle) CreatePerson(p *Person) error {
	_, err := dbh.ORM.Insert(p)
	if err != nil {
		return err
	}
	return nil
}

func (dbh *DBHandle) UpdatePerson(*Person) error { return nil }
