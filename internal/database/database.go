package database

import (
	"database/sql"
	"os"

	"github.com/TheWanderingShinobi/gopher-cli-manager/pkg/models"
	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	*sql.DB
}

func NewDB() (*DB, error) {
	const file string = "database.sqlite"
	const create string = `
	CREATE TABLE IF NOT EXISTS cli (
		id INTEGER PRIMARY KEY AUTOINCREMENT,		
		name TEXT,
		description TEXT,
		path TEXT
	);
	`
	db, err := sql.Open("sqlite3", file)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(create)
	if err != nil {
		return nil, err
	}

	return &DB{db}, nil
}

func (db *DB) DeleteAllRecords() error {
	_, err := db.Exec("DELETE FROM cli")
	return err
}

func (db *DB) DeleteRecordById(id int) error {
	stmt, err := db.Prepare("DELETE FROM cli WHERE id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(id)
	return err
}

func (db *DB) UpdateCli(cli models.Cli) error {
	stmt, err := db.Prepare("UPDATE cli SET name = ?, description = ?, path = ? WHERE id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(cli.Name, cli.Description, cli.Path, cli.Id)
	return err
}

func (db *DB) CreateCli(cli models.Cli) error {
	stmt, err := db.Prepare("INSERT INTO cli(name, description, path) VALUES(?,?,?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(cli.Name, cli.Description, cli.Path)
	return err
}

func (db *DB) HasRecords() (bool, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM cli").Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (db *DB) GetEntriesContainingText(text string) ([]models.Cli, error) {
	query := "SELECT name, description, path, id FROM cli ORDER BY name ASC"
	args := []interface{}{}

	if text != "" {
		query = "SELECT name, description, path, id FROM cli WHERE name LIKE ? OR description LIKE ? ORDER BY name ASC"
		args = []interface{}{"%" + text + "%", "%" + text + "%"}
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []models.Cli
	for rows.Next() {
		var entry models.Cli
		err := rows.Scan(&entry.Name, &entry.Description, &entry.Path, &entry.Id)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, nil
}
