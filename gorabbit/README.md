# GoRabbit

A RabbitMQ wrapper for Go built on top of [amqp091-go](https://github.com/rabbitmq/amqp091-go) with support for:

- Topic exchange routing
- Automatic publish retry
- End-to-end message encryption using **AES-256-GCM**
- Panic recovery with auto-reconnect on consumers
- Integrated logger via `gologger` / `logrus`

---

## Installation

```bash
go get github.com/Heian28/go-utils/gorabbit
```

---

## Configuration

```go
type GoRabbitConfiguration struct {
    Host     string // RabbitMQ hostname, e.g. "localhost"
    Port     string // RabbitMQ port, e.g. "5672"
    User     string // username
    Password string // password
    Secret   string // optional — enables AES-GCM encryption when set (min. 24 characters)
    Debug    bool   // if true, message body is printed to the log
}
```

> **Encryption note:** When `Secret` is provided, all messages are automatically encrypted using AES-GCM. Both publisher and consumer **must use the same Secret**. When omitted, messages are sent as plain JSON.

---

## Usage

### 1. Initialize

```go
import "github.com/Heian28/go-utils/gorabbit"

rb := gorabbit.New(gorabbit.GoRabbitConfiguration{
    Host:     "localhost",
    Port:     "5672",
    User:     "guest",
    Password: "guest",
}, nil) // nil = use default logger

defer rb.Close()
```

---

### 2. Publish a Message

Get a publisher instance via `rb.Publisher()`, then call `Publish`.

```go
type OrderCreatedEvent struct {
    OrderID string `json:"order_id"`
    Amount  int    `json:"amount"`
}

err := rb.Publisher().Publish(ctx, gorabbit.GoRabbitPublisherOption{
    Topic:   "order.created",
    Message: OrderCreatedEvent{
        OrderID: "ORD-001",
        Amount:  150000,
    },
})
if err != nil {
    log.Fatal(err)
}
```

#### Publish Options

| Field        | Type   | Default | Description                                      |
|--------------|--------|---------|--------------------------------------------------|
| `Topic`      | string | —       | Routing key (required)                           |
| `Message`    | any    | —       | Payload, marshalled to JSON (required)           |
| `Retries`    | int    | 1       | Maximum number of publish attempts               |
| `RetryDelay` | int    | 2000    | Delay between retries in milliseconds            |
| `UserId`     | string | —       | Optional metadata                                |
| `AppId`      | string | —       | Optional metadata                                |

**Example with retry:**

```go
err := rb.Publisher().Publish(ctx, gorabbit.GoRabbitPublisherOption{
    Topic:      "payment.processed",
    Message:    payload,
    Retries:    3,
    RetryDelay: 1000, // 1 second
})
```

---

### 3. Consume Messages

`Listen` accepts a nested map: `queue name → topic → handler`.

A single queue can listen to multiple topics at once.

```go
import "github.com/rabbitmq/amqp091-go"

rb.Listen(gorabbit.GoRabbitConsumerMessages{
    "my-service-queue": {
        "order.created": gorabbit.GoRabbitConsumer{
            Consume: func(msg amqp091.Delivery) error {
                var event OrderCreatedEvent
                if err := json.Unmarshal(msg.Body, &event); err != nil {
                    return err
                }
                fmt.Printf("Order received: %s\n", event.OrderID)
                return nil
            },
        },
        "order.cancelled": gorabbit.GoRabbitConsumer{
            Consume: func(msg amqp091.Delivery) error {
                fmt.Printf("Order cancelled, message ID: %s\n", msg.MessageId)
                return nil
            },
        },
    },
})
```

> `Listen` runs blocking inside an internal goroutine. Call it at the end of your program or inside a separate goroutine.

---

## Scenario: Publisher & Consumer in the Same Instance

Suitable for a monolith or a single binary that both publishes and consumes.

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "os"
    "os/signal"

    "github.com/Heian28/go-utils/gorabbit"
    "github.com/rabbitmq/amqp091-go"
)

type NotifEvent struct {
    UserID  string `json:"user_id"`
    Message string `json:"message"`
}

func main() {
    rb := gorabbit.New(gorabbit.GoRabbitConfiguration{
        Host:     "localhost",
        Port:     "5672",
        User:     "guest",
        Password: "guest",
    }, nil)
    defer rb.Close()

    // Consumer runs in a background goroutine
    go rb.Listen(gorabbit.GoRabbitConsumerMessages{
        "notif-queue": {
            "notif.send": gorabbit.GoRabbitConsumer{
                Consume: func(msg amqp091.Delivery) error {
                    var event NotifEvent
                    json.Unmarshal(msg.Body, &event)
                    fmt.Printf("[Consumer] Send notification to user %s: %s\n", event.UserID, event.Message)
                    return nil
                },
            },
        },
    })

    // Publisher runs in the main goroutine
    ctx := context.Background()
    rb.Publisher().Publish(ctx, gorabbit.GoRabbitPublisherOption{
        Topic: "notif.send",
        Message: NotifEvent{
            UserID:  "USR-123",
            Message: "Your order has been shipped!",
        },
    })

    quit := make(chan os.Signal, 1)
    signal.Notify(quit, os.Interrupt)
    <-quit
}
```

---

## Scenario: Publisher & Consumer in Separate Instances (Microservices)

This is the most common scenario. **When encryption is enabled, both services must use the same `Secret`.**

### Service A — Publisher (e.g. `order-service`)

```go
package main

