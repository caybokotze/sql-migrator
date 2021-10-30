package main

import (
	"github.com/bxcodec/faker/v3"
	_ "github.com/bxcodec/faker/v3"
	"testing"
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

func createRandomSchema() (Schema, error) {
	schema := Schema{}
	err := faker.FakeData(&schema)
	if err != nil {
		return Schema{}, err
	}
	return schema, nil
}
