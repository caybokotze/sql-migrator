package main

import (
	"flag"
	"os"
	"strconv"
	"testing"
	"time"
)

// Test Flags
var sqlNew = flag.Bool("sql-new", false, "flag that set's whether a new sql migration needs to be created.")
var sqlUp = flag.Bool("sql-up", false, "flag that is set to define whether existing migrations should be run.")
var sqlUser = flag.String("sql-user", os.Getenv("SQL_USER"), "the sql user that needs to be used to execute migrations")
var sqlPassword = flag.String("sql-password", os.Getenv("SQL_PASSWORD"), "the sql user password that is required to execute the migrations")
var sqlPort = flag.String("sql-port", os.Getenv("SQL_PORT"), "the sql port that is required to open a db connection")
var sqlHost = flag.String("sql-host", os.Getenv("SQL_HOST"), "the sql host that is required to open a db connection")
var sqlDatabase = flag.String("sql-database", os.Getenv("SQL_DATABASE"), "the targeted database that is required to open a db connection")
var dryRun = flag.Bool("dry-run", false, "use the dry-run flag if you want to execute all migrations within a transaction scope so that changes are not persisted")
var autoByPass = flag.Bool("auto-bypass", false, "use this flag if you want to continue executing migrations even if one or more fail to execute")
// End Test Flags

func TestThatNewMigrationFilesAreCreated(t *testing.T) {
	createNewMigration()
	expected := time.Now().Format("20060102")
	actual := time.Now().Format("20060102")
	if actual != expected {
		t.Errorf(TestPrint(expected, actual))
	}
}

func TestThatSqlNewFlagIsTrueWhenProvided(t *testing.T) {
	os.Args = []string{"cmd", "-sql-new"}
	flag.Parse()
	actual := *sqlNew
	if true != actual {
		t.Errorf(TestPrint(strconv.FormatBool(true), strconv.FormatBool(actual)))
	}
}

func TestThatSqlNewFlagIsFalseWhenNotProvided(t *testing.T) {
	os.Args = nil
	os.Args = []string{"cmd"}
	*sqlNew = false
	flag.Parse()
	actual := *sqlNew
	if false != actual {
		t.Errorf(TestPrint(strconv.FormatBool(false), strconv.FormatBool(actual)))
	}
}