# Go Utils

A collection of utility packages for Go application development covering databases, HTTP frameworks, logging, messaging, storage, and various other tools.

## 📋 Table of Contents

-   [Features](#features)
-   [Installation](#installation)
-   [Available Packages](#available-packages)
-   [Usage](#usage)
-   [Contributing](#contributing)

## ✨ Features

-   **Database**: PostgreSQL and MongoDB support with GORM
-   **HTTP Framework**: Fiber framework utilities (error handling, response, middleware)
-   **Logging**: Structured logging with Logrus and file rotation
-   **Storage**: AWS S3 and MinIO clients
-   **Messaging**: RabbitMQ publisher and consumer
-   **Cache**: Redis client with JSON serialization
-   **Email**: SMTP email sender with template support
-   **Encryption**: AES encryption/decryption utilities
-   **Testing**: Test utilities for development

## 🚀 Installation

```bash
go get github.com/Heian28/go-utils
```

## 📦 Available Packages

### Database

#### PostgreSQL (`db/gopostgres`)

Package for PostgreSQL database connection and management using GORM.

**Features:**

-   Configurable connection pooling
-   Retry mechanism for connections
-   Query logging with slow query detection
-   Transaction support
-   Multiple connection support

**Usage Example:**

```go
import "github.com/Heian28/go-utils/db/gopostgres"

config := gopostgres.GoPostgresConfiguration{
    Host:                  "localhost",
    Port:                  "5432",
    User:                  "postgres",
    Password:              "password",
    Database:              "mydb",
    SSLMode:               "disable",
    MaximumIdleConnection: 10,
    MaximumOpenConnection: 100,
    ConnectionMaxLifeTime: 1 * time.Hour,
    SlowSqlThreshold:      2 * time.Second,
    EnableQueryLogging:    true,
    Retries:               3,
    RetryInterval:         5 * time.Second,
}

db := gopostgres.New(false, config, nil)
defer db.Close()

// Using GORM
gormDB := db.Database()

// Using standard SQL
sqlDB := db.SQL()
```

**Transaction:**

```go
import "github.com/Heian28/go-utils/db/gopostgres"

trx := gopostgres.NewGoPostgresTransaction(db.Database())

err := trx.WithTransaction(ctx, func(tx *gorm.DB) error {
    // Database operations here
    return nil
})
```

#### MongoDB (`db/gomongo`)

Package for MongoDB connection (under development).

### HTTP Framework (Fiber)

#### Error Handling (`fiber/goerror`)

Package for handling errors with consistent formatting.

**Usage Example:**

```go
import "github.com/Heian28/go-utils/fiber/goerror"

// Setup error handler
app.Use(goerror.NewErrorHandler(log))

// In handler
return goerror.ComposeClientError(goerror.BAD_REQUEST, errors.New("invalid input"))
```

#### Response (`fiber/goresponse`)

Package for creating consistent HTTP responses with pagination support.

**Usage Example:**

```go
import "github.com/Heian28/go-utils/fiber/goresponse"

response := goresponse.NewGoResponseClient()

// Response with data
response.Jsonify(c, fiber.StatusOK, "Success", data, nil)

// Response with pagination
meta := response.CreateMeta(page, perPage, total)
response.Jsonify(c, fiber.StatusOK, "Success", data, meta)
```

#### Middleware (`fiber/gomiddleware`)

**Request ID Middleware:**

```go
import "github.com/Heian28/go-utils/fiber/gomiddleware"

app.Use(gomiddleware.RequestID())
```

**HTTP Logger Middleware:**

```go
import "github.com/Heian28/go-utils/fiber/gomiddleware"

app.Use(gomiddleware.HttpLogger(log))
```

### Logging (`gologger`)

Package for structured logging with Logrus and file rotation.

**Features:**

-   JSON formatted logs
-   File rotation with Lumberjack
-   Colorized console output
-   Production/Development mode
-   Custom fields support

**Usage Example:**

```go
import "github.com/Heian28/go-utils/gologger"

gologger.New(
    gologger.SetIsProduction(false),
    gologger.SetServiceName("My Service"),
    gologger.SetPrettyPrint(true),
    gologger.SetFields(map[string]any{
        "version": "1.0.0",
    }),
)

// Using logger
gologger.Logger.Info("Application started")
gologger.Logger.Error("Error occurred")
```

**HTTP Logger:**

```go
import "github.com/Heian28/go-utils/gologger"

gologger.CreateLogger(c, "request", log, nil, nil, nil)
gologger.CreateLogger(c, "response", log, nil, nil, nil)
gologger.CreateLogger(c, "error", log, httpErr, errorMsg, errorStack)
```

### Storage

#### AWS S3 (`aws/gos3`)

Package for uploading and managing files on AWS S3.

**Usage Example:**

```go
import "github.com/Heian28/go-utils/aws/gos3"

config := gos3.GoS3Config{
    Region:          "us-east-1",
    BaseEndpoint:    "https://s3.amazonaws.com",
    AccessKeyID:     "your-access-key",
    SecretAccessKey: "your-secret-key",
    ProjectName:     "my-project",
}

gos3.InitGoS3(ctx, config, log)

// Upload file
result, err := gos3.GoS3Client.UploadFile(ctx, "bucket-name", "file-key", file)
```

#### MinIO (`gominio`)

Package for uploading and managing files on MinIO.

**Usage Example:**

```go
import "github.com/Heian28/go-utils/gominio"

config := gominio.GoMinioConfig{
    Endpoint:    "localhost:9000",
    AccessKey:   "minioadmin",
    SecretKey:   "minioadmin",
    Location:    "us-east-1",
    ProjectName: "my-project",
    UseSSL:      false,
}

gominio.InitGoMinio(config, log)

// Upload file
result, err := gominio.GoMinioClient.UploadFile(
    ctx,
    "location",
    "file-key",
    fileBuffer,
    fileSize,
    "image/jpeg",
)

// Extract file info from multipart
fileInfo, err := gominio.GoMinioClient.MExtractFileInfo(fileHeader)
```

### Messaging

#### RabbitMQ (`gorabbit`)

Package for publishing and consuming messages from RabbitMQ with encryption support.

**Features:**

-   Publisher with retry mechanism
-   Consumer with panic recovery
-   Message encryption (AES)
-   Topic-based routing
-   Multiple queue support

**Usage Example:**

```go
import "github.com/Heian28/go-utils/gorabbit"

config := gorabbit.GoRabbitConfiguration{
    Host:     "localhost",
    Port:     "5672",
    User:     "guest",
    Password: "guest",
    Secret:   "your-encryption-secret-min-24-chars",
    Debug:    false,
}

rabbit := gorabbit.New(config, log)
defer rabbit.Close()

// Publish message
publisher := rabbit.Publisher()
err := publisher.Publish(ctx, gorabbit.GoRabbitPublisherOption{
    Topic:      "user.created",
    Message:    userData,
    Retries:    3,
    RetryDelay: 2000,
})

// Consume messages
consumers := gorabbit.GoRabbitConsumerMessages{
    "user-queue": {
        "user.created": gorabbit.GoRabbitConsumer{
            Consume: func(msg amqp091.Delivery) error {
                // Process message
                return nil
            },
        },
    },
}

rabbit.Listen(consumers)
```

### Cache

#### Redis (`goredis`)

Package for cache operations using Redis with JSON serialization.

**Usage Example:**

```go
import "github.com/Heian28/go-utils/goredis"

config := goredis.GoRedisConfig{
    Addr:     "localhost:6379",
    Password: "",
    Database: 0,
}

redis := goredis.New(config, log, false)

// Save data
err := redis.Save(ctx, "user:123", userData, 1*time.Hour)

// Get data
var userData User
err := redis.Get(ctx, "user:123", &userData)

// Delete data
err := redis.Delete(ctx, "user:123")

// Delete by pattern
err := redis.DeleteByPattern(ctx, "user:*", 100)
```

### Email (`gomail`)

Package for sending emails via SMTP with template support.

**Usage Example:**

```go
import "github.com/Heian28/go-utils/gomail"

config := gomail.GoMailConfig{
    Host:         "smtp.gmail.com",
    Port:         587,
    Username:     "your-email@gmail.com",
    Password:     "your-password",
    From:         "your-email@gmail.com",
    RootTemplate: "./templates",
}

gomail.Init(config)

// Create email
email := gomail.NewEmail(
    "Subject",
    []string{"recipient@example.com"},
    "sender@example.com",
    []string{},
)

// Load template
err := email.LoadTemplate("welcome.html", templateData)

// Send email
err := email.SendMail("")
```

### Encryption (`goencrypt`)

Package for encrypting and decrypting data using AES.

**Usage Example:**

```go
import "github.com/Heian28/go-utils/goencrypt"

encryptor := goencrypt.New("your-secret-key-min-24-chars")

// Encrypt
encrypted, err := encryptor.Encrypt(data)

// Decrypt
decrypted, err := encryptor.Decrypt(encrypted)
```

### Testing (`testutil`)

Package utilities for testing (under development).

## 🔧 Configuration

### Environment Variables

Some packages require configuration through structs. Make sure to configure them correctly according to your application needs.

### Logging Configuration

Logs will be saved in the `logs/` directory with the following format:

-   Format: `app-YYYY-MM-DD.log`
-   Rotation: Max 10MB per file
-   Retention: 28 days
-   Backup: 3 files
-   Compression: Enabled

## 📝 Complete Example

Here is an example of using multiple packages together:

```go
package main

import (
    "context"
    "time"

    "github.com/Heian28/go-utils/gologger"
    "github.com/Heian28/go-utils/db/gopostgres"
    "github.com/Heian28/go-utils/goredis"
    "github.com/Heian28/go-utils/fiber/goerror"
    "github.com/Heian28/go-utils/fiber/goresponse"
    "github.com/Heian28/go-utils/fiber/gomiddleware"
    "github.com/gofiber/fiber/v3"
)

func main() {
    // Initialize logger
    gologger.New(
        gologger.SetIsProduction(false),
        gologger.SetServiceName("My API"),
    )

    // Initialize database
    dbConfig := gopostgres.GoPostgresConfiguration{
        Host:     "localhost",
        Port:     "5432",
        User:     "postgres",
        Password: "password",
        Database: "mydb",
    }
    db := gopostgres.New(false, dbConfig, gologger.Logger)
    defer db.Close()

    // Initialize Redis
    redisConfig := goredis.GoRedisConfig{
        Addr:     "localhost:6379",
        Database: 0,
    }
    redis := goredis.New(redisConfig, gologger.Logger, false)

    // Initialize Fiber app
    app := fiber.New()

    // Middleware
    app.Use(gomiddleware.RequestID())
    app.Use(gomiddleware.HttpLogger(gologger.Logger))

    // Error handler
    app.Use(goerror.NewErrorHandler(gologger.Logger))

    // Routes
    app.Get("/", func(c fiber.Ctx) error {
        response := goresponse.NewGoResponseClient()
        return response.Jsonify(c, fiber.StatusOK, "Hello World", nil, nil)
    })

    app.Listen(":3000")
}
```

## 🤝 Contributing

Contributions are welcome! Please create an issue or pull request for:

-   Bug fixes
-   New features
-   Documentation improvements
-   Performance optimizations

## 📄 License

This repository uses the license determined by the repository owner.

## 🔗 Dependencies

This repository uses several popular libraries:

-   [GORM](https://gorm.io/) - ORM for database
-   [Fiber](https://gofiber.io/) - Web framework
-   [Logrus](https://github.com/sirupsen/logrus) - Structured logger
-   [Redis Go Client](https://github.com/redis/go-redis) - Redis client
-   [RabbitMQ Go Client](https://github.com/rabbitmq/amqp091-go) - RabbitMQ client
-   [MinIO Go Client](https://github.com/minio/minio-go) - MinIO client
-   [AWS SDK Go v2](https://github.com/aws/aws-sdk-go-v2) - AWS SDK

---

**Note**: Some packages are still under development. Please check the documentation or source code for more complete implementation details.
