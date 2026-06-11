# Changelog

All notable changes to this project will be documented in this file.

This project follows:

- [Semantic Versioning](https://semver.org/)
- [Keep a Changelog](https://keepachangelog.com/)

---

## [0.1.2] - 2026-06-11

### Changed

- Updated `github.com/netlifeguru/db` dependency from `v0.1.1` to `v0.1.3`.
- Updated indirect dependencies in `go.sum` to match the new shared database package version.

## [0.1.1] - 2026-05-24

### Changed

- Updated `github.com/netlifeguru/db` dependency from `v0.1.0` to `v0.1.1`.

### Fixed

- Fixed `TestSelectSQL` to use the correct `db.DialectSQL` field name `Postgres` instead of `Postgresql`.
- Verified MySQL dialect SQL selection returns the MySQL query variant.

## [0.1.0] - 2026-05-23

### Added

- Initial public release
- MySQL driver for the NetLife Guru Go database layer
- MySQL database connection support
- Connection implementation compatible with the shared `db.Conn` interface
- Integration with `github.com/netlifeguru/db` query, exec, transaction, and repository helpers
- Mapper-backed result scanning through `github.com/netlifeguru/mapper`
- Struct, map, and scalar result handling through the shared database layer
- Support for MySQL `model.sql` files
- Transaction workflow support through the shared database layer
- Documentation links for NetLife Guru docs and pkg.go.dev

### Notes

- This package is the MySQL driver for the shared NetLife Guru database layer.
- Repository code can depend on `github.com/netlifeguru/db`, while application setup imports this driver.
- This is the first public `v0` release.
- The API may still change before `v1.0.0`.