package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/vingarcia/badger-cli/internal"
	"github.com/vingarcia/badger-cli/internal/badger"
)

func main() {
	ctx := context.Background()

	if len(os.Args) <= 1 {
		log.Fatalf("missing badger db filepath")
	}

	filePath := os.Args[1]
	baseFilename := filepath.Base(filePath)

	db, err := badger.New(ctx, filePath)
	if err != nil {
		log.Fatalf("error connecting to database: %s", err)
	}

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Printf("%s> ", baseFilename)
		more := scanner.Scan()
		line := strings.TrimLeft(scanner.Text(), "\t ")
		if !more || strings.HasPrefix(line, "exit") {
			// Add a \n so it looks better on the terminal
			fmt.Println()

			break
		}

		if line == "" {
			continue
		}

		err := runCommand(ctx, db, line)
		if err != nil {
			fmt.Println(err)
		}
	}

	err = scanner.Err()
	if err != nil {
		log.Fatal(err)
	}
}

func runCommand(ctx context.Context, db badger.Client, line string) error {
	args := strings.SplitN(line, " ", 3)
	cmd := args[0]
	switch cmd {
	case "get":
		if len(args) < 2 {
			return fmt.Errorf("missing <key>, usage: get <key>")
		}

		key := args[1]
		value, err := db.Get(ctx, key)
		if err != nil {
			return err
		}
		fmt.Println(value)
		return nil

	case "set":
		if len(args) < 2 {
			return fmt.Errorf("missing [key] and [value], usage: set <key> [value]")
		}

		key := args[1]
		value := ""
		if len(args) >= 3 {
			value = args[2]
		}
		return db.Set(ctx, key, value)

	case "list":
		prefix := ""
		if len(args) >= 2 {
			prefix = args[1]
		}

		keys, err := db.List(ctx, prefix)
		if err != nil {
			return err
		}

		for _, key := range keys {
			fmt.Println("-", key)
		}

	case "find":
		prefix := ""
		if len(args) >= 2 {
			prefix = args[1]
		}

		kvs, err := db.Find(ctx, prefix)
		if err != nil {
			return err
		}

		for _, kv := range kvs {
			fmt.Printf("- %s: '%s'\n", kv.Key, kv.Value)
		}

	case "del":
		if len(args) < 2 {
			return fmt.Errorf("missing <key>, usage: delete <key>")
		}

		key := args[1]
		return db.Delete(ctx, key)

	default:
		return internal.ErrUnrecognizedCmd
	}
	return nil
}
