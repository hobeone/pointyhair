package db

import "time"

type Note struct {
	Id       int64     `json:"id"`
	Date     time.Time `json:"date"`
	Person   *Person   `orm:"rel(fk)"  json:"-"`
	Text     string    `orm:"type(text)" json:"text"`
	Category string    `json:"category"`
}

// Returns all people if ids arguement is empty
func (dbh *DBHandle) GetNotesById(ids []int64) ([]*Note, error) {
	var p []*Note
	_, err := dbh.ORM.QueryTable("note").All(&p)
	return p, err
}

func (dbh *DBHandle) GetNoteById(id int64) (*Note, error) {
	p := Note{Id: id}
	err := dbh.ORM.Read(&p)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (dbh *DBHandle) CreateNote(p *Note) error {
	_, err := dbh.ORM.Insert(p)
	if err != nil {
		return err
	}
	return nil
}

func (dbh *DBHandle) UpdateNote(note *Note) error {
	if _, err := dbh.ORM.Update(note); err != nil {
		return err
	}
	return nil
}

func (dbh *DBHandle) RemoveNote(note *Note) error {
	if _, err := dbh.ORM.Delete(note); err != nil {
		return err
	}
	return nil
}
