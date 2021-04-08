package main

import (
	"flag"
	"fmt"
	"github.com/gookit/color"
	"os"
	"time"
)

func main() {
	sqlNew := flag.Bool("sql-new", false, "flag that set's whether a new sql migration needs to be created.")
	sqlUp := flag.Bool("sql-up", false, "flag that is set to define whether existing migrations should be run.")
	sqlUser := flag.String("sql-user", os.Getenv("SQL_USER"), "the sql user that needs to be used to execute migrations")
	sqlPassword := flag.String("sql-password", os.Getenv("SQL_PASSWORD"), "the sql user password that is required to execute the migrations")
	sqlPort := flag.String("sql-port", os.Getenv("SQL_PORT"), "the sql port that is required to open a db connection")
	sqlHost := flag.String("sql-host", os.Getenv("SQL_HOST"), "the sql host that is required to open a db connection")
	sqlDatabase := flag.String("sql-database", os.Getenv("SQL_DATABASE"), "the targeted database that is required to open a db connection")
	envDryRun := false
	if os.Getenv("DRY_RUN") != "" {
		envDryRun = true
	}

	dryRun := flag.Bool("dry-run", envDryRun, "use the dry-run flag if you want to execute all migrations within a transaction scope so that changes are not persisted")
	envAutoByPass := false

	if os.Getenv("AUTO_BYPASS") != "" {
		envAutoByPass = true
	}

	autoByPass := flag.Bool("auto-bypass", envAutoByPass, "use this flag if you want to continue executing migrations even if one or more fail to execute")

	flag.Parse()

	if *sqlPort == "" {
		*sqlPort = "3306"
	}
	if *sqlHost == "" {
		*sqlHost = "127.0.0.1"
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

type Schema struct {
	id int64
	name string
	dateexecuted time.Time
}

