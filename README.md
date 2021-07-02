# sql-migrator
*A sql migration tool for managing migrations.*

> This will build for Windows or Unix environments.

## Supports
- MySql
- MsSql (Coming soon)
- Oracle (A definate maybe)
- Postgres (Most likely)
- SqlLite (Could be useful)


## How to get started

### You will need
- Previous .exe release (on Windows) here: https://github.com/caybokotze/sql-migrator/releases/tag/1.1
- package.json file (for conveniance but not strictly required)

### Package.json example
*The package.json file isn't required, but if you have node will save you the time of specifying all the flags manually each time.*

```json
{
  "name": "sql-migration-runner",
  "version": "1.0.0",
  "description": "It migrates migrations...",
  "scripts": {
    "sql-up": "mysql-migrator.exe -sql-up=true -sql-user=dbuser -sql-password=dbpassword -sql-host=localhost -sql-port=3306 -sql-database=demodb",
    "sql-new": "mysql-migrator.exe -sql-new=true -sql-user=dbuser -sql-password=dbpassword -sql-host=localhost -sql-port=3306 -sql-database=demodb"
  },
  "config": {
    "dbUser": "dbuser",
    "dbPassword": "dbpassword",
    "sqlHost": "localhost",
    "sqlPort": "3306",
    "sqlDatabase": "demodb"
  }
}
```
