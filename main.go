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

const version = "1.3.0"

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
		driver = kingpin.Flag("driver",
			"database driver to use: mysql (default) or postgres").Default("").String()
		dsn = kingpin.Flag("dsn",
			"full connection string, replaces user/password/host/port/database (postgres://... for postgres, user:pass@tcp(host:port)/db for mysql)").Default("").String()
		user = kingpin.Flag("user",
			"username required to connect to the database").Default("sqltracking").Short('u').String()
		password = kingpin.Flag("password",
			"password required to connect to the database. Passwords can also be provided as environment variables. Use `ENV:NameOfVariable` formatting...").Default("").Short('p').String()
		port               = kingpin.Flag("port", "port number the database is active on (3306 for mysql, 5432 for postgres)").Default("").String()
		database           = kingpin.Flag("database", "database to connect to").Default("").String()
		host               = kingpin.Flag("host", "host that the database is running on").Default("localhost").String()
		schema             = kingpin.Flag("schema", "postgres only: run migrations inside this schema (created if missing, tracking table lives inside it)").Default("").String()
		sslMode            = kingpin.Flag("ssl-mode", "postgres only: sslmode for the connection (disable, require, verify-ca, verify-full)").Default("").String()
		fanOut             = kingpin.Flag("fan-out", "postgres only: run against every schema returned by the schema discovery query").Default("false").Bool()
		schemaQuery        = kingpin.Flag("schema-query", "postgres only: query that returns the schemas to fan out over (first column = schema name)").Default("").String()
		dryRun             = kingpin.Flag("dry-run", "run migrations without committing the transaction to tests for any issues.").Default("false").Bool()
		autoByPass         = kingpin.Flag("auto-bypass", "bypass problematic migrations -> record them as if they were completed.").Default("false").Bool()
		migrationTableName = kingpin.Flag("migration-table-name", "set name for migration table").Default("__migrations").String()
		verbose            = kingpin.Flag("verbose", "make more noise").Default("false").Bool()
	)

	var buildDatabaseConfig = func() DatabaseOptions {
		config := loadConfigFromJsonFile(*fileName)

		if *driver == "" {
			if config.SqlDriver != "" {
				driver = &config.SqlDriver
			}
		}
		if *dsn == "" {
			if config.SqlDsn != "" {
				dsn = &config.SqlDsn
			}
		}
		if *dsn != "" {
			if strings.HasPrefix(*dsn, "ENV:") {
				envVariableValue := os.Getenv(strings.SplitN(*dsn, ":", 2)[1])
				dsn = &envVariableValue
				fmt.Println("Using connection string from environment variable")
			}
		}
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
		if *port == "" {
			if config.SqlPort != "" {
				port = &config.SqlPort
			}
		}
		if *database == "" {
			if config.SqlDatabase != "" {
				database = &config.SqlDatabase
			}
		}
		if *schema == "" {
			if config.SqlSchema != "" {
				schema = &config.SqlSchema
			}
		}
		if *sslMode == "" {
			if config.SqlSslMode != "" {
				sslMode = &config.SqlSslMode
			}
		}
		if *schemaQuery == "" {
			if config.SchemaQuery != "" {
				schemaQuery = &config.SchemaQuery
			}
		}

		if *migrationTableName == "__migrations" {
			if config.MigrationTableName != "" {
				migrationTableName = &config.MigrationTableName
			}
		}

		options := DatabaseOptions{
			SqlDriver:          *driver,
			SqlDsn:             *dsn,
			SqlUser:            *user,
			SqlPassword:        *password,
			SqlHost:            *host,
			SqlPort:            *port,
			SqlDatabase:        *database,
			SqlSchema:          *schema,
			SqlSslMode:         *sslMode,
			SchemaQuery:        *schemaQuery,
			DryRun:             *dryRun,
			AutoByPass:         *autoByPass,
			Verbose:            *verbose,
			MigrationTableName: *migrationTableName,
			ConfigFileName:     *fileName,
			MigrationDirectory: *scriptDirectory,
		}

		if options.SqlPort == "" {
			options.SqlPort = defaultPortForDriver(options.SqlDriver)
		}

		return options
	}

	switch kingpin.Parse() {
	case newCommand.FullCommand():
		createNewMigration(*scriptDirectory)
		return
	case upCommand.FullCommand():
		options := buildDatabaseConfig()
		checkForEmptyRequiredFields(options)
		checkForConflictingSchemaFlags(options, *fanOut)
		if *fanOut {
			exitOnError(runFanOut(options, "sql-up", runMigrations))
			return
		}
		exitOnError(runMigrations(options))
		return
	case rollbackCommand.FullCommand():
		options := buildDatabaseConfig()
		checkForEmptyRequiredFields(options)
		checkForConflictingSchemaFlags(options, *fanOut)
		targetId := *rollbackTargetId
		if *fanOut {
			exitOnError(runFanOut(options, "rollback "+targetId, func(schemaOptions DatabaseOptions) error {
				return rollbackMigrations(schemaOptions, targetId)
			}))
			return
		}
		exitOnError(rollbackMigrations(options, targetId))
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

func defaultPortForDriver(driver string) string {
	dialect, err := resolveDialect(driver)
	if err == nil && dialect.DriverName() == DriverPostgres {
		return "5432"
	}
	return "3306"
}

func exitOnError(err error) {
	if err != nil {
		color.Red.Println(err.Error())
		os.Exit(1)
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

func checkForEmptyRequiredFields(configuration DatabaseOptions) {
	dialect, err := resolveDialect(configuration.SqlDriver)
	if err != nil {
		color.Red.Println(err.Error())
		os.Exit(1)
	}
	if configuration.SqlDsn != "" {
		if err := dialect.ValidateDsn(configuration.SqlDsn); err != nil {
			color.Red.Println(err.Error())
			os.Exit(1)
		}
		return
	}
	if configuration.SqlUser == "" || configuration.SqlPassword == "" || configuration.SqlDatabase == "" {
		color.Red.Println("You are required to provide a sql user, password and database name (or a full --dsn), either as an argument or environment variable")
		fmt.Println(`e.g: sql-migrator sql-up --database="doggy_db" --user="doggo" --password="le-woof"`)
		os.Exit(1)
	}
}

func checkForConflictingSchemaFlags(configuration DatabaseOptions, fanOut bool) {
	if fanOut && configuration.SqlSchema != "" {
		color.Red.Println("--schema and --fan-out cannot be combined: fan-out discovers its schemas with the schema query")
		os.Exit(1)
	}
	if configuration.SqlSchema != "" || fanOut {
		dialect, _ := resolveDialect(configuration.SqlDriver)
		if dialect != nil && !dialect.SupportsSchemas() {
			color.Red.Println("--schema and --fan-out require the postgres driver (--driver=postgres)")
			os.Exit(1)
		}
	}
}
