package main

import (
	"fmt"
	"github.com/alecthomas/kingpin"
	"github.com/gookit/color"
	"os"
)

func main() {
	initialiseParameterOptions()
}

func initialiseParameterOptions() {
	// Commands
	var (
		newCommand = kingpin.Command("sql-new", "creates a new sql migration script")
		upCommand  = kingpin.Command("sql-up", "run migrate up")
	)
	// Flags
	config := loadConfigFromJsonFile()
	var (
		user = kingpin.Flag("user",
			"username required to connect to mysql").Default(config.SqlUser).Short('u').String()
		password = kingpin.Flag("password",
			"password required to connect to mysql").Default(config.SqlPassword).Short('p').String()
		port       = kingpin.Flag("port", "port number mysql is active on").Default(config.SqlPort).String()
		database   = kingpin.Flag("database", "database to connect to").Default(config.SqlDatabase).String()
		host       = kingpin.Flag("host", "host that mysql is running on").Default(config.SqlHost).String()
		dryRun     = kingpin.Flag("dry-run", "runs the migrations in 'ISOLATION UNCOMMITTED' mode to test them").Default("false").Bool()
		autoByPass = kingpin.Flag("auto-bypass", "bypass problematic migrations -> record them as if they were completed.").Default("false").Bool()
		verbose    = kingpin.Flag("verbose", "make more noise").Default("false").Short('v').Bool()
	)

	switch kingpin.Parse() {
	case newCommand.FullCommand():
		createNewMigration()
		return
	case upCommand.FullCommand():
		tryRunMigrations(DatabaseOptions{
			SqlUser:     *user,
			SqlPassword: *password,
			SqlHost:     *host,
			SqlPort:     *port,
			SqlDatabase: *database,
			DryRun:      *dryRun,
			AutoByPass:  *autoByPass,
			Verbose:     *verbose,
		})
		return
	}
}

func tryRunMigrations(configuration DatabaseOptions) {
	if configuration.Verbose {
		printOutMigrations()
	}
	if requiredFieldsAreEmpty(
		configuration.SqlUser,
		configuration.SqlPassword,
		configuration.SqlDatabase) {
		color.Red.Println("You are required to provide a sql user, password and database name, either as an argument or environment variable")
		fmt.Println(`e.g: mysql-migrator -sql-up -sql-database="doggy_db" -sql-user="doggo" -sql-password="le-woof"`)
		os.Exit(1)
	}
	runMigrations(configuration)
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

