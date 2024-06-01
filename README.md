-  [Getting Started](#getting-started)
-  [Project Structure](#project-structure)
-  [Used libraries](#used-libraries)


## _Getting Started with_
This is a telegram bot for _GolangConf 2024_.

Easy steps to start a project 
1. git pull ...
2. set up an .env file. In project, you can find an .env.example file, make sure that you are using the same VARIABLES= as they named in the example file.
3. docker-compose ...


## Used libraries
- [gotgbot](https://github.com/PaulSonOfLars/gotgbot)
- [godotenv](https://github.com/joho/godotenv)
- [env](https://github.com/caarlos0/env)
- [mongo-go-driver](https://github.com/mongodb/mongo-go-driver)
- [go-redis](https://github.com/redis/go-redis)

## Project Structure
- `cmd/`: Contains the entry point of the application.
    - `telegram-bot-go/`: Houses the main application executable.
- `internal/`: Core application code that is not intended to be exported.
    - `bot/`: Telegram bot handlers and routers.
      - `fsm/`:
      - `handlers/`
    - `config/`:
    - `models/`:
    - `usecase/`: Application-specific business rules.
    - `repository/`: Data access implementations.
        - `mongodb/`: PostgreSQL repository implementations.
        - `redis/`: Redis cache implementations.
- `go.mod` and `go.sum`: Go module files for managing dependencies.


