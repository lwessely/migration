package migration

import "fmt"

type QueryResult interface {
	LastInsertId() (int64, error)
	RowsAffected() (int64, error)
}

type QueryRows interface {
	Next() bool
	Close() error
	Scan(dest ...any) error
}

type DatabaseTransaction interface {
	Exec(query string, args ...any) (QueryResult, error)
	Rollback() error
	Commit() error
}

type DatabaseConnection interface {
	Exec(query string, args ...any) (QueryResult, error)
	Query(query string, args ...any) (QueryRows, error)
	Begin() (DatabaseTransaction, error)
}

type MigrationError struct {
	Message string
}

func (err *MigrationError) Error() string {
	return err.Message
}

type Migration struct {
	Name      string
	UpQuery   string
	DownQuery string
	Previous  *Migration
	Next      *Migration
}

type MigrationPlan struct {
	First *Migration
	Last  *Migration
}

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

func (mp *MigrationPlan) ensureMigrationsTable(database DatabaseConnection) error {
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS migrations (
			id INT AUTO_INCREMENT,
			migrationName VARCHAR(255) UNIQUE NOT NULL
		)
	`
	_, err := database.Exec(createTableQuery)

	return err
}

func (mp *MigrationPlan) getCurrentMigration(database DatabaseConnection) (currentMigration *Migration, err error) {
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
		err = rows.Scan(migrationName)

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

func (mp *MigrationPlan) Up(database DatabaseConnection) (migrationsPerformed uint, err error) {
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

func (mp *MigrationPlan) Down(database DatabaseConnection) (migrationsPerformed uint, err error) {
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

func (mp *MigrationPlan) Latest(database DatabaseConnection) (migrationsPerformed uint, err error) {
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

func (mp *MigrationPlan) Reset(database DatabaseConnection) (migrationsPerformed uint, err error) {
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
