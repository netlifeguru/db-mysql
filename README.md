# NLG MySQL Driver

MySQL driver for the NetLife Guru Go database layer.

> Use mysql to connect MySQL databases to the shared `github.com/netlifeguru/db` repository API with mapper-backed result scanning.

[![Go Reference](https://pkg.go.dev/badge/github.com/netlifeguru/db-mysql.svg)](https://pkg.go.dev/github.com/netlifeguru/db-mysql)
[![Go Report Card](https://goreportcard.com/badge/github.com/netlifeguru/db-mysql)](https://goreportcard.com/report/github.com/netlifeguru/db-mysql)
[![Go Version](https://img.shields.io/badge/go-%3E=1.24-blue)](https://go.dev)
[![License](https://img.shields.io/badge/license-MIT-brightgreen)](LICENSE)

---

## About

`mysql` is the MySQL driver package for the NetLife Guru Go database stack.

It provides real MySQL database connections that implement the shared `db.Conn` interface from `github.com/netlifeguru/db`.

Application code normally imports this package to open MySQL connections, while repository code can depend on the shared `db` package.

## How It Fits Together

| Layer  | Package                           | Purpose                         |
|--------|-----------------------------------|---------------------------------|
| Mapper | `github.com/netlifeguru/mapper`   | Row-to-struct and map scanning  |
| DB     | `github.com/netlifeguru/db`       | Shared query and repository API |
| Driver | `github.com/netlifeguru/db-mysql` | Real MySQL database connections |

## Features

- **MySQL Driver**: Connects the shared NetLife Guru database layer to MySQL
- **Shared DB Interface**: Provides connections compatible with `db.Conn`
- **Repository Friendly**: Lets repository code use common `db` helpers such as `List`, `Get`, `Value`, and `Maps`
- **Mapper Integration**: Uses `github.com/netlifeguru/mapper` for struct, map, and scalar result scanning
- **SQL Model Support**: Works with MySQL `model.sql` files through the shared `db` package
- **Transaction Support**: Supports transaction workflows through the shared database layer
- **Explicit SQL**: Designed for applications that prefer direct SQL and typed repository helpers
- **Standard Go Friendly**: Built around context-aware operations, interfaces, structs, and explicit error handling

## Requirements

This package requires Go 1.24 or newer.

- **Go:** `1.24` or newer
- **Shared dependencies:** `github.com/netlifeguru/db`, `github.com/netlifeguru/mapper`
- **Database:** MySQL-compatible server

## Installation

Add the MySQL driver to your project using `go get`:

```bash
go get github.com/netlifeguru/db-mysql
```

This also installs the shared `db` and `mapper` packages required by the driver.

## Basic Usage

```go
import (
	"context"

	"github.com/netlifeguru/db"
	"github.com/netlifeguru/db-mysql"
)
```

Once a MySQL connection is created, repository code can work with the shared `db.Conn` interface:

```go
func ListUsers(ctx context.Context, conn db.Conn) ([]User, error) {
	return db.List[User](ctx, conn, db.Raw(`
		SELECT *
		FROM users
		ORDER BY created_at DESC
	`))
}
```

## Related Packages

- [github.com/netlifeguru/db](https://github.com/netlifeguru/db) — shared database layer
- [github.com/netlifeguru/mapper](https://github.com/netlifeguru/mapper) — row-to-struct mapper
- [github.com/netlifeguru/postgresql](https://github.com/netlifeguru/postgresql) — PostgreSQL driver
- [github.com/netlifeguru/scylla](https://github.com/netlifeguru/scylla) — ScyllaDB driver

---

## Documentation

Full package documentation, guides, and examples are available at:

[https://netlife.guru/docs/go/mysql](https://netlife.guru/docs/go/mysql)

Shared database layer documentation:

[https://netlife.guru/docs/go/db](https://netlife.guru/docs/go/db)

API reference is also available on pkg.go.dev:

[https://pkg.go.dev/github.com/netlifeguru/db-mysql](https://pkg.go.dev/github.com/netlifeguru/db-mysql)

---

## Notes

- This package is the MySQL driver for the shared NetLife Guru database layer.
- Repository code can depend on `github.com/netlifeguru/db` while application setup imports this driver.
- Review package-specific concurrency behavior before using it in highly parallel workloads.
- Check performance characteristics when using this package in latency-sensitive paths.
- See the package documentation and examples for limitations and recommended usage patterns.

---

## Versioning

This project follows Semantic Versioning.

See [`CHANGELOG.md`](CHANGELOG.md) for release history and breaking changes.

---

## Contributing

Community contributions, feedback, and improvements are welcome.

Please read [`CONTRIBUTING.md`](./CONTRIBUTING.md) before submitting pull requests or opening issues.

---

## Code of Conduct

This project follows a Code of Conduct.

Please read [`CODE_OF_CONDUCT.md`](CODE_OF_CONDUCT.md) before contributing or participating in discussions.

---

## Author

Created and maintained by [NetLife Guru s.r.o.](https://netlife.guru)

- Documentation: [https://netlife.guru/docs/go/mysql](https://netlife.guru/docs/go/mysql)
- GitHub: [https://github.com/netlifeguru](https://github.com/netlifeguru)
- Contact: [info@netlife.guru](mailto:info@netlife.guru)

---

## License

MIT License. See [`LICENSE`](LICENSE).