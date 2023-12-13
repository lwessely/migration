# README
> github.com/lwessely/migration is a library that manages database migrations for projects in golang.
> Currently, mysql/mariadb are supported.

## Usage
Here's an example of how to create a few migrations, roll them out, and roll them back.

```go
import (
  "fmt"
  "database/sql"

  "github.com/go-sql-driver/mysql"
  "github.com/lwessely/migration"
)

// Test migrations

tm1 := Migration{
	Name: "migration-1",
	UpQuery: `
		CREATE TABLE testTable1 (
			id INT KEY AUTO_INCREMENT,
			testColumn1 VARCHAR(255)
		)
	`,
	DownQuery: "DROP TABLE testTable1",
}
tm2 := Migration{
	Name: "migration-2",
	UpQuery: `
		CREATE TABLE testTable2 (
			id INT KEY AUTO_INCREMENT,
			testColumn2 VARCHAR(255)
		)
	`,
	DownQuery: "DROP TABLE testTable2",
}
tm3 := Migration{
	Name: "migration-3",
	UpQuery: `
		CREATE TABLE testTable3 (
			id INT KEY AUTO_INCREMENT,
			testColumn2 VARCHAR(255)
		)
	`,
	DownQuery: "DROP TABLE testTable3",
}
tm4 := Migration{
	Name: "migration-4",
	UpQuery: `
		CREATE TABLE testTable4 (
			id INT KEY AUTO_INCREMENT,
			testColumn2 VARCHAR(255)
		)
	`,
	DownQuery: "DROP TABLE testTable4",
}

// Connect to database

var database *sql.DB

cfg := mysql.Config{
  User:                 "test",
  Passwd:               "test",
  Net:                  "tcp",
  Addr:                 "127.0.0.1:3306",
  DBName:               "test",
  AllowNativePasswords: true,
}

database, err := sql.Open("mysql", cfg.FormatDSN())

if nil != err {
  fmt.Println(err)
}

// Create first migration plan

mp1 := MigrationPlan{}
mp1.Add(tm1).Add(tm2)

// Create second migration plan

mp2 := MigrationPlan{}
mp2.Add(tm3).Add(tm4)

// Create third migration plan, and merge the first two plans into it

mp := MigrationPlan{}
mp.Concat(&mp1, &mp2)

// Roll out first migration

count, err := mp.Up(database)

if nil != err {
  fmt.Println(err)
}

fmt.Printf("Performed %d migration(s).\n", count)

// Roll out rest of migrations

count, err = mp.Latest(database)

if nil != err {
  fmt.Println(err)
}

fmt.Printf("Performed %d migration(s).\n", count)

// Roll back last migration

count, err = mp.Down(database)

if nil != err {
  fmt.Println(err)
}

fmt.Printf("Rolled back %d migration(s).\n", count)

// Roll back rest of migrations

count, err = mp.Reset(database)

if nil != err {
  fmt.Println(err)
}

fmt.Printf("Rolled back %d migration(s).\n", count)

// Clean up

err = database.Close()

if nil != err {
  fmt.Println(err)
}
```