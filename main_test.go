package main

import (
	"testing"
)

func TestDefaultPortForDriver(t *testing.T) {
	cases := map[string]string{
		"":         "3306",
		"mysql":    "3306",
		"mariadb":  "3306",
		"postgres": "5432",
		"pg":       "5432",
	}
	for driver, expected := range cases {
		if actual := defaultPortForDriver(driver); actual != expected {
			t.Errorf("defaultPortForDriver(%q): expected %s, got %s", driver, expected, actual)
		}
	}
}
