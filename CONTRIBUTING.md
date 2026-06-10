# Contributing to netlifeguru/db-mysql

Hello! 👋 Thanks for your interest in contributing to the `db-mysql` project.

The project provides the MySQL driver integration for [`github.com/netlifeguru/db`](https://github.com/netlifeguru/db).

These guidelines help keep MySQL behavior, shared DB compatibility, tests, and documentation consistent.

---

## 🐛 Reporting Issues

If you find a bug or have a feature suggestion, use the [GitHub Issues](https://github.com/netlifeguru/db-mysql/issues) tab.

When reporting an issue, include:

* A clear description of the problem.
* A minimal example that reproduces the issue.
* What you expected to happen.
* What actually happened.
* The complete error message.
* Your operating system.
* Your Go version.
* Your MySQL or MariaDB version.
* The `db` and `db-mysql` package versions.
* Relevant SQL, configuration, or connection settings.

Remove passwords, connection strings, certificates, tokens, and production data before posting an issue.

---

## 📦 Pull Requests

> **Note:** Do not create pull requests directly against the `main` branch.
> Create a feature or fix branch first, such as `fix/last-insert-id` or `feat/pool-config`.

Pull requests should:

* Include tests for new functionality and bug fixes.
* Reference the related issue, if one exists.
* Be clearly named and scoped.
* Keep changes focused.
* Avoid mixing unrelated updates.
* Preserve compatibility with `github.com/netlifeguru/db`.
* Update documentation and examples when public behavior changes.

Changes should consider their effect on:

* MySQL connection pools
* transactions
* `?` placeholders
* SQL analysis
* `LastInsertId`
* rows affected
* dialect SQL selection
* row scanning
* context cancellation
* DSN configuration
* connection sharing and lifecycle

---

## 🌿 Branch Names

Use short, descriptive branch names.

Examples:

```text
feat/custom-tls-mode
fix/last-insert-id
fix/backtick-analysis
docs/connection-config
test/transaction-rollback
refactor/query-rows
```

Recommended prefixes include:

* `feat/`
* `fix/`
* `docs/`
* `test/`
* `refactor/`
* `chore/`

---

## ✏️ Commit Messages

Use conventional commit messages.

Supported prefixes include:

* `feat:` – new functionality
* `fix:` – bug fixes
* `docs:` – documentation changes
* `test:` – new or updated tests
* `refactor:` – internal code restructuring
* `chore:` – maintenance and dependency updates
* `style:` – formatting-only changes
* `perf:` – performance improvements
* `ci:` – CI/CD changes
* `build:` – build or dependency changes
* `revert:` – reverting a previous change

Examples:

```text
feat: support custom MySQL TLS mode
```

```text
fix: ignore placeholders inside backtick identifiers
```

```text
test: add shared pool lifecycle coverage
```

Each commit should represent one meaningful change.

---

## 🧪 Testing

Run the full test suite before submitting a pull request:

```bash
go test ./...
```

Disable cached results when verifying a change:

```bash
go test -count=1 ./...
```

For concurrency-sensitive changes, also run:

```bash
go test -race ./...
```

Run static analysis:

```bash
go vet ./...
```

Driver-specific changes may require integration tests against MySQL or MariaDB.

Tests should cover:

* successful connections
* invalid configuration
* DSN construction
* connection pool lifecycle
* shared pool reference counting
* transactions
* commit and rollback
* query execution
* row scanning
* context cancellation
* placeholder detection
* quoted strings
* backtick identifiers
* SQL comments
* `LastInsertId`
* rows affected
* empty queries
* missing connections

---

## 🐬 MySQL-Specific Behavior

Contributions should preserve MySQL-specific behavior explicitly.

Important MySQL characteristics include:

* `?` placeholders
* auto-increment IDs
* `LastInsertId`
* backtick identifiers
* MySQL comments
* MySQL transaction behavior
* DSN options
* `parseTime`
* connection character sets
* MySQL TLS modes

Do not introduce PostgreSQL-style returning behavior into the MySQL driver.

---

## ❓ Placeholder Analysis

Changes to SQL analysis must correctly distinguish placeholders from literal question marks.

The analyzer should ignore `?` characters inside:

* single-quoted strings
* double-quoted strings
* backtick identifiers
* line comments
* block comments

Examples that must not be treated as placeholders:

```sql
SELECT '?'
```

```sql
SELECT "?"
```

```sql
SELECT `column?`
```

```sql
-- ?
SELECT 1
```

```sql
/* ? */
SELECT 1
```

Changes in this area must include focused tests.

---

## 🆔 LastInsertId

MySQL commonly exposes generated auto-increment IDs through `LastInsertId`.

Example:

```go
result, err := db.Insert(ctx, conn, query, args...)
if err != nil {
	return err
}

id := result.LastInsertId()
```

Changes involving inserts must preserve:

* `LastInsertId`
* `RowsAffected`
* transaction behavior
* consistent shared `db.Result` behavior

Do not return fabricated IDs when the driver does not provide one.

---

## 🔄 Transactions

Transaction changes should include tests for:

* successful commit
* explicit rollback
* rollback after callback error
* rollback after panic
* invalid transaction handles
* context cancellation
* unsupported nested transaction behavior

Transaction behavior should remain compatible with the shared `db.Conn` interface.

---

## 🧩 Shared DB Compatibility

This package implements MySQL behavior for `github.com/netlifeguru/db`.

Changes to exported methods or behavior should remain compatible with the shared interfaces, including:

* `db.Conn`
* `db.Querier`
* `db.Execer`
* `db.Transactioner`
* `db.DialectSelector`
* `db.ModelLoader`
* `db.RowsProvider`

Changes to shared interfaces should be coordinated with the main `db` repository.

---

## 🧱 Connection Pools

Pool-related changes should consider:

* default values
* maximum and minimum connections
* connection lifetime
* idle timeout
* connect timeout
* pool sharing
* pool identifiers
* reference counting
* clean shutdown

Conflicting configurations must not silently reuse the same identifier.

Concurrency-sensitive pool changes must be tested with the race detector.

---

## ⚙️ DSN and Configuration

Changes to DSN construction should consider:

* host
* port
* database
* username
* password
* time zone
* timeout
* character set
* TLS mode
* `parseTime`

Passwords and sensitive connection details must not be logged.

Configuration defaults should remain predictable and documented.

---

## 📚 Documentation and Examples

Public behavior should be documented with practical examples.

Examples are maintained in:

```text
https://github.com/netlifeguru/examples/db/mysql
```

When changing public behavior, update the relevant examples where applicable, including:

* connection examples
* select helpers
* dialect SQL
* insert results
* updates and deletes
* transactions
* SQL files
* multi-driver SQL files

---

## 🎨 Code Style

Format Go code before submitting changes:

```bash
gofmt -w .
```

Prefer:

* small functions
* explicit errors
* wrapped errors with useful context
* predictable driver behavior
* focused public APIs
* tests for MySQL-specific syntax

Avoid:

* hidden query rewriting
* silent placeholder changes
* swallowing MySQL errors
* unrelated refactoring
* leaking credentials in logs

---

## 🔐 Security

Do not include credentials, certificates, connection strings, private data, or production records in:

* issues
* tests
* examples
* documentation
* commits

Use synthetic configuration values.

Avoid logging passwords or complete DSNs.

---

## 🙋 Questions

If you are unsure about MySQL behavior, shared DB compatibility, or contribution scope, open an issue or ask in an existing discussion.

Thanks for contributing! 🚀
