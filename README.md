# Badger CLI

A simple CLI tool for accessing badger databases.

This project was inspired on `redis-cli` so we will try to keep
the behavior simillar, for instance using it is very simple:

```bash
$ go install github.com/vingarcia/badger-cli@latest
$ badger-cli some.db
some.db> get foo
record not found
some.db> set foo bar
some.db> get foo
bar
some.db> list
- foo
- foo2
some.db> exit
```

You can also open encrypted badger instances using the `-p` argument, e.g.:

```bash
badger-cli -p some_encrypted.db
Password:
some_encrypted.db> list
- foo
- foo2
```

You can also use a pipe for passing the password to `badger-cli`, but only on \*nix systems:

```bash
cat password_file.txt | badger-cli -p some_encrypted.db
some_encrypted.db> list
- foo
- foo2
```

## Installation

You can install using the `go install` command:

```bash
go install github.com/vingarcia/badger-cli@latest
```

Which will install it on either `~/go/bin` or `$GOPATH/bin`, which you might need
to add to your `PATH`.

## Supported Commands

On the list below elements surrounded by `<>` are mandatory and
elements surrounded by `[]` are optional:

- `set <key> [value]`: Adds an item to the database, if value is not present an empty string is assumed.
- `get <key>`: Reads an item from the database (if the item doesn't exists an error is displayed)
- `list [prefix]`: Lists all keys starting with `prefix`, if `prefix` is not passed lists all keys
- `find [prefix]`: Finds all items starting with `prefix`, if `prefix` is not passed lists all values
- `del <key>`: Deletes the target item
