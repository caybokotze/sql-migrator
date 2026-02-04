package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/alecthomas/kingpin"
	"github.com/gookit/color"
	"github.com/inconshreveable/go-update"
)

const version = "1.2.0"

func main() {
	initialiseParameterOptions()
}

func initialiseParameterOptions() {
	// Commands
	var (
		newCommand      = kingpin.Command("sql-new", "creates a new sql migration script")
		upCommand       = kingpin.Command("sql-up", "run pending migrations")
		rollbackCommand = kingpin.Command("rollback", "rollback migrations to a specific migration ID (inclusive)")
		versionCommand  = kingpin.Command("version", "display the version of the migration tool")
		updateCommand   = kingpin.Command("update", "auto-update the migration tool from GitHub releases")
	)

	// Rollback command argument
	var rollbackTargetId = rollbackCommand.Arg("migration-id", "the migration ID (timestamp) to rollback to (inclusive)").Required().String()

	// Flags
	var fileName = kingpin.Flag("configuration-file", "set configuration file name").Default("migrator-config.json").String()
	var scriptDirectory = kingpin.Flag("migration-directory", "set the directory where the migration scripts can be found").Default("./scripts/").String()

	var (
		user = kingpin.Flag("user",
			"username required to connect to mysql").Default("sqltracking").Short('u').String()
		password = kingpin.Flag("password",
			"password required to connect to mysql. Passwords can also be provided as environment variables. Use `ENV:NameOfVariable` formatting...").Default("").Short('p').String()
		port               = kingpin.Flag("port", "port number mysql is active on").Default("3306").String()
		database           = kingpin.Flag("database", "database to connect to").Default("").String()
		host               = kingpin.Flag("host", "host that mysql is running on").Default("localhost").String()
		dryRun             = kingpin.Flag("dry-run", "run migrations without committing the transaction to tests for any issues.").Default("false").Bool()
		autoByPass         = kingpin.Flag("auto-bypass", "bypass problematic migrations -> record them as if they were completed.").Default("false").Bool()
		migrationTableName = kingpin.Flag("migration-table-name", "set name for migration table").Default("__migrations").String()
		verbose            = kingpin.Flag("verbose", "make more noise").Default("false").Bool()
	)

	var buildDatabaseConfig = func() DatabaseOptions {
		config := loadConfigFromJsonFile(*fileName)

		if *user == "sqltracking" {
			if config.SqlUser != "" {
				user = &config.SqlUser
			}
		}
		if *password == "" {
			if config.SqlPassword != "" {
				password = &config.SqlPassword
			}
		}
		if *password != "" {
			if strings.HasPrefix(*password, "ENV:") {
				envVariableValue := os.Getenv(strings.SplitN(*password, ":", 2)[1])
				password = &envVariableValue
				visible := password

				if len(*password) > 5 {
					hiddenPart := strings.Repeat("*", len(*password)-5)
					tempVisible := fmt.Sprintf("%s%s", (*password)[:5], hiddenPart)
					visible = &tempVisible
				}

				fmt.Println("Using environment variable password with value: " + *visible)
			}
		}
		if *host == "localhost" {
			if config.SqlHost != "" {
				host = &config.SqlHost
			}
		}
		if *port == "3306" {
			if config.SqlPort != "" {
				port = &config.SqlPort
			}
		}
		if *database == "" {
			if config.SqlDatabase != "" {
				database = &config.SqlDatabase
			}
		}

		if *migrationTableName == "__migrations" {
			if config.MigrationTableName != "" {
				migrationTableName = &config.MigrationTableName
			}
		}

		return DatabaseOptions{
			SqlUser:            *user,
			SqlPassword:        *password,
			SqlHost:            *host,
			SqlPort:            *port,
			SqlDatabase:        *database,
			DryRun:             *dryRun,
			AutoByPass:         *autoByPass,
			Verbose:            *verbose,
			MigrationTableName: *migrationTableName,
			ConfigFileName:     *fileName,
			MigrationDirectory: *scriptDirectory,
		}
	}

	switch kingpin.Parse() {
	case newCommand.FullCommand():
		createNewMigration(*scriptDirectory)
		return
	case upCommand.FullCommand():
		tryRunMigrations(buildDatabaseConfig())
		return
	case rollbackCommand.FullCommand():
		tryRollbackMigrations(buildDatabaseConfig(), *rollbackTargetId)
		return
	case versionCommand.FullCommand():
		fmt.Printf("sql-migrator version %s\n", version)
		return
	case updateCommand.FullCommand():
		err := doUpdate()
		if err != nil {
			log.Fatal("Could not update the application")
		}
		return
	}
}

func doUpdate() error {
	resp, err := http.Get("https://github.com/caybokotze/sql-migrator/releases/latest/download/sql-migrator.exe")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	color.Blue.Println("Downloading update...")
	err = update.Apply(resp.Body, update.Options{})
	if err == nil {
		color.Green.Println("Updated sql migrator successfully")
	}
	if err != nil {
		return err
	}
	return err
}

func tryRollbackMigrations(configuration DatabaseOptions, rollbackId string) {
	checkForEmptyRequiredFields(configuration)
	rollbackMigrations(configuration, rollbackId)
}

func tryRunMigrations(configuration DatabaseOptions) {
	if configuration.Verbose {
		printOutMigrations(configuration.ConfigFileName)
	}
	checkForEmptyRequiredFields(configuration)
	runMigrations(configuration)
}

func printOutMigrations(fileName string) {
	options := loadConfigFromJsonFile(fileName)
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
