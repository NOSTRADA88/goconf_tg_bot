package config

import (
	"fmt"
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"time"
)

type Config struct {
	Conference
	Database
	Telegram
	Redis
	DebugLevel int `env:"DEBUG_LEVEL" envDefault:"0"`
}

type Database struct {
	Host     string `env:"DB_HOST" envDefault:"localhost"`
	Port     int    `env:"DB_PORT" envDefault:"5432"`
	Password string `env:"DB_PASSWORD" envDefault:"postgres"`
	User     string `env:"DB_USER" envDefault:"postgres"`
}

type Telegram struct {
	Token string `env:"TELEGRAM_TOKEN" env-required:"true"`
	Administrators
}

type Administrators struct {
	IDs      []int `env:"ADMIN_IDS_LIST" envSeparator:","`
	IDsInMap map[int]bool
}

type Redis struct {
	Host string `env:"REDIS_HOST" envDefault:"localhost"`
	Port int    `env:"REDIS_PORT" envDefault:"6379"`
}

type confTime time.Time

type Conference struct {
	Name                 string   `env:"CONFERENCE_NAME" env-required:"true"`
	URL                  string   `env:"CONFERENCE_URL" env-required:"true"`
	TimeFrom             confTime `env:"CONFERENCE_FROM_TIME" env-required:"true"`
	TimeUntil            confTime `env:"CONFERENCE_UNTIL_TIME" env-required:"true"`
	TimeReviewsAvailable confTime `env:"CONFERENCE_REVIEWS_AVAILABLE_TIME" env-required:"true"`
}

func (c *confTime) UnmarshalText(text []byte) error {
	t, err := time.Parse("02/01/2006 15:04:05", string(text))
	*c = confTime(t)
	if err != nil {
		panic(fmt.Sprintf("con't parse time: %v", err))
	}
	return nil
}

func New() (*Config, error) {

	var err error

	err = godotenv.Load(".env")

	if err != nil {
		return nil, err
	}

	conf := new(Conference)

	err = env.Parse(conf)

	if err != nil {
		return nil, err
	}

	cfg := new(Config)

	cfg.Conference = *conf
	err = env.Parse(cfg)

	if err != nil {
		return nil, err
	}

	if time.Time(cfg.Conference.TimeFrom).After(time.Time(cfg.Conference.TimeUntil)) {
		panic("time CONFERENCE_FROM_TIME bigger than CONFERENCE_UNTIL_TIME, it should be the other way around")
	}

	if time.Time(cfg.Conference.TimeUntil).After(time.Time(cfg.Conference.TimeReviewsAvailable)) {
		panic("time CONFERENCE_REVIEWS_AVAILABLE_TIME less than CONFERENCE_UNTIL_TIME, it should be the other way around")
	}

	cfg.Administrators.IDsInMap = make(map[int]bool, len(cfg.IDs))

	for _, v := range cfg.Administrators.IDs {
		if _, exists := cfg.Administrators.IDsInMap[v]; !exists {
			cfg.Administrators.IDsInMap[v] = true
		}
	}

	return cfg, nil
}
