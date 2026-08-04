package main

import "testing"

func TestFetchConfiguration(t *testing.T) {
	// arrange
	expected := createDbOptionsMock()
	// act
	config := loadConfigFromJsonFile("migrator-config.json")
	// assert
	if config != expected {
		t.Errorf("expected %+v, got %+v", expected, config)
	}
}
