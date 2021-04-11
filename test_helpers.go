package main

import "fmt"

func TestPrint(expected, actual string) string {
	return fmt.Sprintf("Test failed. Expected: '%s', but got: '%s'", expected, actual)
}

func ReturnBooleanFlag(flag bool) bool {
	return flag
}