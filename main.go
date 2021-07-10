package main

import (
	"flag"
	"fmt"
	"github.com/alecthomas/kingpin"
	"github.com/gookit/color"
	"os"
)

func main() {
	var (
		newCommand = kingpin.Command("sql-new", "creates a new sql migration script")
		upCommand = kingpin.Command("sql-up", "run migrate up")
	)
	switch kingpin.Parse() {
	case newCommand.FullCommand():
		createNewMigration()
		return
	case upCommand.FullCommand():
		tryRunMigrations()
		return
	}

	Initialise()
}

func tryRunMigrations() {
	config := loadConfigFromJsonFile()
	var (
		user = kingpin.Flag("user",
			"username required to connect to mysql").Default(config.sqlUser).Short('u').Required().String()
		password = kingpin.Flag("password",
			"password required to connect to mysql").Default(config.sqlPassword).Short('p').Required().String()
		port = kingpin.Flag("port", "port number mysql is active on").Default(config.sqlPort).String()
		database = kingpin.Flag("database", "database to connect to").Default(config.sqlDatabase).String()
		host = kingpin.Flag("host", "host that mysql is running on").Default(config.sqlHost).String()
		dryRun = kingpin.Flag("dry-run", "runs the migrations in 'ISOLATION UNCOMMITTED' mode to test them").Bool()
		autoByPass = kingpin.Flag("auto-bypass", "bypass problematic migrations -> record them as if they were completed.").Bool()
		verbose = kingpin.Flag("verbose", "make more noise").Default("false").Short('v').Bool()
	)
	if *verbose {
		printOutMigrations()
	}
	if requiredFieldsAreEmpty(*user, *password, *database) {
		color.Red.Println("You are required to provide a sql user, password and database name, either as an argument or environment variable")
		fmt.Println(`e.g: mysql-migrator -sql-up -sql-database="doggy_db" -sql-user="doggo" -sql-password="le-woof"`)
		os.Exit(1)
	}
	runMigrations(DatabaseOptions{
		sqlUser:     *user,
		sqlPassword: *password,
		sqlHost:     *host,
		sqlPort:     *port,
		sqlDatabase: *database,
		dryRun:      *dryRun,
		autoByPass:  *autoByPass,
	})
}

func printOutMigrations() {
	options := loadConfigFromJsonFile()
	var migrationFiles = findMigrationToExecute(options)
	for _, s := range migrationFiles {
		fmt.Println(s.name, s.id)
	}
}

func requiredFieldsAreEmpty(sqlUser, sqlPassword, sqlDatabase string) bool {
	if sqlUser == "" || sqlPassword == "" || sqlDatabase == "" {
		return true
	}
	return false
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

	config := loadConfigFromJsonFile()

	if requiredFieldsAreEmpty(*sqlUser, *sqlPassword, *sqlDatabase){
		*sqlUser = config.sqlUser
		*sqlPassword = config.sqlPassword
		*sqlDatabase = config.sqlDatabase
		*sqlHost = config.sqlHost
		*sqlPort = config.sqlHost
	}

	if config.autoByPass || config.dryRun {
		*dryRun = config.dryRun
		*autoByPass = config.autoByPass
	}
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
		if requiredFieldsAreEmpty(*sqlUser, *sqlPassword, *sqlDatabase) {
			color.Red.Println("You are required to provide a sql user, password and database name, either as a argument or environment variable")
			fmt.Println(`e.g: mysql-migrator -sql-up -sql-database="doggy_db" -sql-user="doggo" -sql-password="le-woof"`)
			os.Exit(1)
		}

		runMigrations(DatabaseOptions{
			sqlUser:     *sqlUser,
			sqlPassword: *sqlPassword,
			sqlHost:     *sqlHost,
			sqlPort:     *sqlPort,
			sqlDatabase: *sqlDatabase,
			dryRun:      *dryRun,
			autoByPass:  *autoByPass,
		})

		os.Exit(0)
	}
}

