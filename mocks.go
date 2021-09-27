package main

func createDbOptionsMock() DatabaseOptions {
	return DatabaseOptions{
		SqlUser:            "sqltracking",
		SqlPassword:        "sqltracking",
		SqlHost:            "localhost",
		SqlPort:            "3306",
		SqlDatabase:        "demodb",
		DryRun:             false,
		AutoByPass:         false,
		MigrationTableName: "__migrations",
		Verbose:            false,
	}
}
