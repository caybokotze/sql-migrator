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
	fmt.Print("-> ")
	text, _ := reader.ReadString('\n')
	re := regexp.MustCompile(`\r?\n`)
	text = re.ReplaceAllString(text, "")
	scriptName := getTimestampAsString() + "_" + text
	upScript := scriptName + "_up"
	downScript := scriptName + "_down"
	_ = os.Mkdir("scripts", 0755)
	err := ioutil.WriteFile(fmt.Sprintf("./scripts/%s.sql", upScript), []byte(""), 0755)
	if err != nil {
		panic(err.Error())
	}
	_ = ioutil.WriteFile(fmt.Sprintf("./scripts/%s.sql", downScript), []byte(""), 0755)
}

func getTimestampAsString() string {
	return time.Now().Format("20060102150405")
}
