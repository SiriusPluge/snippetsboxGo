package mysql

import (
	"database/sql"
	"errors"
	_ "github.com/go-sql-driver/mysql"
	"snip/pkg/models"
)

type Mysql struct {
	db *sql.DB
}

type Snip struct {
	id  int
	title string
}

func (s Snip) Close() {

}

func New() (*Mysql, error) {
	db, err := sql.Open("mysql", "root:12345@/snippetbox?parseTime=true")
	if err != nil {
		return nil, err
	}
	return &Mysql{db}, nil
}

func (m *Mysql) Close() error {
	return m.db.Close()
}

func (m *Mysql)GetSnip()(snip Snip, err error){
	err = m.db.QueryRow("SELECT id, title from snippets").Scan(&snip.id, &snip.title)
	if err != nil {
		return
	}

	return
}

func (m *Mysql) Insert(title string, content string, expires string) (int, error) {

	sql := `INSERT INTO snippets (title, content, created, expires)
    VALUES(?, ?, UTC_TIMESTAMP(), DATE_ADD(UTC_TIMESTAMP(), INTERVAL ? DAY))`

	result, err := m.db.Exec(sql, title, content, expires)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return  0, err
	}

	return int(id), nil
}

func (m *Mysql) Get(id int) (*models.Snippet, error) {
	stmt := `SELECT id, title, content, created, expires FROM snippets
    WHERE expires > UTC_TIMESTAMP() AND id = ?`

	row := m.db.QueryRow(stmt, id)
	s := &models.Snippet{}
	err := row.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrNoRecord
		} else {
			return nil, err
		}
	}
	return s, nil
}

func (m *Mysql) Latest() ([]*models.Snippet, error) {
	stmt := `SELECT id, title, content, created, expires FROM snippets
    WHERE expires > UTC_TIMESTAMP() ORDER BY created DESC LIMIT 10`

	rows, err := m.db.Query(stmt)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var snippets []*models.Snippet

	for rows.Next() {
		s := &models.Snippet{}

		err = rows.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
		if err != nil {
			return nil, err
		}

		snippets = append(snippets, s)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return snippets, nil
}