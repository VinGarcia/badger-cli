# Badger CLI

A simple CLI tool for accessing badger databases.

This project was inspired on `redis-cli` so we will try to keep
the behavior simillar, for instance using it is very simple:

```bash
$ git clone https://github.com/vingarcia/badger-cli
$ cd badger-cli
$ go build -o badger .
$ ./badger foo.db
foo.db> get foo
record not found
foo.db> set foo bar
foo.db> get foo
bar
foo.db>
```

## Supported Commands

On the list below elements surrounded by `<>` are mandatory and
elements surrounded by `[]` are optional:

- `set <key> [value]`: Adds an item to the database, if value is not present an empty string is assumed.
- `get <key>`: Reads an item from the database (if the item doesn't exists an error is displayed)
- `list [prefix]`: Lists all keys starting with `prefix`, if `prefix` is not passed lists all keys
- `find [prefix]`: Finds all items starting with `prefix`, if `prefix` is not passed lists all values
- `del <key>`: Deletes the target item


