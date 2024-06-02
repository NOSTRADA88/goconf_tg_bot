- [Used libraries](#used-libraries)
- [Project Structure](#project-structure)
- [Getting Started](#getting-started)


## Used libraries
- [gotgbot](https://github.com/PaulSonOfLars/gotgbot)
- [godotenv](https://github.com/joho/godotenv)
- [env](https://github.com/caarlos0/env)
- [mongo-go-driver](https://github.com/mongodb/mongo-go-driver)
- [go-redis](https://github.com/redis/go-redis)

## Project Structure
- **cmd/**: Contains the entry point of the application.
    - **telegram-bot-go/**: Houses the main application executable.
        - **main.go**: Main application file.

- **internal/**: Core application code that is not intended to be exported.
    - **bot/**: Telegram bot handlers, routers, and keyboards.
        - **fsm/**: Simple finite state machine.
        - **handlers/**: Handlers for the bot.
    - **config/**: Configuration files.
        - **config.go**: Main configuration file.
        - **config_test.go**: Tests for the configuration.
    - **logger/**: Logging.
        - **logger.go**: Main logging file.
        - **logger_test.go**: Tests for logging.
    - **models/**: Data models.
        - **models.go**: Main data models.
    - **repository/**: Data access implementations.
        - **mongodb/**: MongoDB repository implementations.
        - **redis/**: Redis cache implementations.

- **go.mod** and **go.sum**: Go module files for managing dependencies.

## Getting Started
This is a telegram bot for _GolangConf 2024_.

#### Few easy steps to start a project :
1. `git pull https://github.com/NOSTRADA88/telegram-bot-go`
2. `Create and set up your own ".env" file (in the pulled directory). In project, you can find an ".env.example", make sure that you are using THE SAME VARIABLES= as they named in the example file.`
3. `If ".env" is ready, then type in cmd line` `docker-compose build`
4. `docker-compose up -d`
