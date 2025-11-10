package gorabbit

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Heian28/go-utils/gologger"

	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

// ConsumerFunc is a type that represents the consumer function.
// It is used to represent the consumer function.
type ConsumerFunc func(msg amqp091.Delivery) error

// GoRabbitConsumer is a struct that represents the consumer.
// It is used to represent the consumer.
type GoRabbitConsumer struct {
	Consume ConsumerFunc
}

// GoRabbitPublisherOption is a struct that represents the publisher option.
// It is used to represent the publisher option.
type GoRabbitPublisherOption struct {
	Topic      string
	Message    any
	UserId     string
	AppId      string
	Retries    int
	RetryDelay int
}

// Publisher is an interface that defines the methods for the publisher.
// It is used to define the methods for the publisher.
type Publisher interface {
	Publish(ctx context.Context, opt GoRabbitPublisherOption) error
}

// GoRabbitConsumerMessages is a type that represents the consumer messages.
// It is used to represent the consumer messages.
type GoRabbitConsumerMessages map[string]map[string]GoRabbitConsumer

// GoRabbit is an interface that defines the methods for the GoRabbit.
// It is used to define the methods for the GoRabbit.
type GoRabbit interface {
	Publisher() Publisher
	Listen(consumers GoRabbitConsumerMessages)
	Close()
}

// rbt is a struct that represents the rabbitmq.
// It is used to represent the rabbitmq.
type rbt struct {
	acn                   *amqp091.Connection
	ach                   *amqp091.Channel
	log                   *logrus.Logger
	withMessageEncryption bool
	cr                    *crypto
	conf                  GoRabbitConfiguration
}

