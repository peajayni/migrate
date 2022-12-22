package migrate

import (
	"io/fs"
	"path/filepath"
	"sort"

	"golang.org/x/exp/slog"
)

type Migration struct {
	Name string
	SQL  string
}

type Driver interface {
	Setup() error
	Applied() ([]string, error)
	Apply(migration Migration) error
}

func Migrate(driver Driver, migrationsFS fs.FS, logger slog.Logger) error {
	logger.Info("Starting migration")
	err := driver.Setup()
	if err != nil {
		return err
	}
	applied, err := driver.Applied()
	if err != nil {
		return err
	}
	newMigrations, err := loadNewMigrations(migrationsFS, applied, logger)
	if err != nil {
		return err
	}
	if len(newMigrations) == 0 {
		logger.Info("Completed migration - nothing to do")
		return nil
	}
	for _, migration := range newMigrations {
		err = driver.Apply(migration)
		if err != nil {
			return err
		}
	}
	logger.Info("Completed migration")
	return nil
}

func loadNewMigrations(migrationsFS fs.FS, applied []string, logger slog.Logger) ([]Migration, error) {
	appliedSet := make(map[string]struct{})
	for _, a := range applied {
		appliedSet[a] = struct{}{}
	}

	paths, err := fs.Glob(migrationsFS, "*.sql")
	if err != nil {
		return []Migration{}, err
	}

	sort.Strings(paths)

	logger.Info("Found migrations", slog.Int("n", len(paths)))

	var migrations []Migration
	for _, path := range paths {
		name := filepath.Base(path)
		_, exists := appliedSet[name]
		if exists {
			continue
		}
		sql, err := fs.ReadFile(migrationsFS, path)
		if err != nil {
			return []Migration{}, err
		}
		migration := Migration{
			Name: name,
			SQL:  string(sql),
		}
		migrations = append(migrations, migration)
	}

	logger.Info("Migrations to apply", slog.Int("n", len(migrations)))
	return migrations, nil
}
