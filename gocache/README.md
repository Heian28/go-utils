# GoCache

A Redis cache manager for Go with MessagePack serialization for fast and compact data storage.

## Features

- **MessagePack encoding** - Faster and smaller payload compared to JSON
- **Pattern-based operations** - Query and delete keys by pattern (e.g. `user:*`)
- **Upsert support** - Update existing data with or without replacing the TTL
- **Configurable connection pool** - Tune pool size, timeouts, and retries
- **Built-in logging** - Integrates with `gologger` or any `logrus.Logger`

## Installation

```bash
go get github.com/Heian28/go-utils/gocache
```

## Configuration

```go
type GoCacheConfig struct {
    URI          string        // Redis address (e.g. "localhost:6379")
    User         string        // Redis username (optional)
    Password     string        // Redis password (optional)
    Database     int           // Redis database index
    PoolSize     int           // Connection pool size
    MinIdleConns int           // Minimum idle connections
    MaxRetries   int           // Maximum command retries
    ReadTimeout  time.Duration // Read timeout
    WriteTimeout time.Duration // Write timeout
    PoolTimeout  time.Duration // Pool wait timeout
}
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/Heian28/go-utils/gocache"
)

func main() {
    cache := gocache.New(gocache.GoCacheConfig{
        URI:      "localhost:6379",
        Database: 0,
    }, nil)

    ctx := context.Background()

    // Save a value with 5 minute TTL
    err := cache.Save(ctx, "user:1", map[string]string{
        "name":  "John",
        "email": "john@example.com",
    }, 5*time.Minute)
    if err != nil {
        panic(err)
    }

    // Get the value back
    var result map[string]string
    err = cache.Get(ctx, "user:1", &result)
    if err != nil {
        panic(err)
    }
    fmt.Println(result) // map[email:john@example.com name:John]
}
```

## API Reference

### `New(conf GoCacheConfig, log *logrus.Logger) GoCacheManager`

Creates a new cache manager instance. Pass `nil` for log to use the default `gologger`.

### `Save(ctx, key, value, duration) error`

Stores a value with the given TTL. The value is serialized using MessagePack.

```go
type User struct {
    Name  string
    Email string
}

err := cache.Save(ctx, "user:1", User{Name: "John", Email: "john@example.com"}, 10*time.Minute)
```

### `Get(ctx, key, output) error`

Retrieves a value by key. The `output` parameter must be a pointer.

```go
var user User
err := cache.Get(ctx, "user:1", &user)
```

### `Delete(ctx, key) error`

Deletes a single key.

```go
err := cache.Delete(ctx, "user:1")
```

### `GetByPattern(ctx, pattern, output) error`

Retrieves all values matching a Redis key pattern. Uses `SCAN` to iterate keys and `MGET` to fetch values.

```go
var users []User
err := cache.GetByPattern(ctx, "user:*", &users)
```

### `DeleteByPattern(ctx, pattern, batch) error`

Deletes all keys matching a pattern. Keys are scanned in batches of the given size.

```go
err := cache.DeleteByPattern(ctx, "session:*", 100)
```

### `Upsert(ctx, key, value, duration) error`

Updates an existing key's value. The `duration` parameter is a pointer:

- `&duration` - Replace both the data and the TTL
- `nil` - Update the data only, preserving the existing TTL

```go
// Update data only, keep existing TTL
err := cache.Upsert(ctx, "user:1", User{Name: "Jane", Email: "jane@example.com"}, nil)

// Update data and set new TTL
newTTL := 15 * time.Minute
err := cache.Upsert(ctx, "user:1", User{Name: "Jane", Email: "jane@example.com"}, &newTTL)
```

## Usage with Custom Logger

```go
import "github.com/sirupsen/logrus"

log := logrus.New()
log.SetLevel(logrus.DebugLevel)

cache := gocache.New(gocache.GoCacheConfig{
    URI:      "localhost:6379",
    Database: 0,
}, log)
```
