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
		scanner.Scan()
		line := scanner.Text()
		if len(line) == 0 {
			break
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
	if len(args) < 2 {
		return internal.ErrUnrecognizedCmd
	}

	cmd := args[0]
	key := args[1]
	switch cmd {
	case "get":
		value, err := db.Get(ctx, key)
		if err != nil {
			return err
		}
		fmt.Println(value)
		return nil
	case "set":
		if len(args) < 3 {
			return fmt.Errorf("missing third value")
		}
		return db.Set(ctx, key, args[2])
	}
	return nil
}
