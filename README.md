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
