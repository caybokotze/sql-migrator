package main

import (
	"github.com/bxcodec/faker/v3"
	"log"
	"testing"
	"time"
)

/*
When migrations are being executed
*/
func TestThatMigrationsPersistToTheDb(t *testing.T) {
	schema, err := createRandomSchema()
	if err != nil {
		t.Error("Could not fake the data required.")
	}
	if schema.name == "" {
		t.Fail()
	}
}

func TestThatGenerateSchemaFromFileDoesCreateSchemaFromFile(t *testing.T) {
	testStrings := []string{
		"20211029162223_SomeNewMigrationFileName_up",
		"20211029162223_SomeNewMigrationFileName_down",
		"20211029162223_Some_New_Migration_File_Name_up",
	}

	for _, s := range testStrings {
		var result = generateSchemaFromFileName(s)
		if result.dateExecuted != time.Now() {
			t.Error("The tests failed because the time that was set was wrong.")
		}
		if result.id == 0 {
			t.Error("The id is wrong")
		}
		if result.name != "SomeNewMigrationFileName" {
			t.Error("the full migration name is not included.")
		}
		if result.name == "" {
			t.Error("the name should not be empty.")
		}
	}
}

// Given Migration > When Containing Multiple Lines >
//func ShouldRollbackWhenMigrationFails(t *testing.T) {
//	var file = tests.GetJsonConfigFile()
//}

func createRandomSchema() (Schema, error) {
	schema := Schema{}
	err := faker.FakeData(&schema)
	if err != nil {
		log.Fatal("The things happened.")
	}
	return schema, nil
}
