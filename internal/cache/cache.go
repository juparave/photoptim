package cache

import (
	"encoding/json"
	"errors"
	"time"

	bolt "go.etcd.io/bbolt"
)

// DirectoryCache stores serialized directory listings with TTL.
type DirectoryCache struct {
	db  *bolt.DB
	ttl time.Duration
}

type record struct {
	StoredAt time.Time       `json:"storedAt"`
	Data     json.RawMessage `json:"data"`
}

// Open initializes / opens the bbolt database file.
func Open(path string, ttl time.Duration) (*DirectoryCache, error) {
	db, err := bolt.Open(path, 0o600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}
	return &DirectoryCache{db: db, ttl: ttl}, nil
}

// Put writes a directory listing.
func (c *DirectoryCache) Put(dir string, value any) error { //nolint:ireturn
	if c == nil {
		return errors.New("cache not initialized")
	}
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.db.Update(func(tx *bolt.Tx) error {
		bkt, err := tx.CreateBucketIfNotExists([]byte("dir"))
		if err != nil {
			return err
		}
		rec := record{StoredAt: time.Now(), Data: b}
		rb, _ := json.Marshal(rec)
		return bkt.Put([]byte(dir), rb)
	})
}

// Get unmarshals cached listing into target if fresh; returns bool fresh.
func (c *DirectoryCache) Get(dir string, target any) (bool, error) { //nolint:ireturn
	if c == nil {
		return false, errors.New("cache not initialized")
	}
	var rb []byte
	if err := c.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte("dir"))
		if bkt == nil {
			return nil
		}
		rb = bkt.Get([]byte(dir))
		return nil
	}); err != nil {
		return false, err
	}
	if rb == nil {
		return false, nil
	}
	var rec record
	if err := json.Unmarshal(rb, &rec); err != nil {
		return false, err
	}
	if time.Since(rec.StoredAt) > c.ttl {
		return false, nil
	}
	return true, json.Unmarshal(rec.Data, target)
}

func (c *DirectoryCache) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}
