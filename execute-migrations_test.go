package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseSchemaFromFileName(t *testing.T) {
	cases := []struct {
		fileName    string
		expectId    int64
		expectName  string
		expectBase  string
		expectError bool
	}{
		{"20211029162223_SomeNewMigrationFileName_up.sql", 20211029162223, "SomeNewMigrationFileName", "20211029162223_SomeNewMigrationFileName", false},
		{"20211029162223_SomeNewMigrationFileName_down.sql", 20211029162223, "SomeNewMigrationFileName", "20211029162223_SomeNewMigrationFileName", false},
		{"20211029162223_Some_New_Migration_File_Name_up.sql", 20211029162223, "SomeNewMigrationFileName", "20211029162223_Some_New_Migration_File_Name", false},
		{"20211029162223_SomeNewMigrationFileName_up", 20211029162223, "SomeNewMigrationFileName", "20211029162223_SomeNewMigrationFileName", false},
		{"NotATimestamp_Name_up.sql", 0, "", "", true},
		{"20211029162223_MissingDirection.sql", 0, "", "", true},
		{"readme.md", 0, "", "", true},
	}

	for _, c := range cases {
		result, err := parseSchemaFromFileName(c.fileName)
		if c.expectError {
			if err == nil {
				t.Errorf("%s: expected an error, got none", c.fileName)
			}
			continue
		}
		if err != nil {
			t.Errorf("%s: unexpected error: %s", c.fileName, err.Error())
			continue
		}
		if result.id != c.expectId {
			t.Errorf("%s: expected id %d, got %d", c.fileName, c.expectId, result.id)
		}
		if result.name != c.expectName {
			t.Errorf("%s: expected name '%s', got '%s'", c.fileName, c.expectName, result.name)
		}
		if result.fileBase != c.expectBase {
			t.Errorf("%s: expected fileBase '%s', got '%s'", c.fileName, c.expectBase, result.fileBase)
		}
	}
}

func TestUpScriptNamePrefersStoredScriptName(t *testing.T) {
	fromDb := Schema{id: 1, scriptName: "20250101_Stored_Name_up.sql", fileBase: "20250101_Other"}
	if fromDb.UpScriptName() != "20250101_Stored_Name_up.sql" {
		t.Errorf("expected the stored script_name to win, got '%s'", fromDb.UpScriptName())
	}

	fromDisk := Schema{id: 1, fileBase: "20250101_From_Disk"}
	if fromDisk.UpScriptName() != "20250101_From_Disk_up.sql" {
		t.Errorf("expected the fileBase-derived name, got '%s'", fromDisk.UpScriptName())
	}
}

func TestFindUnexecutedMigrations(t *testing.T) {
	files := []Schema{{id: 1}, {id: 2}, {id: 3}}
	executed := []Schema{{id: 1}, {id: 3}}

	pending := findUnexecutedMigrations(files, executed)

	if len(pending) != 1 || pending[0].id != 2 {
		t.Errorf("expected only migration 2 to be pending, got %v", pending)
	}
}

func TestReadSchemaContentStripsUtf8Bom(t *testing.T) {
	dir := t.TempDir()

	sqlBody := "CREATE TABLE example (id INT);"
	bomPrefixed := append([]byte{0xEF, 0xBB, 0xBF}, []byte(sqlBody)...)

	fileBase := "20250319205322_ExampleMigration"
	filePath := filepath.Join(dir, fileBase+"_up.sql")
	if err := os.WriteFile(filePath, bomPrefixed, 0o644); err != nil {
		t.Fatalf("could not write fixture file: %s", err.Error())
	}

	schema := Schema{id: 20250319205322, fileBase: fileBase}

	content, err := readSchemaContent(schema, true, dir)
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}

	if strings.HasPrefix(content, string(rune(0xFEFF))) {
		t.Errorf("expected the BOM to be stripped, but the content still starts with it")
	}
	if len(content) >= 3 && content[0] == 0xEF && content[1] == 0xBB && content[2] == 0xBF {
		t.Errorf("expected no leading BOM bytes, got content starting with %q", content[:3])
	}
	if content != sqlBody {
		t.Errorf("expected content %q, got %q", sqlBody, content)
	}
}

func TestStripUtf8Bom(t *testing.T) {
	withBom := append([]byte{0xEF, 0xBB, 0xBF}, []byte("SELECT 1;")...)
	stripped := stripUtf8Bom(withBom)
	if string(stripped) != "SELECT 1;" {
		t.Errorf("expected BOM stripped, got %q", string(stripped))
	}

	noBom := []byte("SELECT 1;")
	stripped = stripUtf8Bom(noBom)
	if string(stripped) != "SELECT 1;" {
		t.Errorf("expected no change for content without a BOM, got %q", string(stripped))
	}
}

func TestRemoveDuplicateSchemas(t *testing.T) {
	schemas := []Schema{{id: 1, name: "a"}, {id: 2, name: "b"}, {id: 1, name: "c"}}

	unique := removeDuplicateSchemas(schemas)

	if len(unique) != 2 {
		t.Errorf("expected 2 unique schemas, got %d", len(unique))
	}
	// The later entry wins, matching the original behaviour.
	if unique[0].name != "c" {
		t.Errorf("expected the later duplicate to win, got '%s'", unique[0].name)
	}
}
