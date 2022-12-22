package sqlite3driver

import (
	"database/sql"

	"github.com/hashicorp/go-multierror"
	"github.com/peajayni/migrate"
	"golang.org/x/exp/slog"
)

func NewSqlite3Driver(db *sql.DB, logger slog.Logger) *Sqlite3Driver {
	return &Sqlite3Driver{
		db:     db,
		logger: logger,
	}
}

type Sqlite3Driver struct {
	db     *sql.DB
	logger slog.Logger
}

func (d *Sqlite3Driver) Setup() error {
	d.logger.Info("Ensuring schema_migration table exists")
	query := `
	create table if not exists schema_migration(
		name string not null primary key,
		executed_at timestamp not null default current_timestamp
	)
	`
	_, err := d.db.Exec(query)
	if err != nil {
		return err
	}
	d.logger.Info("schema_migration table exists")
	return nil
}

func (d *Sqlite3Driver) Applied() ([]string, error) {
	d.logger.Info("Getting applied migrations")
	query := `
	select name
	from schema_migration
	order by executed_at asc
	`
	rows, err := d.db.Query(query)
	if err != nil {
		return []string{}, err
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		err := rows.Scan(&name)
		if err != nil {
			return []string{}, err
		}
		names = append(names, name)
	}
	err = rows.Err()
	if err != nil {
		return []string{}, err
	}
	d.logger.Info("Got applied migrations", slog.Int("n", len(names)))

	return names, nil
}

func (d *Sqlite3Driver) Apply(migration migrate.Migration) error {
	d.logger.Info("Applying migration", slog.String("name", migration.Name))
	tx, txErr := d.db.Begin()
	if txErr != nil {
		return txErr
	}
	_, execErr := tx.Exec(migration.SQL)
	if execErr != nil {
		errRollback := tx.Rollback()
		if errRollback != nil {
			execErr = multierror.Append(execErr, errRollback)
		}
		return execErr
	}

	query := `
		insert into schema_migration(name)
		values (?)
	`
	_, execErr = tx.Exec(query, migration.Name)
	if execErr != nil {
		errRollback := tx.Rollback()
		if errRollback != nil {
			execErr = multierror.Append(execErr, errRollback)
		}
		return execErr
	}

	commitErr := tx.Commit()
	if commitErr != nil {
		return commitErr
	}
	d.logger.Info("Applied migration", slog.String("name", migration.Name))
	return nil
}
