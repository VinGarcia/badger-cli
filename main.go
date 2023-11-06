package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/vingarcia/badger-cli/internal"
	"github.com/vingarcia/badger-cli/internal/badger"
	"gopkg.in/yaml.v2"
)

func main() {
	ctx := context.Background()

	var filePath string
	var password string
	err := parseArgs(os.Args, map[any]arg{
		1: arg{
			Target: &filePath,
			Desc:   "filepath to the database file",
		},
		"p": arg{
			Target: &password,
			Desc:   "specify the database password",
		},
		"password": arg{
			Target: &password,
			Desc:   "specify the database password",
		},
	})
	if err != nil {
		fmt.Printf("error parsing command line args: %s\n", err)
		os.Exit(1)
	}

	if filePath == "" {
		fmt.Println("missing badger db filepath")
		os.Exit(1)
	}

	baseFilename := filepath.Base(filePath)

	db, err := badger.New(ctx, filePath, []byte(password))
	if err != nil {
		fmt.Printf("error connecting to database: %s\n", err)
		os.Exit(1)
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
		fmt.Printf("unexpected error parsing input: %s\n", err)
		os.Exit(1)
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

type arg struct {
	Target any
	Desc   string
}

func parseArgs(args []string, config map[any]arg) error {
	var numPosArgs int
	for i := 0; i < len(args); i++ {
		if args[i][0] != '-' {
			c, found := config[numPosArgs]
			numPosArgs++
			if !found {
				continue
			}

			err := yaml.Unmarshal([]byte(args[i]), c.Target)
			if err != nil {
				return fmt.Errorf("unable to parse cli arg on pos '%d' with value '%s': %w", i, args[i], err)
			}

			continue
		}

		key := strings.TrimLeft(args[i], "-")
		if len(key) == 0 {
			continue
		}

		t := reflect.TypeOf(config[key].Target)
		fmt.Println("value of t", t, "key is:", key)
		if t == nil || t.Kind() != reflect.Pointer {
			return fmt.Errorf("code error: expected arg.Target to be a pointer but got: %v", t)
		}

		t = t.Elem()

		value := "true"
		if t.Kind() != reflect.Bool {
			i++
			value = args[i]
		}

		err := yaml.Unmarshal([]byte(value), config[key].Target)
		if err != nil {
			return fmt.Errorf("unable to parse cli arg '%s' with value '%s': %w", key, value, err)
		}
	}

	return nil
}
