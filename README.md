# GitHub updates watcher for Telegram

## Motivation
We often use fixed versions of opensource code when describing our infrastructure as a code to avoid accidental updates.  
When there are a lot of such repositories used, it becomes almost impossible to keep track of updates, so this tool appeared.  
Of course, you can subscribe to RSS feeds or e-mail notifications, but this is not applicable for team work.

Project based on [go-telegram-bot-api](https://github.com/go-telegram-bot-api/telegram-bot-api).

You can add [my telegram bot](https://t.me/awesome_gh_bot) or run your own.

## Requirements:
- GitHub account to access API
- Telegram account to interact with Telegram API
- Postgresql instance to store Bot data
- Docker
- Golang 1.20

## Configuration
You can use both `config.yml` and environment variables to set up github-releases-bot. If environment variables set, they are prior.

### Get GitHub token
Use the official [GitHub documentation](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token#creating-a-fine-grained-personal-access-token) to create your personal token.

### Get Telegram token
According to the official [Telegram documentation](https://core.telegram.org/bots/features#botfather):
> Creating a new bot
> Use the /newbot command to create a new bot. [@BotFather](https://t.me/botfather) will ask you for a name and username, then generate an authentication token for your new bot.
> - The name of your bot is displayed in contact details and elsewhere.
> - The username is a short name, used in search, mentions and t.me links. Usernames are 5-32 characters long and not case sensitive – but may only include Latin characters, numbers, and underscores. Your bot's username must end in 'bot’, like 'tetris_bot' or 'TetrisBot'.
> - The token is a string, like `110201543:AAHdqTcvCH1vGWJxfSeofSAs0K5PALDsaw`, which is required to authorize the bot and send requests to the Bot API. Keep your token secure and store it safely, it can be used by anyone to control your bot.


### config.yml example
```yaml
# your GitHub token with read repos rights
github_token: "*****"

# telegram token given by BotFather
telegram_token: "*****"

# set `debug` to `true` to view debug messages from go-telegram-bot-api
debug: false

# Interval to chech repos updates. Be careful! There are GitHub API limitations in 5000 requests per hour.
update_interval: 10m

# Database settings. For now only Postgres is supported.
database:
  user: "github-bot"
  pass: "VeryStrongPassword"
  host: "127.0.0.1"
  port: 5432
  dbname: "github-bot-db"
```

### Environment variables
All config settings can be overwritten by environment variables

| Config key       | Environment variable | type                                  |
|------------------|----------------------|---------------------------------------|
| github_token     | BOT_GITHUB_TOKEN     | string                                |
| telegram_token   | BOT_TELEGRAM_TOKEN   | string                                |
| debug            | BOT_DEBUG            | bool                                  |
| update_interval  | BOT_UPDATE_INTERVAL  | time in s(seconds), h(hours), d(days) |
| database.user    | BOT_DB_USER          | string                                |
| database.pass    | BOT_DB_PASS          | string                                |
| database.host    | BOT_DB_HOST          | string                                |
| database.port    | BOT_DB_PORT          | int                                   |
| database.dbname  | BOT_DB_NAME          | string                                |


### Flags

| Flag        | Description                                                                                                           | Default | 
|-------------|-----------------------------------------------------------------------------------------------------------------------|---------|
| -config     | path to .yml config file                                                                                              | -       |
| -migrations | if true, only migrations will be applied, bot will not start                                                          | false   |
| -cloud      | if true, bot will use database connection string formatted to use Google cloud SQL instance, described in `main.tf`   | false   |


## Database
For now only Postgres is supported.
`github-releases-bot` needs minimal instance of Postgres (2 CPU cores, 4 GB RAM, 5 GB disk).

### Migrations
Before we start we need to prepare database. To run migrations you can:
1. Install go migrations package ang run migrations manually:
```bash
git clone https://github.com/s3kkt/github-releases-bot.git && cd github-releases-bot
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
export DB_URL="postgres://<db-user>:<password>@<db-host>:<db-port>/<db-name>?sslmode=disable"
migrate -database ${DB_URL} -path ./internal/database/migrations up
```

## Run bot

### Docker-compose
You can run test installation of `github-releases-bot` via docker-compose.
1. Prepare bot environment variables as described above and put it in `configs/.env` subfolder of this project
2. Check your database port
3. Run `docker-compose up` command

```bash
docker-compose up -d --build
```

### Docker
To start production variant of `github-releases-bot` in docker with dedicated Postgres instance, follow this simple steps:
1. Prepare database as described above
2. 

```bash
docker run -v <path-to config.yml>:/etc/gh-bot/config.yml s3kkt/github-releases-bot:latest -config="/etc/gh-bot/config.yml"
```

or use environment variables:
```bash
docker run --env-file <path to .env file> s3kkt/github-releases-bot:latest
```

### Google Cloud Platform
ToDo

## Known issues
1. Latest release database table is not updating when rolling back GitHub release
2. Bot does not work in group chats for now :(