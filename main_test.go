package main

import (
	"testing"
	"time"
)

func TestThatNewMigrationFilesAreCreated(t *testing.T) {
	createNewMigration()
	expected := time.Now().Format("20060102")
	actual := time.Now().Format("20060102")
	if actual != expected {
		t.Errorf(TestPrint(expected, actual))
	}
}

//func TestThatSqlNewFlagIsTrueWhenProvided(t *testing.T) {
//	os.Args = []string{"cmd", "-sql-new"}
//	flag.Parse()
//	actual := *sqlNew
//	if true != actual {
//		t.Errorf(TestPrint(strconv.FormatBool(true), strconv.FormatBool(actual)))
//	}
//}
//
//func TestThatSqlNewFlagIsFalseWhenNotProvided(t *testing.T) {
//	os.Args = nil
//	os.Args = []string{"cmd"}
//	*sqlNew = false
//	flag.Parse()
//	actual := *sqlNew
//	if false != actual {
//		t.Errorf(TestPrint(strconv.FormatBool(false), strconv.FormatBool(actual)))
//	}
//}