package bot

import (
	"context"
	"fmt"
	"github.com/NOSTRADA88/telegram-bot-go/internal/bot/fsm"
	"github.com/NOSTRADA88/telegram-bot-go/internal/bot/handlers"
	"github.com/NOSTRADA88/telegram-bot-go/internal/bot/notificator"
	"github.com/NOSTRADA88/telegram-bot-go/internal/config"
	"github.com/NOSTRADA88/telegram-bot-go/internal/logger"
	"github.com/NOSTRADA88/telegram-bot-go/internal/storage/mongodb"
	"github.com/NOSTRADA88/telegram-bot-go/internal/storage/redis"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"time"
)

func Start() error {
	log := logger.New(logger.DebugLevel)

	log.Info("loading config...")

	cfg, err := config.New()

	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
		return err
	}

	log.Info("config loaded successfully")

	bot, err := gotgbot.NewBot(cfg.Token, nil)

	if err != nil {
		log.ErrorF("failed to created bot struct: %v", err)
	}

	set, err := bot.SetMyCommands([]gotgbot.BotCommand{{"start", "Используйте для начала работы с ботом, а также, чтобы вернуться в основное меню"}, {"help", "Информация по использованию бота"}}, nil)

	if err != nil {
		log.ErrorF("failed to set default commands: %v", err)
	}

	if set {
		log.Info("default commands set successfully")
	}

	dispatcher := ext.NewDispatcher(&ext.DispatcherOpts{
		Error: func(b *gotgbot.Bot, ctx *ext.Context, err error) ext.DispatcherAction {
			log.Error("an error occurred while handling update:", err.Error())
			return ext.DispatcherActionNoop
		},
		MaxRoutines: ext.DefaultMaxRoutines,
	})

	log.Info("connecting database...")

	db, err := mongodb.New(cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password)
	if err != nil {
		log.ErrorF("an error occurred on connection to mongo: %v", err)
		panic(fmt.Sprintf("failed to load config: %s", err.Error()))
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errInit := db.Init(ctx)

	if errInit != nil {
		log.ErrorF("failed to init user with unique field tgID: %v", errInit)
	}

	defer func() {
		if errC := db.Close(); errC != nil {
			log.WarnF("failed to close connection to mongo: %v", errC)
		}
	}()

	log.Info("database was connected successfully")

	client := handlers.Client{
		FSM:           fsm.New(redis.New(cfg.Redis.Host, cfg.Redis.Port), ctx),
		Cfg:           cfg,
		Database:      db,
		NotifiedUsers: make(map[string]bool, 100),
	}

	handlers.Set(dispatcher, &client)

	updater := ext.NewUpdater(dispatcher, nil)
	not := notificator.Notificator{NotifiedUsers: make(map[string]bool, 100), Database: db, Cfg: cfg}

	log.Info("start polling")

	err = updater.StartPolling(bot, &ext.PollingOpts{
		DropPendingUpdates: true,
		GetUpdatesOpts: &gotgbot.GetUpdatesOpts{
			Timeout: 9,
			RequestOpts: &gotgbot.RequestOpts{
				Timeout: time.Second * 10,
			},
		},
	})

	not.StartNotificationScheduler(bot)

	updater.Idle()

	return nil
}
