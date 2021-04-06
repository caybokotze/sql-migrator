package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

func main() {
	sqlNew := flag.Bool("sql-new", true, "flag that set's whether a new sql migration needs to be created.")
	sqlUp := flag.Bool("sql-up", true, "flag that is set to define whether existing migrations should be run.")

	//sql_user := flag.String("sql_user", "sqltracking", "the sql user that needs to be used to execute migrations")
	//user := os.Getenv("SQL_USER")
	//password := os.Getenv("SQL_PASSWORD")
	//host := os.Getenv("SQL_HOST")
	//database := os.Getenv("SQL_DATABASE")
	//dryRun := os.Getenv("DRY_RUN")
	//mode := os.Getenv("MODE")
	//port := os.Getenv("SQL_PORT")
	//autoByPass := os.Getenv("AUTO_BYPASS")
	//currentDate := time.Now()

	if *sqlNew == true {
		createNewMigration()
	}
	if *sqlUp == true {
		fmt.Println("do the things...")
	}

	go doDoubleStuff()

	const create_schemaversion = `CREATE TABLE IF NOT EXISTS schemaversion (
id BIGINT NOT NULL AUTO_INCREMENT,
name VARCHAR(512) NULL,
date_executed DATETIME DEFAULT CURRENT_TIMESTAMP,
PRIMARY KEY (id));`

}

func printFlagValues(
	user string,
	host string,
	database string,
	dryRun bool) {
	fmt.Sprintf("user: %s", user)
}

func createNewMigration() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Create a new name for a migration: ")

	for {
		fmt.Println("-> ")
		text, _ := reader.ReadString('\n')
		text = strings.Replace(text, "\n", "", -1)

		if "hi" == text {
			fmt.Println("hello yourself.")
		}
	}
}

func doDoubleStuff() {
	time.Sleep(2000)
}