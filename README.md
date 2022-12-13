# migrate
Golang Database Migration Library

## Quickstart
```
textHandler := slog.NewTextHandler(os.Stdout)
logger := slog.New(textHandler)
driver := sqlite3driver.NewSqlite3Driver(db, logger)
err = migrate.Migrate(driver, migrationsFS, logger)
```

## How it works

The `Migrate` function queries the driver for a list of all applied migrations. It then loads all the SQL files found on the supplied file system.
It will sort these files by filename and then ask the driver to apply each migration that was not included in the original list of applied migrations.

## Why not use `golang-migrate/migrate`?

I needed a simple library to perform Sqlite3 migrations. When you install `golang-migrate/migrate` it pulls in the dependencies
for **all** the drivers it supports. This was leading to long build times.
