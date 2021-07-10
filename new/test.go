package main

import (
	"fmt"
	"os"
	"runtime"
)

func main()  {
	if runtime.GOOS == "windows" {
		fmt.Println("Hi there")
	}
	val, present := os.LookupEnv("path")
	fmt.Println(fmt.Sprintf("Environment variable: %t, val: %s", present, val))
}
