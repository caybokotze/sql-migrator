package main

import (
	"bufio"
	"fmt"
	"github.com/gookit/color"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

func createNewMigration() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Create a new name for a migration: ")
	fmt.Print("-> ")
	text, _ := reader.ReadString('\n')
	fileName := strings.TrimSpace(text)
	fileName = strings.ReplaceAll(fileName, " ", "")
	scriptName := getTimestampAsString() + "_" + fileName
	upScript := scriptName + "_up"
	downScript := scriptName + "_down"
	_ = os.Mkdir("scripts", 0755)
	err := ioutil.WriteFile(fmt.Sprintf("./scripts/%s.sql", upScript), []byte(""), 0755)
	if err != nil {
		panic(err.Error())
	}
	_ = ioutil.WriteFile(fmt.Sprintf("./scripts/%s.sql", downScript), []byte(""), 0755)
	color.Green.Println("Migration was created successfully.")
	color.Cyan.Print("Do you want to open the migrations in VsCode? (y/n): ")
	option, _ := reader.ReadString('\n')
	option = strings.ToLower(option)
	option = strings.TrimSpace(option)
	if option == "y" || option == "yes" {
		openFilesInVsCode(upScript, downScript)
	}
	color.Blue.Println("To run migrations use -sql-up=true flag option.")
}

func getTimestampAsString() string {
	return time.Now().Format("20060102150405")
}
