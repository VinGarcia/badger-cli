package badger

import (
	"context"
	"errors"
	"fmt"
	"time"

	"sync/atomic"

	badger "github.com/dgraph-io/badger/v3"
	"github.com/vingarcia/badger-cli/internal"
	"golang.org/x/sync/errgroup"
)

const (
	KB = 1 << 10
	MB = 1 << 20
	GB = 1 << 30
)

type Client struct {
	// logger log.Provider

	db       *badger.DB
	g        *errgroup.Group
	doneCh   chan struct{}
	isClosed *atomic.Bool
}

func New(
	ctx context.Context,
	// logger log.Provider,
	filepath string,
	password []byte,
) (Client, error) {
	// An empty password is interpreted by badger as no password:
	if password == nil {
		password = []byte{}
	}

	db, err := badger.Open(
		badger.DefaultOptions(filepath).

			// The default badger logger is very noisy so
			// we are forcing it to stay silent:
			WithLogger(nil).

			// The deffault value for logFileSize was 1GB, so by setting it
			// to 100MB we expect it to use less disk space:
			WithValueLogFileSize(100 * MB).
			WithEncryptionKey(password).
			// Using an index cache is recommended when using a password, so we always use a 50MB one:
			WithIndexCacheSize(50 << 20),
	)
	if err != nil {
		return Client{}, fmt.Errorf("unable to connect to badger on path '%s': %w", filepath, err)
	}

	gcWorkerDoneCh := make(chan struct{})

	var g errgroup.Group
	g.Go(func() error {
		// Garbage collects every 25 minutes:
		garbageCollectorWorker(db, gcWorkerDoneCh, 25*time.Minute)
		return nil
	})

	return Client{
		// logger: logger,

		db:     db,
		g:      &g,
		doneCh: gcWorkerDoneCh,

		isClosed: &atomic.Bool{},
	}, nil
}

func (c Client) Close() error {
	if c.isClosed.Load() {
		return nil
	}
	defer func() {
		c.isClosed.Store(true)
	}()

	// This tells the garbage collector to stop:
	close(c.doneCh)

	err := errors.Join(
		c.g.Wait(),
		c.db.Close(),
	)
	if err != nil {
		return fmt.Errorf("error closing badger Client: %w", err)
	}
	return nil
}

// garbageCollectorWorker was copied from badger docs:
//
// - https://dgraph.io/docs/badger/get-started/#garbage-collection
//
// there it is recommended to run it when the workload is low on our app,
// but since we will not have big workloads anyway I am just running it periodically.
func garbageCollectorWorker(db *badger.DB, doneCh chan struct{}, gcInterval time.Duration) {
	ticker := time.NewTicker(gcInterval)
	defer ticker.Stop()
	for {
		select {
		case <-doneCh:
			return
		case <-ticker.C:
		}

		// This loops is meant to run until we get an error saying
		// no garbage collection was performed:
		var err error
		for err == nil {
			// This input is a threashold: It calculates how much space will be gained
			// if a particular section is garbage collected and if it is less than 50%
			// of the space used it ignores it, this number must be between 0 and 1.
			err = db.RunValueLogGC(0.5)
		}
	}
}

func (c Client) Set(ctx context.Context, key string, value string) error {
	return c.db.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(key), []byte(value))
		if err != nil {
			return fmt.Errorf("error saving '%s' to db: %w", key, err)
		}
		return nil
	})
}

func (c Client) Get(ctx context.Context, key string) (value string, _ error) {
	return value, c.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err == badger.ErrKeyNotFound {
			return internal.ErrRecordNotFound
		} else if err != nil {
			return fmt.Errorf("unexpected error reading key '%s' from db: %w", key, err)
		}

		err = item.Value(func(val []byte) error {
			value = string(val)
			return nil
		})
		if err != nil {
			return fmt.Errorf("error parsing the value for key '%s' from db: %w", key, err)
		}

		return nil
	})
}

func (c Client) List(ctx context.Context, prefixStr string) (keys []string, _ error) {
	prefix := []byte(prefixStr)

	return keys, c.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			keys = append(keys, string(item.Key()))
		}
		return nil
	})
}

func (c Client) Find(ctx context.Context, prefixStr string) (kv []internal.KeyValue, _ error) {
	prefix := []byte(prefixStr)

	return kv, c.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			k := item.Key()
			err := item.Value(func(v []byte) error {
				kv = append(kv, internal.KeyValue{
					Key:   string(k),
					Value: string(v),
				})
				return nil
			})
			if err != nil {
				return fmt.Errorf("error reading value for key: '%s': %w", k, err)
			}
		}
		return nil
	})
}

func (c Client) Delete(ctx context.Context, key string) error {
	return c.db.Update(func(txn *badger.Txn) error {
		err := txn.Delete([]byte(key))
		if err == badger.ErrKeyNotFound {
			return internal.ErrRecordNotFound
		} else if err != nil {
			return fmt.Errorf("error deleting '%s' from db: %w", key, err)
		}
		return nil
	})
}
