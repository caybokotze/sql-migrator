# sql-migrator

A command-line SQL migration tool for managing database schema changes.

## Features

- Run pending migrations
- Create new migration files (up and down scripts)
- Rollback migrations to a specific point
- Dry-run mode for testing migrations without committing
- Auto-bypass mode for skipping problematic migrations
- Configuration file support
- Environment variable support for passwords
- Auto-update from GitHub releases

## Supported Databases

- MySQL (fully supported)
- MariaDB (compatible with MySQL driver)

## Installation

Download the latest release from [GitHub Releases](https://github.com/caybokotze/sql-migrator/releases) or build from source:

```bash
go build -o sql-migrator .
```

## Quick Start

1. Create a configuration file `migrator-config.json`:

```json
{
  "sql-migrator-config": {
    "sqlUser": "your_user",
    "sqlPassword": "your_password",
    "sqlHost": "localhost",
    "sqlPort": "3306",
    "sqlDatabase": "your_database"
  }
}
```

2. Create a new migration:

```bash
sql-migrator sql-new
```

3. Edit the generated SQL files in the `scripts/` directory.

4. Run migrations:

```bash
sql-migrator sql-up
```

## Commands

### sql-new

Creates a new migration file pair (up and down scripts).

```bash
sql-migrator sql-new
```

This will prompt for a migration name and create two files:
- `scripts/{timestamp}_{name}_up.sql` - Forward migration
- `scripts/{timestamp}_{name}_down.sql` - Rollback migration

### sql-up

Executes all pending migrations in order.

```bash
sql-migrator sql-up
```

### rollback

Rolls back migrations from the most recent to the specified migration ID (inclusive).

```bash
sql-migrator rollback <migration-id>
```

Example:
```bash
sql-migrator rollback 20250319205322
```

This will rollback all migrations from the most recent one down to and including the specified migration.

### version

Displays the current version of sql-migrator.

```bash
sql-migrator version
```

### update

Auto-updates the tool from the latest GitHub release.

```bash
sql-migrator update
```

## Configuration

### Configuration File

By default, sql-migrator looks for `migrator-config.json` in the current directory. You can specify a different file:

```bash
sql-migrator sql-up --configuration-file=custom-config.json
```

Configuration file structure:

```json
{
  "sql-migrator-config": {
    "sqlUser": "username",
    "sqlPassword": "password",
    "sqlHost": "localhost",
    "sqlPort": "3306",
    "sqlDatabase": "database_name",
    "migrationTableName": "__migrations"
  }
}
```

### Command-Line Flags

All configuration options can be overridden via command-line flags:

| Flag | Default | Description |
|------|---------|-------------|
| `--configuration-file` | `migrator-config.json` | Path to configuration file |
| `--migration-directory` | `./scripts/` | Directory containing migration files |
| `-u, --user` | `sqltracking` | Database username |
| `-p, --password` | (none) | Database password |
| `--host` | `localhost` | Database host |
| `--port` | `3306` | Database port |
| `--database` | (none) | Database name |
| `--migration-table-name` | `__migrations` | Table name for tracking migrations |
| `--dry-run` | `false` | Execute migrations without committing |
| `--auto-bypass` | `false` | Skip failed migrations and record them as completed |
| `--verbose` | `false` | Enable verbose output |

### Environment Variables for Passwords

For security, passwords can be provided via environment variables:

```bash
sql-migrator sql-up --password=ENV:MY_DB_PASSWORD
```

This will read the password from the `MY_DB_PASSWORD` environment variable.

## Migration Files

### Naming Convention

Migration files follow this naming pattern:

```
{timestamp}_{migration_name}_{direction}.sql
```

- `timestamp`: YYYYMMDDHHmmss format (e.g., 20250319205322)
- `migration_name`: Descriptive name without spaces
- `direction`: Either `up` (forward) or `down` (rollback)

Example:
```
20250319205322_CreateUsersTable_up.sql
20250319205322_CreateUsersTable_down.sql
```

### Example Migration

**Up migration** (`CreateUsersTable_up.sql`):
```sql
CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

**Down migration** (`CreateUsersTable_down.sql`):
```sql
DROP TABLE users;
```

## Migration Tracking

Executed migrations are tracked in a database table (default: `__migrations`) with the following schema:

| Column | Type | Description |
|--------|------|-------------|
| id | BIGINT | Migration timestamp (primary key) |
| name | VARCHAR(255) | Migration name |
| date_executed | DATETIME | When the migration was executed |

## Usage Examples

### Run migrations with specific credentials

```bash
sql-migrator sql-up --user=admin --password=secret --database=myapp
```

### Test migrations without applying changes

```bash
sql-migrator sql-up --dry-run
```

### Rollback the last 3 migrations

First, list your migrations to find the target ID, then rollback:

```bash
sql-migrator rollback 20250319205322
```

### Use a custom scripts directory

```bash
sql-migrator sql-up --migration-directory=./database/migrations/
```

## License

MIT License