// Publish is a function that publishes the message.
// It takes a context and a GoRabbitPublisherOption and returns an error.
// This is used to publish the message.
func (r *rbt) Publish(
	ctx context.Context,
	opt GoRabbitPublisherOption,
) error {
	r.log.Info("[GoRabbit] Publishing message...")

	if r.trimSpace(opt.Topic) == "" {
		return errors.New("topic is required")
	}

	if opt.Message == nil {
		return errors.New("message is required")
	}

	var (
		sb         = 1
		sbInterval = 2000 * time.Millisecond
		idxProcess = 0
		success    = false
		uid        = uuid.New().String()

		mut sync.Mutex
		err error
	)

	if opt.Retries > 0 {
		sb = opt.Retries
	}

	if opt.RetryDelay > 0 {
		sbInterval = time.Duration(opt.RetryDelay) * time.Millisecond
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	for ok := true; ok; ok = idxProcess < sb && !success {
		mut.Lock()

		r.log.Infof(
			"[GoRabbit] [%d] [%s] Publishing topic: '%s'",
			idxProcess,
			uid,
			opt.Topic,
		)
		if r.conf.Debug {
			r.log.Debug(opt.Message)
		}

		var (
			msg  []byte
			err2 error
		)

		if r.withMessageEncryption {
			msg, err2 = r.cr.encrypt(opt.Message)
			if err2 != nil {
				r.log.Errorf(
					"[GoRabbit] [%d] [%s] Error encrypting message: %s",
					idxProcess,
					uid,
					err2.Error(),
				)
				idxProcess++
				err = err2
				mut.Unlock()
				time.Sleep(sbInterval)
				continue
			}
		} else {
			msg, err2 = r.cr.toBytes(opt.Message)
			if err2 != nil {
				r.log.Errorf(
					"[GoRabbit] [%d] [%s] Error converting message to bytes: %s",
					idxProcess,
					uid,
					err2.Error(),
				)
				idxProcess++
				err = err2
				mut.Unlock()
				time.Sleep(sbInterval)
				continue
			}
		}

		err2 = r.ach.PublishWithContext(
			ctx,
			"exchange",
			opt.Topic,
			false,
			false,
			amqp091.Publishing{
				ContentType: "text/plain",
				MessageId:   uid,
				Body:        msg,
			},
		)
		if err2 != nil {
			r.log.Errorf(
				"[GoRabbit] [%d] [%s] Error publishing message: %s",
				idxProcess,
				uid,
				err2.Error(),
			)
			idxProcess++
			err = err2
			mut.Unlock()
			time.Sleep(sbInterval)
			continue
		}

		success = true
		mut.Unlock()
	}

	if err != nil && !success {
		r.log.Errorf(
			"[GoRabbit] [%s] Error publishing message. Attempts: %d/%d",
			uid,
			idxProcess,
			sb,
		)
	} else {
		r.log.Infof(
			"[GoRabbit] [%s] Message published successfully",
			uid,
		)
	}

	return nil
}

// Listen is a function that listens to the messages.
// It takes a GoRabbitConsumerMessages and returns nothing.
// This is used to listen to the messages.
func (r *rbt) Listen(consumers GoRabbitConsumerMessages) {
	exc := "exchange"

	if err := r.ach.ExchangeDeclare(
		exc,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		r.log.Fatalf("[GoRabbit] Error declaring exchange: %s", err.Error())
	}

	for queue := range consumers {
		q, err := r.ach.QueueDeclare(
			queue,
			true,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			r.log.Fatalf("[GoRabbit] Error declaring queue: '%s'", err.Error())
		} else {
			r.log.Infof("[GoRabbit] Group queue declared: '%s'", q.Name)
		}

		topics := consumers[queue]
		for topic := range topics {
			if err := r.ach.QueueBind(
				q.Name,
				topic,
				"exchange",
				false,
				nil,
			); err != nil {
				r.log.Fatalf(
					"[GoRabbit] Error binding queue: '%s' -> '%s'",
					q.Name,
					topic,
				)
			} else {
				r.log.Infof(
					"[GoRabbit] Group queue binded: '%s' -> '%s'",
					q.Name,
					topic,
				)
			}
		}

		messages, err := r.ach.Consume(
			q.Name,
			"",
			true,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			r.log.Fatalf("[GoRabbit] Error consuming queue: '%s'", err.Error())
		}

		go r.consumeMessages(consumers, queue, messages)

		r.log.Infof("[GoRabbit] Queue '%s' started", queue)
	}
}

// consumeMessages is a function that consumes the messages.
// It takes a GoRabbitConsumerMessages, a string, and a channel of amqp091.Delivery and returns nothing.
// This is used to consume the messages.
func (r *rbt) consumeMessages(
	consumers GoRabbitConsumerMessages,
	queue string,
	msgs <-chan amqp091.Delivery,
) {
	var (
		topic     string
		messageId string
		mu        sync.Mutex

		delay = 5000 * time.Millisecond
	)

	defer func(id *string, t *string) {
		if rc := recover(); rc != nil {
			r.log.Errorf(
				"[GoRabbit] [%s] [%s] Consumer panic: %v",
				*id,
				*t,
				rc,
			)
			r.log.Info("[GoRabbit] Rejoin rabbitmq...")
			time.Sleep(delay)
			r.Listen(consumers)
		}
	}(&messageId, &topic)

	for msg := range msgs {
		mu.Lock()

		messageId = msg.MessageId
		topic = msg.RoutingKey

		r.log.Infof(
			"[GoRabbit] [%s] [%s] Consuming topic...",
			messageId,
			topic,
		)

		qu := consumers[queue]

		if _, ok := qu[topic]; !ok {
			r.log.Errorf(
				"[GoRabbit] [%s] [%s] Consumer not found",
				messageId,
				topic,
			)
			mu.Unlock()
			continue
		}

		if r.withMessageEncryption {
			decrypted, err := r.cr.decrypt(string(msg.Body))
			if err != nil {
				r.log.Errorf(
					"[GoRabbit] [%s] [%s] Error decrypting message: %s",
					messageId,
					topic,
					err.Error(),
				)
				mu.Unlock()
				continue
			}
			msg.Body = decrypted
		}

		if r.conf.Debug {
			r.log.Info(string(msg.Body))
		}

		if err := qu[topic].Consume(msg); err != nil {
			r.log.Errorf(
				"[GoRabbit] [%s] [%s] Error consuming message: %s",
				messageId,
				topic,
				err.Error(),
			)
			mu.Unlock()
			continue
		}

		mu.Unlock()
	}
}

// Close is a function that closes the rabbitmq.
// It takes nothing and returns nothing.
// This is used to close the rabbitmq.
func (r *rbt) Close() {
	if r.ach != nil {
		if err := r.ach.Close(); err != nil {
			r.log.Errorf("[GoRabbit] Error closing channel: %s", err.Error())
		} else {
			r.log.Info("[GoRabbit] Channel closed successfully")
		}
	}

	if r.acn != nil {
		if err := r.acn.Close(); err != nil {
			r.log.Errorf("[GoRabbit] Error closing connection: %s", err.Error())
		} else {
			r.log.Info("[GoRabbit] Connection closed successfully")
		}
	}
}

// trimSpace is a function that trims the space from the string.
// It takes a string and returns a string.
// This is used to trim the space from the string.
func (r *rbt) trimSpace(s string) string {
	return strings.TrimSpace(s)
}

// Publisher is a function that returns the publisher.
// It takes nothing and returns a Publisher.
// This is used to return the publisher.
func (r *rbt) Publisher() Publisher {
	return r
}

// GoRabbitConfiguration is a struct that represents the configuration for the GoRabbit.
// It is used to represent the configuration for the GoRabbit.
type GoRabbitConfiguration struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Secret   string `mapstructure:"secret"`
	Debug    bool   `mapstructure:"debug"`
}

