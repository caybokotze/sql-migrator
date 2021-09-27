package main

import "testing"
import tk "github.com/caybokotze/go-testing-kit"

func TestFetchConfiguration(t *testing.T) {
	// arrange
	var expected = createDbOptionsMock()
	// act
	var config = loadConfigFromJsonFile()
	// assert
	if !tk.Compare(config).To(expected) {
		t.Fail()
	}
}

func TestCreateDbConnection(t *testing.T) {
	// arrange
	// act
	// assert
}