package migration

import (
	"database/sql"
	"fmt"
)

// MigrationError implements the error interface for this module.
type MigrationError struct {
	Message string
}

// Error returns the error message from MigrationError.
func (err *MigrationError) Error() string {
	return err.Message
}

// Migration represents a database migration.
type Migration struct {
	Name      string
	UpQuery   string
	DownQuery string
	Previous  *Migration
	Next      *Migration
}

// MigrationPlan is a linked list of database migrations that are intended to be executed
// in order.
type MigrationPlan struct {
	First *Migration
	Last  *Migration
}

// Add appends a copy of the migration provided to the mp. A pointer to mp is returned, so
// calls to Add can be chained.
func (mp *MigrationPlan) Add(migration Migration) (result *MigrationPlan) {
	migration.Next = nil
	migration.Previous = nil

	if nil == mp.First || nil == mp.Last {
		mp.First = &migration
		mp.Last = &migration

		return mp
	}

	migration.Previous = mp.Last
	mp.Last.Next = &migration
	mp.Last = &migration

	return mp
}

// Concat copies all migrations from migrationPlans and appends them to the end of mp. A pointer
// to mp is returned, so calls to Concat can be chained.
func (mp *MigrationPlan) Concat(migrationPlans ...*MigrationPlan) (result *MigrationPlan) {
	for _, plan := range migrationPlans {
		migration := plan.First

		for nil != migration {
			mp.Add(*migration)
			migration = migration.Next
		}
	}

	return mp
}

// ensureMigrationsTable creates the migration tracking table if it does not exist.
func (mp *MigrationPlan) ensureMigrationsTable(database *sql.DB) error {
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS migrations (
			id INT KEY AUTO_INCREMENT,
			migrationName VARCHAR(255) UNIQUE NOT NULL
		)
	`
	_, err := database.Exec(createTableQuery)

	return err
}

// getCurrentMigration returns a pointer to the struct representing the last migration that has been performed
// according to the migration tracking table. If no migrations have been performed yet, nil is returned.
func (mp *MigrationPlan) getCurrentMigration(database *sql.DB) (currentMigration *Migration, err error) {
	if err = mp.ensureMigrationsTable(database); err != nil {
		return nil, err
	}

	currentMigrationNameQuery := `
		SELECT
			migrationName
		FROM
			migrations
	`
	rows, err := database.Query(currentMigrationNameQuery)
	defer func() {
		rows.Close()
	}()

	if nil != err {
		return nil, err
	}

	migrationToCheck := mp.First

	for rows.Next() {
		var migrationName string
		err = rows.Scan(&migrationName)

		if migrationToCheck == nil {
			return nil, &MigrationError{
				Message: fmt.Sprintf(
					"Mismatch between migration plan and performed migrations: '%s' is not in migration plan.",
					migrationName,
				),
			}
		}

		if nil != err {
			return nil, err
		}

		if migrationName != migrationToCheck.Name {
			return nil, &MigrationError{
				Message: fmt.Sprintf(
					"Mismatch between migration plan and performed migrations: '%s' != '%s'.",
					migrationName,
					currentMigration.Name,
				),
			}
		}

		currentMigration = migrationToCheck
		migrationToCheck = migrationToCheck.Next
	}

	return currentMigration, nil
}

// Up performs the first migration in the migration plan that has not yet been performed.
// If the migration was performed successfully, migrationsPerformed is 1, and err is nil.
// If there are no more migrations to perform, migrationsPerformed is 0 and err is nil.
// If an error occurred, error is non-nil.
func (mp *MigrationPlan) Up(database *sql.DB) (migrationsPerformed uint, err error) {
	currentMigration, err := mp.getCurrentMigration(database)

	if nil != err {
		return 0, err
	}

	var nextMigration *Migration

	if nil == currentMigration {
		nextMigration = mp.First
	} else {
		nextMigration = currentMigration.Next
	}

	if nil == nextMigration {
		return 0, nil
	}

	trx, err := database.Begin()

	if nil != err {
		return 0, err
	}

	if _, err = trx.Exec(nextMigration.UpQuery); nil != err {
		if rollbackErr := trx.Rollback(); nil != rollbackErr {
			return 0, rollbackErr
		}

		return 0, err
	}

	saveMigrationQuery := `
		INSERT INTO migrations (
			migrationName
		) VALUES (
			?
		)
	`
	if _, err = trx.Exec(saveMigrationQuery, nextMigration.Name); nil != err {
		if rollbackErr := trx.Rollback(); nil != rollbackErr {
			return 0, rollbackErr
		}

		return 0, err
	}

	if err = trx.Commit(); nil != err {
		return 0, err
	}

	return 1, nil
}

// Down rolls back the last migration that has been performed in the migration plan.
// If the migration was performed successfully, migrationsPerformed is 1, and err is nil.
// If there are no more migrations to roll back, migrationsPerformed is 0, and err is nil.
// If an error occurred, err is non-nil.
func (mp *MigrationPlan) Down(database *sql.DB) (migrationsPerformed uint, err error) {
	currentMigration, err := mp.getCurrentMigration(database)

	if nil != err {
		return 0, err
	}

	if nil == currentMigration {
		return 0, nil
	}

	trx, err := database.Begin()

	if _, err = trx.Exec(currentMigration.DownQuery); nil != err {
		if rollbackErr := trx.Rollback(); nil != rollbackErr {
			return 0, rollbackErr
		}

		return 0, err
	}

	saveMigrationQuery := `
		DELETE FROM migrations
		WHERE migrationName = ?
	`
	if _, err = trx.Exec(saveMigrationQuery, currentMigration.Name); nil != err {
		if rollbackErr := trx.Rollback(); nil != rollbackErr {
			return 0, rollbackErr
		}

		return 0, err
	}

	if err = trx.Commit(); nil != err {
		return 0, err
	}

	return 1, nil
}

// Latest performs all migrations in the migration plan that have not yet been performed.
// migrationsPerformed is the number of migrations that were performed successfully. If all
// migrations were performed successfully, err is nil. If an error occurred, no further
// migrations are performed, and err is non-nil. Successful migrations will not be rolled
// back on error.
func (mp *MigrationPlan) Latest(database *sql.DB) (migrationsPerformed uint, err error) {
	for true {
		var migrationCount uint
		migrationCount, err = mp.Up(database)

		if nil != err || migrationCount == 0 {
			break
		}

		migrationsPerformed += migrationCount
	}

	return migrationsPerformed, err
}

// Reset rolls back all migrations in the migration plan. migrationsPerformed is the number
// of migrations that were successfully rolled back. If all migrations were rolled back
// successfully, err is nil. If an error occurred, no further migrations are rolled back,
// and err is non-nil.
func (mp *MigrationPlan) Reset(database *sql.DB) (migrationsPerformed uint, err error) {
	for true {
		var migrationCount uint
		migrationCount, err = mp.Down(database)

		if nil != err || migrationCount == 0 {
			break
		}

		migrationsPerformed += migrationCount
	}

	return migrationsPerformed, err
}
