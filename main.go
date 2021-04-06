package main

import (
	"bufio"
	"database/sql"
	"flag"
	"fmt"
	"github.com/gookit/color"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"time"
)

func main() {
	sqlNew := flag.Bool("sql-new", false, "flag that set's whether a new sql migration needs to be created.")
	sqlUp := flag.Bool("sql-up", false, "flag that is set to define whether existing migrations should be run.")

	flag.Parse()

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

	if *sqlNew == false && *sqlUp == false {
		color.Red.Println("You didn't supply any arguments... Please try again, use -h for help.")
		os.Exit(1)
	}
	if *sqlNew == true && *sqlUp == true {
		color.Cyan.Println("You can not run sql-new and sql-up at the same time, only sql-new will be run...")
		os.Exit(1)
	}
	if *sqlNew == true {
		createNewMigration()
		os.Exit(1)
	}
	if *sqlUp == true {
		runMigrations()
		os.Exit(1)
	}
}

func createNewMigration() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Create a new name for a migration: ")
	fmt.Println("-> ")
	text, _ := reader.ReadString('\n')
	re := regexp.MustCompile(`\r?\n`)
	text = re.ReplaceAllString(text, "")
	var scriptName string = time.Now().Format("20060102") + "-" + text
	var upScript string = scriptName + "_up"
	var downScript string = scriptName + "_down"
	err := ioutil.WriteFile("./scripts/" +upScript+ ".sql", []byte(""), 0755)
	_ = ioutil.WriteFile("./scripts/"+downScript+".sql", []byte(""), 0755)
	if err != nil {
		fmt.Printf("Unable to write file: %v\n", err)
	}
}

type Schema struct {
	id int64
	name string
	dateexecuted time.Time
}

func runMigrations() {

}

func getAllMigrations() {
	db, err := sql.Open("mysql", "root:pass1@tcp(127.0.0.1:3306)/tuts")

	if err != nil {
		log.Print(err.Error())
	}

	defer db.Close()

	results, err := db.Query("SELECT `id`, `name`, `dateexecuted` FROM `schemaversion`")

	if err != nil {
		panic(err.Error())
	}

	for results.Next() {
		var schema Schema

		err = results.Scan(&schema.id, &schema.name)
		if err != nil {
			panic(err.Error())
		}

		log.Printf(schema.name)
	}
}

func createSchemaVersionTable(dbUser string, dbPassword string, ipAddress string, port string) {
	const createSchemaversion = `CREATE TABLE IF NOT EXISTS schemaversion (
	id BIGINT NOT NULL AUTO_INCREMENT,
	name VARCHAR(512) NULL,
	date_executed DATETIME DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (id));`

	db, err := sql.Open("mysql",
		fmt.Sprintf("%s:%s@tcp(%s:%s)", dbUser, dbPassword, ipAddress, port))

	if err != nil {
		panic(err.Error())
	}

	defer db.Close()

	insert, err := db.Query(createSchemaversion)

	if err != nil {
		panic(err.Error())
	}

	defer insert.Close()
}

func doDoubleStuff() {
	time.Sleep(2000)
}