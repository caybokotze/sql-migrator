# sql-migrator

A command-line SQL migration tool for managing database schema changes.

## Features

- Run pending migrations
- Create new migration files (up and down scripts)
- Rollback migrations to a specific point
- MySQL/MariaDB and PostgreSQL support
- Schema-per-tenant support on PostgreSQL (`--schema`)
- Fan-out mode: apply migrations (or rollbacks) to every schema returned by a discovery query (`--fan-out`)
- Dry-run mode for testing migrations without committing (fully transactional on PostgreSQL)
- Auto-bypass mode for skipping problematic migrations
- Configuration file support
- Environment variable support for passwords
- Auto-update from GitHub releases

## Supported Databases

- MySQL / MariaDB (`--driver=mysql`, the default)
- PostgreSQL (`--driver=postgres`)

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
    "sqlDriver": "mysql",
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

Rolls back applied migrations from the most recent down to and including the specified migration ID. Applied migrations below the target are left alone. If nothing at or above the target is applied, the rollback is a no-op, so re-running a partially failed rollback is safe.

```bash
sql-migrator rollback <migration-id>
```

Example:
```bash
sql-migrator rollback 20250319205322
```

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

## PostgreSQL

Select the postgres driver with `--driver=postgres` or `"sqlDriver": "postgres"` in the configuration file. The default port becomes 5432 and TLS is controlled with `--ssl-mode` (`disable` by default; also `require`, `verify-ca`, `verify-full`).

### Single schema

PostgreSQL has schemas as namespaces inside one database. Pass `--schema` to run migrations inside one schema:

```bash
sql-migrator sql-up --driver=postgres --schema=tenant_a
```

- The schema is created if it does not exist.
- The `search_path` is baked into the connection string (`<schema>,public`), so unqualified names in your migration SQL resolve inside the target schema.
- The migration tracking table lives **inside** the schema, so every schema tracks its own position independently.

### Fan-out: one migration set, many schemas

For schema-per-tenant databases, fan-out applies the migration set to every schema returned by a discovery query. Configure the query in `migrator-config.json` (or pass `--schema-query`); the first column of each row is taken as a schema name:

```json
{
  "sql-migrator-config": {
    "sqlDriver": "postgres",
    "sqlUser": "postgres",
    "sqlPassword": "ENV:PGPASSWORD",
    "sqlHost": "127.0.0.1",
    "sqlPort": "5432",
    "sqlDatabase": "app_db",
    "schemaQuery": "SELECT schema_name FROM public.tenants ORDER BY schema_name"
  }
}
```

```bash
sql-migrator sql-up --fan-out
sql-migrator rollback 20250319205322 --fan-out
```

- Schemas already up to date are skipped (each schema has its own tracking table).
- A failing schema does **not** stop the remaining schemas; the tool reports every failure at the end and exits with code 1 if any schema failed.
- New schemas returned by the query are migrated from scratch, so provisioning a tenant is just "insert a row, run fan-out".
- `--schema` and `--fan-out` are mutually exclusive.

### Tracking table shape

The two dialects track applied migrations differently:

| Dialect | Columns | Keyed by |
|---------|---------|----------|
| MySQL/MariaDB | `id` (migration timestamp), `name`, `date_executed` | parsed timestamp (legacy shape, unchanged) |
| PostgreSQL | `id BIGSERIAL`, `script_name VARCHAR(255) UNIQUE`, `applied_at TIMESTAMPTZ` | migration file name |

The PostgreSQL shape is filename-keyed, which means the tool can adopt tracking tables created by other filename-tracking migration runners in place — pre-recorded rows are honoured and no state migration is needed.

### Dry runs

On PostgreSQL, `--dry-run` executes the entire pending batch inside **one transaction** and rolls it back at the end. Migrations that depend on earlier migrations in the same batch are validated correctly, and nothing is persisted. On MySQL, DDL causes an implicit commit, so a dry run there is only an approximation.

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
    "sqlDriver": "mysql | postgres",
    "sqlUser": "username",
    "sqlPassword": "password",
    "sqlHost": "localhost",
    "sqlPort": "3306",
    "sqlDatabase": "database_name",
    "sqlSchema": "optional_postgres_schema",
    "sqlSslMode": "disable",
    "schemaQuery": "optional fan-out discovery query",
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
| `--driver` | `mysql` | Database driver: `mysql` or `postgres` |
| `-u, --user` | `sqltracking` | Database username |
| `-p, --password` | (none) | Database password |
| `--host` | `localhost` | Database host |
| `--port` | `3306` (mysql), `5432` (postgres) | Database port |
| `--database` | (none) | Database name |
| `--schema` | (none) | Postgres only: target schema |
| `--ssl-mode` | `disable` | Postgres only: sslmode for the connection |
| `--fan-out` | `false` | Postgres only: run against every discovered schema |
| `--schema-query` | (none) | Postgres only: schema discovery query for fan-out |
| `--migration-table-name` | `__migrations` | Table name for tracking migrations |
| `--dry-run` | `false` | Execute migrations without committing |
| `--auto-bypass` | `false` | Skip failed migrations and record them as completed |
| `--verbose` | `false` | Enable verbose output |

### Environment Variables for Passwords

For security, passwords can be provided via environment variables:

```bash
sql-migrator sql-up --password=ENV:MY_DB_PASSWORD
```

This will read the password from the `MY_DB_PASSWORD` environment variable. The same `ENV:` form works in the configuration file.

## Migration Files

### Naming Convention

Migration files follow this naming pattern:

```
{timestamp}_{migration_name}_{direction}.sql
```

- `timestamp`: YYYYMMDDHHmmss format (e.g. 20250319205322)
- `migration_name`: Descriptive name (underscores inside the name are preserved)
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

## Usage Examples

### Run migrations with specific credentials

```bash
sql-migrator sql-up --user=admin --password=secret --database=myapp
```

### Test migrations without applying changes

```bash
sql-migrator sql-up --dry-run
```

### Migrate every tenant schema on a postgres server

```bash
sql-migrator sql-up --driver=postgres --fan-out --schema-query="SELECT schema_name FROM public.tenants"
```

### Use a custom scripts directory

```bash
sql-migrator sql-up --migration-directory=./database/migrations/
```

## License

MIT License
