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
		rollbackCommand = kingpin.Command("rollback", "rollback a migration")
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
		dryRun     = kingpin.Flag("dry-run", "run migrations without committing the transaction to test for any issues.").Default("false").Bool()
		autoByPass = kingpin.Flag("auto-bypass", "bypass problematic migrations -> record them as if they were completed.").Default("false").Bool()
		rollbackId = kingpin.Flag("rollback-id", "set the rollback id to rollback migrations to").Default("").String()
		migrationTableName = kingpin.Flag("migration-table-name", "set name for migration table").Default(config.MigrationTableName).String()
		verbose    = kingpin.Flag("verbose", "make more noise").Default("false").Short('v').Bool()
	)

	var buildDatabaseConfig = func() DatabaseOptions {
		return DatabaseOptions{
			SqlUser:     *user,
			SqlPassword: *password,
			SqlHost:     *host,
			SqlPort:     *port,
			SqlDatabase: *database,
			DryRun:      *dryRun,
			AutoByPass:  *autoByPass,
			Verbose:     *verbose,
			MigrationTableName: *migrationTableName,
		}
	}

	switch kingpin.Parse() {
	case newCommand.FullCommand():
		createNewMigration()
		return
	case upCommand.FullCommand():
		tryRunMigrations(buildDatabaseConfig())
		return
	case rollbackCommand.FullCommand():
		tryRollbackMigrations(buildDatabaseConfig(), *rollbackId)
		return
	}
}

func tryRollbackMigrations(configuration DatabaseOptions, rollbackId string) {
	if rollbackId == "" {
		panic("Rollback Id must be specified")
	}
	rollbackMigrations(configuration, rollbackId)
}

func tryRunMigrations(configuration DatabaseOptions) {
	if configuration.Verbose {
		printOutMigrations()
	}
	checkForEmptyRequiredFields(configuration)
	runMigrations(configuration)
}

func printOutMigrations() {
	options := loadConfigFromJsonFile()
	var migrationFiles = findMigrationToExecute(options)
	for _, s := range migrationFiles {
		fmt.Println(s.name, s.id)
	}
}

func checkForEmptyRequiredFields(configuration DatabaseOptions) {
	if configuration.SqlUser == "" || configuration.SqlPassword == "" || configuration.SqlDatabase == "" {
		color.Red.Println("You are required to provide a sql user, password and database name, either as an argument or environment variable")
		fmt.Println(`e.g: mysql-migrator -sql-up -sql-database="doggy_db" -sql-user="doggo" -sql-password="le-woof"`)
		os.Exit(1)
	}
}
