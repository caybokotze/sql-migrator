package main

func createDbOptionsMock() DatabaseOptions {
	return DatabaseOptions{
		SqlUser:            "sqltracking",
		SqlPassword:        "SqlTracking#1234",
		SqlHost:            "localhost",
		SqlPort:            "3306",
		SqlDatabase:        "sql-migrator",
		DryRun:             false,
		AutoByPass:         false,
		MigrationTableName: "__migrations",
		Verbose:            false,
	}
}