import (
    "context"
    "os"
    "os/signal"

    "github.com/Heian28/go-utils/gorabbit"
)

type OrderCreatedEvent struct {
    OrderID    string `json:"order_id"`
    CustomerID string `json:"customer_id"`
    Total      int64  `json:"total"`
}

func main() {
    rb := gorabbit.New(gorabbit.GoRabbitConfiguration{
        Host:     "rabbitmq.internal",
        Port:     "5672",
        User:     "admin",
        Password: "secret",
        Secret:   "this-is-a-shared-secret-key-xyz!", // min 24 chars, must match across all services
    }, nil)
    defer rb.Close()

    ctx := context.Background()

    err := rb.Publisher().Publish(ctx, gorabbit.GoRabbitPublisherOption{
        Topic: "order.created",
        Message: OrderCreatedEvent{
            OrderID:    "ORD-999",
            CustomerID: "CUST-42",
            Total:      250000,
        },
        Retries:    3,
        RetryDelay: 500,
    })
    if err != nil {
        panic(err)
    }

    quit := make(chan os.Signal, 1)
    signal.Notify(quit, os.Interrupt)
    <-quit
}
```

### Service B — Consumer (e.g. `notif-service`)

```go
package main

import (
    "encoding/json"
    "fmt"
    "os"
    "os/signal"

    "github.com/Heian28/go-utils/gorabbit"
    "github.com/rabbitmq/amqp091-go"
)

type OrderCreatedEvent struct {
    OrderID    string `json:"order_id"`
    CustomerID string `json:"customer_id"`
    Total      int64  `json:"total"`
}

func main() {
    rb := gorabbit.New(gorabbit.GoRabbitConfiguration{
        Host:     "rabbitmq.internal",
        Port:     "5672",
        User:     "admin",
        Password: "secret",
        Secret:   "this-is-a-shared-secret-key-xyz!", // must match the publisher's Secret
    }, nil)
    defer rb.Close()

    rb.Listen(gorabbit.GoRabbitConsumerMessages{
        "notif-service-queue": {
            "order.created": gorabbit.GoRabbitConsumer{
                Consume: func(msg amqp091.Delivery) error {
                    var event OrderCreatedEvent
                    if err := json.Unmarshal(msg.Body, &event); err != nil {
                        return err
                    }
                    fmt.Printf(
                        "[Consumer] Order %s from customer %s for $%d received\n",
                        event.OrderID, event.CustomerID, event.Total,
                    )
                    return nil
                },
            },
        },
    })

    quit := make(chan os.Signal, 1)
    signal.Notify(quit, os.Interrupt)
    <-quit
}
```

> Because encryption uses **AES-GCM**, a random nonce is generated per message and sent alongside the ciphertext. There is no encryption state to share between instances — only the `Secret` needs to match.

---

## Topic & Queue Structure

GoRabbit uses a single **topic exchange** named `"exchange"`. Routing is handled via routing keys (topics).

```
Publisher                     RabbitMQ                      Consumer
   │                             │                              │
   │── topic: "order.created" ──▶  exchange (topic)            │
   │                             │── routing key match ───────▶ queue "notif-service-queue"
   │                             │                              │── handler "order.created"
```

A single queue can be bound to multiple topics:

```go
"inventory-queue": {
    "order.created":   gorabbit.GoRabbitConsumer{Consume: handleOrderCreated},
    "order.cancelled": gorabbit.GoRabbitConsumer{Consume: handleOrderCancelled},
    "stock.updated":   gorabbit.GoRabbitConsumer{Consume: handleStockUpdated},
},
```

---

## AES-GCM Encryption

When `Secret` is provided in the configuration:

- Each message is encrypted with **AES-256-GCM** (authenticated encryption)
- A 12-byte nonce is randomly generated per message and prepended to the ciphertext
- The output is **Base64**-encoded before being sent to RabbitMQ
- The consumer automatically decrypts messages before invoking the handler
- Messages that fail decryption (e.g. wrong Secret or corrupted data) are skipped with an error log

```
JSON Plaintext → AES-GCM Encrypt (random nonce) → Base64 → RabbitMQ → Base64 Decode → AES-GCM Decrypt → JSON to handler
```

---

## Auto-Reconnect

If a consumer goroutine panics, GoRabbit will:

1. Log the panic error along with the message ID and topic
2. Wait 5 seconds
3. Automatically call `Listen` again with the same consumers

---

## Closing the Connection

Always call `Close()` when the application shuts down to gracefully close the channel and connection:

```go
rb := gorabbit.New(conf, nil)
defer rb.Close()
```
