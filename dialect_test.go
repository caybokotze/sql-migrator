package main

import (
	"strings"
	"testing"
)

func TestResolveDialect(t *testing.T) {
	cases := map[string]string{
		"":           "mysql",
		"mysql":      "mysql",
		"mariadb":    "mysql",
		"MySQL":      "mysql",
		"postgres":   "postgres",
		"postgresql": "postgres",
		"pg":         "postgres",
	}
	for input, expected := range cases {
		dialect, err := resolveDialect(input)
		if err != nil {
			t.Errorf("resolveDialect(%q): unexpected error: %s", input, err.Error())
			continue
		}
		if dialect.DriverName() != expected {
			t.Errorf("resolveDialect(%q): expected %s, got %s", input, expected, dialect.DriverName())
		}
	}

	if _, err := resolveDialect("oracle"); err == nil {
		t.Error("expected an error for an unsupported driver")
	}
}

func TestMysqlDsn(t *testing.T) {
	dsn := mysqlDialect{}.BuildDsn(DatabaseOptions{
		SqlUser: "root", SqlPassword: "secret", SqlHost: "localhost", SqlPort: "3306", SqlDatabase: "demo",
	})
	expected := "root:secret@tcp(localhost:3306)/demo?multiStatements=true"
	if dsn != expected {
		t.Errorf("expected '%s', got '%s'", expected, dsn)
	}
}

func TestPostgresDsnDefaultsToSslDisable(t *testing.T) {
	dsn := postgresDialect{}.BuildDsn(DatabaseOptions{
		SqlUser: "postgres", SqlPassword: "secret", SqlHost: "127.0.0.1", SqlPort: "5432", SqlDatabase: "demo",
	})
	if !strings.HasPrefix(dsn, "postgres://postgres:secret@127.0.0.1:5432/demo") {
		t.Errorf("unexpected dsn prefix: %s", dsn)
	}
	if !strings.Contains(dsn, "sslmode=disable") {
		t.Errorf("expected sslmode=disable in dsn: %s", dsn)
	}
	if strings.Contains(dsn, "search_path") {
		t.Errorf("did not expect a search_path without a schema: %s", dsn)
	}
}

func TestPostgresDsnWithSchemaSetsSearchPath(t *testing.T) {
	dsn := postgresDialect{}.BuildDsn(DatabaseOptions{
		SqlUser: "postgres", SqlPassword: "secret", SqlHost: "127.0.0.1", SqlPort: "5432",
		SqlDatabase: "demo", SqlSchema: "org_cold_craft", SqlSslMode: "require",
	})
	if !strings.Contains(dsn, "sslmode=require") {
		t.Errorf("expected sslmode=require in dsn: %s", dsn)
	}
	if !strings.Contains(dsn, "search_path%3Dorg_cold_craft%2Cpublic") {
		t.Errorf("expected an url-encoded search_path option in dsn: %s", dsn)
	}
}

func TestPostgresDsnEncodesCredentials(t *testing.T) {
	dsn := postgresDialect{}.BuildDsn(DatabaseOptions{
		SqlUser: "user@corp", SqlPassword: "p@ss:w/rd", SqlHost: "127.0.0.1", SqlPort: "5432", SqlDatabase: "demo",
	})
	if !strings.Contains(dsn, "user%40corp") {
		t.Errorf("expected the user to be url-encoded: %s", dsn)
	}
	if strings.Contains(dsn, "p@ss:w/rd") {
		t.Errorf("expected the password to be url-encoded: %s", dsn)
	}
}

func TestPostgresTrackingTableShape(t *testing.T) {
	ddl := postgresDialect{}.CreateMigrationTableSql("__migrations")
	for _, fragment := range []string{"script_name", "applied_at", "BIGSERIAL", "UNIQUE"} {
		if !strings.Contains(ddl, fragment) {
			t.Errorf("expected '%s' in the postgres tracking DDL: %s", fragment, ddl)
		}
	}

	insertArgs := postgresDialect{}.InsertMigrationArgs(Schema{id: 20250101120000, fileBase: "20250101120000_Create_Users"})
	if len(insertArgs) != 1 || insertArgs[0] != "20250101120000_Create_Users_up.sql" {
		t.Errorf("expected the full up-script file name as the insert arg, got %v", insertArgs)
	}
}

func TestQuotePgIdentifier(t *testing.T) {
	if quotePgIdentifier("org_a") != `"org_a"` {
		t.Error("expected plain quoting")
	}
	if quotePgIdentifier(`bad"name`) != `"bad""name"` {
		t.Error("expected embedded quotes to be doubled")
	}
}
