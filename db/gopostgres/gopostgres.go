package gopostgres

import (
	"database/sql"
	"errors"
	"fmt"
	golog "log"
	"os"
	"strings"
	"time"

	"github.com/Heian28/go-utils/gologger"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	grmlog "gorm.io/gorm/logger"
)

type GoPostgres interface {
	Database() *gorm.DB
	SQL() *sql.DB
	Close()
}

type GoPostgresConfiguration struct {
	URL                   string        `mapstructure:"url"`
	Host                  string        `mapstructure:"host"`
	Port                  string        `mapstructure:"port"`
	User                  string        `mapstructure:"user"`
	Password              string        `mapstructure:"password"`
	Database              string        `mapstructure:"database"`
	SSLMode               string        `mapstructure:"sslMode"`
	MaximumIdleConnection int           `mapstructure:"maximumIdleConnection"`
	MaximumOpenConnection int           `mapstructure:"maximumOpenConnection"`
	ConnectionMaxLifeTime time.Duration `mapstructure:"connectionMaxLifeTime"`
	SlowSqlThreshold      time.Duration `mapstructure:"slowSqlThreshold"`
	EnableQueryLogging    bool          `mapstructure:"enableQueryLogging"`
	Retries               int           `mapstructure:"retries"`
	RetryInterval         time.Duration `mapstructure:"retryInterval"`
}

type postgre struct {
	db  *gorm.DB
	sql *sql.DB
	log *logrus.Logger
}

func New(
	isProduction bool,
	config GoPostgresConfiguration,
	log *logrus.Logger,
) GoPostgres {
	if log == nil {
		gologger.New(
			gologger.SetIsProduction(isProduction),
			gologger.SetServiceName("GoPostgres Database"),
			gologger.SetPrettyPrint(true),
		)
		log = gologger.Logger
	}

	log.Info("[GoPostgres] Creating database connection...")

	var (
		host     string = "localhost"
		port     string = "5432"
		user     string = "postgres"
		password string = "postgres"
		database string = "postgres"
		sslMode  string = "disable"
	)

	sliceDsn := make([]string, 0)

	if trimString(config.Host) != "" {
		host = trimString(config.Host)
	}

	if trimString(config.Port) != "" {
		port = trimString(config.Port)
	}

	if trimString(config.User) != "" {
		user = trimString(config.User)
	}

	if trimString(config.Password) != "" {
		password = trimString(config.Password)
		sliceDsn = append(sliceDsn, fmt.Sprintf("password=%s", password))
	}

	if trimString(config.Database) != "" {
		database = trimString(config.Database)
	}

	if trimString(config.SSLMode) != "" {
		sslMode = trimString(config.SSLMode)
	}

	sliceDsn = append(sliceDsn, fmt.Sprintf("host=%s", host))
	sliceDsn = append(sliceDsn, fmt.Sprintf("port=%s", port))
	sliceDsn = append(sliceDsn, fmt.Sprintf("user=%s", user))
	sliceDsn = append(sliceDsn, fmt.Sprintf("dbname=%s", database))
	sliceDsn = append(sliceDsn, fmt.Sprintf("sslmode=%s", sslMode))

	var (
		slowThreshold time.Duration = 2 * time.Second
		log_level                   = grmlog.Info
	)

	if isProduction {
		log_level = grmlog.Info
	}

	if config.SlowSqlThreshold > 0 {
		slowThreshold = config.SlowSqlThreshold
	}

	glogger := grmlog.New(
		golog.New(os.Stdout, "\r\n", golog.LstdFlags),
		grmlog.Config{
			SlowThreshold:             slowThreshold,
			LogLevel:                  log_level,
			IgnoreRecordNotFoundError: false,
			Colorful:                  true,
		},
	)

	gormConfig := &gorm.Config{}
	if config.EnableQueryLogging {
		gormConfig.Logger = glogger
	}

	var (
		try      = 0
		success  = false
		retries  = 1
		interval = 5 * time.Second

		pg            *postgre
		errConnection error
	)

	if config.Retries > 0 {
		retries = config.Retries
	}

	if config.RetryInterval > 0 {
		interval = config.RetryInterval
	}

	for ok := true; ok; ok = try < retries && !success {
		dsn := config.URL
		if dsn == "" {
			dsn = strings.Join(sliceDsn, " ")
		}

		db, err := gorm.Open(
			postgres.Open(dsn),
			gormConfig,
		)
		if err != nil {
			if retries > 1 {
				log.Info("[GoPostgres] Retrying connection to database...")
				time.Sleep(interval)
			}
			errConnection = err
			try++
			continue
		}

		sqlDb, err := db.DB()
		if err != nil {
			if retries > 1 {
				log.Info("[GoPostgres] Retrying get underlying *sql.DB...")
				time.Sleep(interval)
			}
			errConnection = err
			try++
			continue
		}

		err = sqlDb.Ping()
		if err != nil {
			if retries > 1 {
				log.Info("[GoPostgres] Retrying ping database...")
				time.Sleep(interval)
			}
			errConnection = err
			try++
			continue
		}

		var (
			maximumIdleConnection = 10
			maximumOpenConnection = 100
			connectionMaxLifeTime = 1 * time.Hour
		)

		if config.MaximumIdleConnection > 0 {
			maximumIdleConnection = config.MaximumIdleConnection
		}

		if config.MaximumOpenConnection > 0 {
			maximumOpenConnection = config.MaximumOpenConnection
		}

		if config.ConnectionMaxLifeTime > 0 {
			connectionMaxLifeTime = config.ConnectionMaxLifeTime
		}

		sqlDb.SetMaxIdleConns(maximumIdleConnection)
		sqlDb.SetMaxOpenConns(maximumOpenConnection)
		sqlDb.SetConnMaxLifetime(connectionMaxLifeTime)

		pg = &postgre{
			db:  db,
			sql: sqlDb,
			log: log,
		}

		success = true
		log.Infof(
			"[GoPostgres] Database connection successful to %s@%s:%s/%s",
			user,
			host,
			port,
			database,
		)
	}

	if !success && errConnection != nil {
		log.Errorf("[GoPostgres] (Attempts: %d/%d) Failed to connect to database", try, retries)
		log.Fatalf("[GoPostgres] Error: %v", errConnection)
	}

	return pg
}

func (p *postgre) Database() *gorm.DB {
	return p.db
}

func (p *postgre) SQL() *sql.DB {
	return p.sql
}

func (p *postgre) Close() {
	if sql := p.sql; sql != nil {
		if err := sql.Close(); err != nil {
			p.log.Errorf("[GoPostgres] Error closing database connection: %v", err)
			os.Exit(1)
		} else {
			p.log.Info("[GoPostgres] Database connection closed successfully")
		}
	}
}

func trimString(s string) string {
	return strings.TrimSpace(s)
}

var GoPostgresConnection map[string]*gorm.DB = make(map[string]*gorm.DB)

func SetupPostgreConnection(connectionName string, db *gorm.DB) {
	GoPostgresConnection[connectionName] = db
}

func GetPostgreConnection(connectionName string) (*gorm.DB, error) {
	if conn, ok := GoPostgresConnection[connectionName]; ok {
		return conn, nil
	} else {
		return nil, errors.New("connection not found")
	}
}
