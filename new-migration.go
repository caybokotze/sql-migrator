package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gookit/color"
)

func createNewMigration(migrationDirectory string) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Enter a name for the new migration:")
	fmt.Print("-> ")
	text, _ := reader.ReadString('\n')
	migrationName := strings.TrimSpace(text)
	migrationName = strings.ReplaceAll(migrationName, " ", "")

	if migrationName == "" {
		color.Red.Println("Migration name cannot be empty.")
		os.Exit(1)
	}

	scriptName := getTimestampAsString() + "_" + migrationName
	upScript := scriptName + "_up"
	downScript := scriptName + "_down"

	if err := os.MkdirAll(migrationDirectory, 0755); err != nil {
		color.Red.Printf("Could not create migration directory: %s\n", err.Error())
		os.Exit(1)
	}

	upPath := fmt.Sprintf("%s/%s.sql", migrationDirectory, upScript)
	downPath := fmt.Sprintf("%s/%s.sql", migrationDirectory, downScript)

	if err := os.WriteFile(upPath, []byte(""), 0644); err != nil {
		color.Red.Printf("Could not create up migration file: %s\n", err.Error())
		os.Exit(1)
	}
	if err := os.WriteFile(downPath, []byte(""), 0644); err != nil {
		color.Red.Printf("Could not create down migration file: %s\n", err.Error())
		os.Exit(1)
	}

	color.Green.Println("Migration created successfully.")
	fmt.Printf("  Up:   %s\n", upPath)
	fmt.Printf("  Down: %s\n", downPath)
	color.Cyan.Println("Run 'sql-migrator sql-up' to execute pending migrations.")
	openFilesInFileEditor(upScript, downScript, migrationDirectory)
}

func getTimestampAsString() string {

	return time.Now().Format("20060102150405")
}
