package main

import (
	"flag"
	"fmt"
	"github.com/gookit/color"
	"os"
)

func main() {
	//Initialise()
	createNewMigration()
	//printOutMigrationsForDb()
}

func printOutMigrations() {
	options := loadConfigFromJsonFile()
	var migrationFiles = findMigrationToExecute(options)
	for _, s := range migrationFiles {
		fmt.Println(s.name, s.id)
	}
}

func Initialise() {
	sqlNew := flag.Bool("sql-new", false, "flag that set's whether a new sql migration needs to be created.")
	sqlUp := flag.Bool("sql-up", false, "flag that is set to define whether existing migrations should be run.")
	sqlUser := flag.String("sql-user", "", "the sql user that needs to be used to execute migrations")
	sqlPassword := flag.String("sql-password", "", "the sql user password that is required to execute the migrations")
	sqlPort := flag.String("sql-port", "", "the sql port that is required to open a db connection")
	sqlHost := flag.String("sql-host", "", "the sql host that is required to open a db connection")
	sqlDatabase := flag.String("sql-database", "", "the targeted database that is required to open a db connection")
	dryRun := flag.Bool("dry-run", false, "dry run will run the migrations in 'ISOLATION UNCOMMITTED' mode")
	autoByPass := flag.Bool("auto-bypass", false, "if auto bypass in enabled, a failed migration would throw the error and be inserted into the db.")
	flag.Parse()

	if *sqlPort == "" {
		*sqlPort = "3306"
	}
	if *sqlHost == "" {
		*sqlHost = "localhost"
	}
	if *sqlNew == false && *sqlUp == false {
		color.Red.Println("You didn't supply any arguments... Please try again, use -h for help.")
		os.Exit(1)
	}
	if *sqlNew == true && *sqlUp == true {
		color.Cyan.Println("You can not run sql-new and sql-up at the same time, only sql-new will be run...")
		createNewMigration()
	}
	if *sqlNew == true {
		createNewMigration()
		os.Exit(0)
	}
	if *sqlUp == true {
		if *sqlUser == "" ||
			*sqlPassword == "" ||
			*sqlDatabase == "" {

			color.Red.Println("You are required to provide a sql user, password and database name, either as a argument or environment variable")
			fmt.Println(`e.g: mysql-migrator -sql-up -sql-database="doggy_db" -sql-user="doggo" -sql-password="le-woof"`)
			os.Exit(1)
		}

		runMigrations(
			*sqlUser,
			*sqlPassword,
			*sqlHost,
			*sqlPort,
			*sqlDatabase,
			*dryRun,
			*autoByPass)

		os.Exit(0)
	}
}