// New is a function that creates a new GoRabbit.
// It takes a GoRabbitConfiguration and a pointer to a logrus.Logger and returns a GoRabbit.
// This is used to create a new GoRabbit.
func New(
	opt GoRabbitConfiguration,
	log *logrus.Logger,
) GoRabbit {
	if log == nil {
		gologger.New()
		log = gologger.Logger
	}

	dsn := fmt.Sprintf(
		"amqp://%s:%s@%s:%s/",
		opt.User,
		opt.Password,
		opt.Host,
		opt.Port,
	)

	conn, err := amqp091.Dial(dsn)
	if err != nil {
		log.Fatalf("[GoRabbit] Error connecting to RabbitMQ: %s", err.Error())
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("[GoRabbit] Error opening channel: %s", err.Error())
	}

	if len(opt.Secret) > 0 {
		if len(opt.Secret) < 24 {
			log.Fatalf("[GoRabbit] Secret must be at least 24 characters long")
		}
	}

	withMessageEncryption := len(opt.Secret) > 0

	return &rbt{
		acn:                   conn,
		ach:                   ch,
		log:                   log,
		withMessageEncryption: withMessageEncryption,
		cr:                    initCrypto(opt.Secret),
		conf:                  opt,
	}
}

// crypto is a struct that represents the crypto.
// It is used to represent the crypto.
type crypto struct {
	secret string
	size   []byte
}

// toBytes is a function that converts the data to bytes.
// It takes a any and returns a []byte and an error.
// This is used to convert the data to bytes.
func (c *crypto) toBytes(data any) ([]byte, error) {
	return json.Marshal(data)
}

// encode is a function that encodes the data to base64.
// It takes a []byte and returns a string.
// This is used to encode the data to base64.
func (c *crypto) encode(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// decode is a function that decodes the data from base64.
// It takes a string and returns a []byte and an error.
// This is used to decode the data from base64.
func (c *crypto) decode(data string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(data)
}

// getBlock is a function that returns the block.
// It takes nothing and returns a cipher.Block and an error.
// This is used to return the block.
func (c *crypto) getBlock() (cipher.Block, error) {
	return aes.NewCipher([]byte(c.secret))
}

// encrypt is a function that encrypts the message.
// It takes a any and returns a []byte and an error.
// This is used to encrypt the message.
func (c *crypto) encrypt(message any) ([]byte, error) {
	var (
		enc  string
		encB []byte
		err  error
	)

	b, err := c.toBytes(message)
	if err != nil {
		return nil, err
	}

	blk, err := c.getBlock()
	if err != nil {
		return nil, err
	}

	cfb := cipher.NewCFBEncrypter(blk, c.size)
	cText := make([]byte, len(b))
	cfb.XORKeyStream(cText, b)

	enc = c.encode(cText)
	encB = []byte(enc)

	return encB, nil
}

// decrypt is a function that decrypts the message.
// It takes a string and returns a []byte and an error.
// This is used to decrypt the message.
func (c *crypto) decrypt(message string) ([]byte, error) {
	blk, err := c.getBlock()
	if err != nil {
		return nil, err
	}

	cText, err := c.decode(message)
	if err != nil {
		return nil, err
	}

	cfb := cipher.NewCFBDecrypter(blk, c.size)
	plain := make([]byte, len(cText))
	cfb.XORKeyStream(plain, cText)

	return plain, nil
}

// initCrypto is a function that initializes the crypto.
// It takes a string and returns a *crypto.
// This is used to initialize the crypto.
func initCrypto(secret string) *crypto {
	bb := make([]byte, 16)
	rand.Read(bb)

	if len(strings.TrimSpace(secret)) < 24 {
		panic("secret for encryption must be at least 24 characters long")
	}

	return &crypto{secret: secret, size: bb}
}
