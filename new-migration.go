package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"time"
)

func createNewMigration() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Create a new name for a migration: ")
	fmt.Println("-> ")
	text, _ := reader.ReadString('\n')
	re := regexp.MustCompile(`\r?\n`)
	text = re.ReplaceAllString(text, "")
	scriptName := getTimestampAsString() + "_" + text
	upScript := scriptName + "_up"
	downScript := scriptName + "_down"
	err := ioutil.WriteFile(fmt.Sprintf("./scripts/%s.sql", upScript), []byte(""), 0755)
	_ = ioutil.WriteFile(fmt.Sprintf("./scripts/%s.sql", downScript), []byte(""), 0755)
	if err != nil {
		fmt.Printf("Unable to write file: %v\n", err)
	}
}

func getTimestampAsString() string {
	return time.Now().Format("20060102150405")
}
