// Package config provides structures and functions for managing application configuration.
package config

import (
	"fmt"
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"time"
)

// Config is the main configuration structure, containing all the sub-configurations.
type Config struct {
	Conference
	Database
	Telegram
	Redis
	DebugLevel int `env:"DEBUG_LEVEL" envDefault:"0"` // DebugLevel is the level of debugging. 0 is default.
}

// Database is the configuration structure for the database.
type Database struct {
	Host     string `env:"DB_HOST" envDefault:"localhost"` // Host is the database host. Default is localhost.
	Port     int    `env:"DB_PORT" envDefault:"27017"`     // Port is the database port. Default is 27017.
	Password string `env:"DB_PASSWORD" envDefault:""`      // Password is the database password. Default is "".
	User     string `env:"DB_USER" envDefault:""`          // User is the database user. Default is "".
}

// Telegram is the configuration structure for Telegram.
type Telegram struct {
	Token string `env:"TELEGRAM_TOKEN" env-required:"true"` // Token is the Telegram bot token. It is required.
	Administrators
}

// Administrators is the configuration structure for Telegram administrators.
type Administrators struct {
	IDs      []int        `env:"ADMIN_IDS_LIST" envSeparator:","` // IDs is the list of administrator IDs.
	IDsInMap map[int]bool // IDsInMap is a map of administrator IDs for quick lookup.
}

// Redis is the configuration structure for Redis.
type Redis struct {
	Host string `env:"REDIS_HOST" envDefault:"localhost"` // Host is the Redis host. Default is localhost.
	Port int    `env:"REDIS_PORT" envDefault:"6379"`      // Port is the Redis port. Default is 6379.
}

// confTime is a custom time type for unmarshalling time from environment variables.
type confTime time.Time

// Conference is the configuration structure for the conference.
type Conference struct {
	Name                 string   `env:"CONFERENCE_NAME" env-required:"true"`                   // Name is the conference name. It is required.
	URL                  string   `env:"CONFERENCE_URL" env-required:"true"`                    // URL is the conference URL. It is required.
	TimeFrom             confTime `env:"CONFERENCE_FROM_TIME" env-required:"true"`              // TimeFrom is the start time of the conference. It is required.
	TimeUntil            confTime `env:"CONFERENCE_UNTIL_TIME" env-required:"true"`             // TimeUntil is the end time of the conference. It is required.
	TimeReviewsAvailable confTime `env:"CONFERENCE_REVIEWS_AVAILABLE_TIME" env-required:"true"` // TimeReviewsAvailable is the time when reviews become available. It is required.
}

// UnmarshalText unmarshals a byte slice into a confTime.
func (c *confTime) UnmarshalText(text []byte) error {
	t, err := time.Parse("02/01/2006 15:04:05", string(text))
	*c = confTime(t)
	if err != nil {
		panic(fmt.Sprintf("con't parse time: %v", err))
	}
	return nil
}

// New creates a new Config, loading values from environment variables.
func New() (*Config, error) {
	var err error

	// Load environment variables from .env file.
	err = godotenv.Load(".env")

	if err != nil {
		return nil, err
	}

	conf := new(Conference)

	// Parse environment variables into Conference struct.
	err = env.Parse(conf)

	if err != nil {
		return nil, err
	}

	cfg := new(Config)

	// Parse environment variables into Config struct.
	cfg.Conference = *conf
	err = env.Parse(cfg)

	if err != nil {
		return nil, err
	}
	fmt.Println(cfg.Administrators.IDs)
	// Conference name should be given
	if cfg.Conference.Name == "" {
		return nil, fmt.Errorf("CONFERENCE_NAME is required")
	}

	// Conference URL should be given
	if cfg.Conference.URL == "" {
		return nil, fmt.Errorf("CONFERENCE_URL is required")
	}

	// Check that conference times are logical.
	if time.Time(cfg.Conference.TimeFrom).After(time.Time(cfg.Conference.TimeUntil)) {
		panic("time CONFERENCE_FROM_TIME bigger than CONFERENCE_UNTIL_TIME, it should be the other way around")
	}

	if time.Time(cfg.Conference.TimeUntil).After(time.Time(cfg.Conference.TimeReviewsAvailable)) {
		panic("time CONFERENCE_REVIEWS_AVAILABLE_TIME less than CONFERENCE_UNTIL_TIME, it should be the other way around")
	}

	// Create a map of administrator IDs for quick lookup.
	cfg.Telegram.Administrators.IDsInMap = make(map[int]bool, len(cfg.Telegram.Administrators.IDs))
	for _, v := range cfg.Telegram.Administrators.IDs {
		if _, exists := cfg.Telegram.Administrators.IDsInMap[v]; !exists {
			cfg.Telegram.Administrators.IDsInMap[v] = true
		}
	}

	return cfg, nil
}
