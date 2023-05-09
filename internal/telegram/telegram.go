package telegram

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/s3kkt/github-releases-bot/internal"
	"github.com/s3kkt/github-releases-bot/internal/config"
	"github.com/s3kkt/github-releases-bot/internal/database"
	"github.com/s3kkt/github-releases-bot/internal/helpers"
	"github.com/s3kkt/github-releases-bot/internal/transport"
	"log"
	"os"
	"reflect"
	"strings"
	"time"
)

var (
	// Menu texts
	firstMenu = "<b>Select GitHub bot action</b>"

	// Button texts
	listButton   = "List repos"
	addButton    = "Add repo"
	deleteButton = "Delete repo"
	helpButton   = "Halp!"

	addMessageText    = "Adding repository. Reply on this message and send GitHub link (format: https://author/repository)"
	deleteMessageText = "Deleting repository. Reply on this message and send GitHub link (format: https://author/repository)"
	helpText          = "https://github.com/s3kkt/github-releases-bot"

	// Keyboard layout for the first menu. One button, one row
	firstMenuMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(listButton, listButton),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(addButton, addButton),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(deleteButton, deleteButton),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL(helpButton, helpText),
		),
	)

	bot *tgbotapi.BotAPI
)

func Bot(conf internal.Config) {
	var err error
	bot, err = tgbotapi.NewBotAPI(conf.TelegramToken)
	if err != nil {
		// Abort if something is wrong
		log.Panic(err)
	}

	// Set this to true to log all interactions with telegram servers
	bot.Debug = conf.Debug

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	//// Create a new cancellable background context. Calling `cancel()` leads to the cancellation of the context
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	// `updates` is a golang channel which receives telegram updates
	updates := bot.GetUpdatesChan(u)

	//// Pass cancellable context to goroutine
	go receiveUpdates(ctx, updates)

	// Tell the user the bot is online
	log.Println("Start listening for updates. Press enter to stop")

	// Wait for a newline symbol, then cancelу handling updates
	bufio.NewReader(os.Stdin).ReadBytes('\n')
	cancel()

}

func receiveUpdates(ctx context.Context, updates tgbotapi.UpdatesChannel) {
	// `for {` means the loop is infinite until we manually stop it
	for {
		select {
		// stop looping if ctx is cancelled
		case <-ctx.Done():
			return
		// receive update from channel and then handle it
		case update := <-updates:
			handleUpdate(update)
		}
	}
}

func handleUpdate(update tgbotapi.Update) {
	switch {
	// Handle messages
	case update.Message != nil:
		//fmt.Printf("DEBUG MESSAGE: %+v\n", update.Message)
		fmt.Printf("DEBUG MESSAGE TEXT:             %+v\n", update.Message.Text)
		fmt.Printf("DEBUG MESSAGE ID:               %+v\n", update.Message.MessageID)
		fmt.Printf("DEBUG MESSAGE FROM:             %+v\n", update.Message.From)
		fmt.Printf("DEBUG MESSAGE FRORWARD FROM ID: %+v\n", update.Message.ForwardFromMessageID)
		fmt.Printf("DEBUG MESSAGE REPLY TO:         %+v\n", update.Message.ReplyToMessage)

		handleMessage(update.Message, update.Message.ReplyToMessage)
		break
	// Handle button clicks
	case update.CallbackQuery != nil:
		handleButton(update.CallbackQuery)

		break
	}
}

func handleMessage(message *tgbotapi.Message, reply *tgbotapi.Message) {
	user := message.From
	text := message.Text

	if user == nil {
		return
	}

	// Print user input to console
	log.Printf("%s(@%s) wrote %s", user.FirstName, user.UserName, text)

	var err error
	if strings.HasPrefix(text, "/") {
		err = handleCommand(message.Chat.ID, text)
	} else if reply != nil {
		if reply.Text == addMessageText {
			if helpers.ValidateRepoUrl(message.Text) == true {
				database.AddRepo(message.Text, message.Chat.ID)
				msg := tgbotapi.NewMessage(message.Chat.ID, "Adding repo: "+message.Text)
				msg.DisableWebPagePreview = true
				_, err = bot.Send(msg)
			} else {
				msg := tgbotapi.NewMessage(message.Chat.ID, "Sorry :( It is not a GitHub repository URL")
				_, err = bot.Send(msg)
			}
		} else if reply.Text == deleteMessageText {
			if helpers.ValidateRepoUrl(message.Text) == true {
				database.DeleteRepo(message.Text, message.Chat.ID)
				msg := tgbotapi.NewMessage(message.Chat.ID, "Deleting repo: "+message.Text)
				msg.DisableWebPagePreview = true
				_, err = bot.Send(msg)
			} else {
				msg := tgbotapi.NewMessage(message.Chat.ID, "Sorry :( It is not a GitHub repository URL.")
				_, err = bot.Send(msg)
			}
		}
	} else {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Unknown command! Send /menu")
		_, err = bot.Send(msg)
	}

	if err != nil {
		log.Printf("An error occured: %s", err.Error())
	}
}

func handleCommand(chatId int64, command string) error {
	var err error

	switch command {
	case "/menu":
		err = sendMenu(chatId)
		break
	case "/list":
		err = listRepos(chatId)
		break
	case "/add":
		err = addRepo(chatId)
		break
	case "/delete":
		err = deleteRepo(chatId)
		break
	case "/help":
		err = sendHelp(chatId)
		break
	}

	return err
}

func handleButton(query *tgbotapi.CallbackQuery) {
	var text string

	markup := tgbotapi.NewInlineKeyboardMarkup()
	message := query.Message

	if query.Data == listButton {
		text = firstMenu
		markup = firstMenuMarkup
		err, reposList := database.GetChatReposList(message.Chat.ID)
		if err != nil {
			log.Fatal("Failed to get repos list: %w", err)
		} else {
			msg := tgbotapi.NewMessage(message.Chat.ID, helpers.ReposListOutput(reposList))
			msg.DisableWebPagePreview = true
			bot.Send(msg)
		}
	} else if query.Data == addButton {
		text = addButton
		msg := tgbotapi.NewMessage(message.Chat.ID, addMessageText)
		bot.Send(msg)
	} else if query.Data == deleteButton {
		text = deleteButton
		msg := tgbotapi.NewMessage(message.Chat.ID, deleteMessageText)
		bot.Send(msg)
	}

	callbackCfg := tgbotapi.NewCallback(query.ID, "")
	bot.Send(callbackCfg)

	// Replace menu text and keyboard
	msg := tgbotapi.NewEditMessageTextAndMarkup(message.Chat.ID, message.MessageID, text, markup)
	msg.ParseMode = tgbotapi.ModeHTML
	//msg.ParseMode = tgbotapi.ModeMarkdownV2
	bot.Send(msg)
}

func sendMenu(chatId int64) error {
	msg := tgbotapi.NewMessage(chatId, firstMenu)
	msg.ParseMode = tgbotapi.ModeHTML
	msg.ReplyMarkup = firstMenuMarkup
	_, err := bot.Send(msg)
	return err
}

func sendHelp(chatId int64) error {
	msg := tgbotapi.NewMessage(chatId, helpText)
	msg.ParseMode = tgbotapi.ModeHTML
	_, err := bot.Send(msg)
	return err
}

func listRepos(chatId int64) error {
	err, reposList := database.GetChatReposList(chatId)
	if err != nil {
		log.Fatal("Failed to get repos list: %w", err)
	} else {
		msg := tgbotapi.NewMessage(chatId, helpers.ReposListOutput(reposList))
		msg.DisableWebPagePreview = true
		bot.Send(msg)
	}
	return err
}

func addRepo(chatId int64) error {
	msg := tgbotapi.NewMessage(chatId, addMessageText)
	_, err := bot.Send(msg)
	return err
}

func deleteRepo(chatId int64) error {
	msg := tgbotapi.NewMessage(chatId, deleteMessageText)
	_, err := bot.Send(msg)
	return err
}

func Notifier(conf internal.Config) {
	connectionString := os.Getenv("DB_CONNECTION_STRING")
	duration, _ := time.ParseDuration(conf.UpdateInterval)
	for range time.Tick(duration) {
		log.Print("Bot notifier: check for updates...")
		_, reposList := database.GetReposList()
		for _, repo := range reposList {
			release, err := transport.GetReleases(config.GetApiURL(repo), conf.GitHubToken)
			if err != nil {
				log.Println(err)
			} else {
				log.Printf("DEBUG: checking if %s is new release for %s", release.TagName, repo)
				ifNew, err := database.CheckIfNew(repo, release.TagName)
				if err != nil {
					log.Println("DEBUG: response from CheckIfNew:", err)
					continue
				} else if ifNew == true {
					log.Println("DEBUG: requesting for repos IDs")
					var chatIds []int64
					sqlStatement := `SELECT chat_id FROM repos WHERE name = $1 and deleted = false;`

					db, err := sql.Open("postgres", connectionString)
					if err != nil {
						log.Fatal(err)
					}

					rows, err := db.Query(sqlStatement, repo)
					if err != nil {
						log.Fatal(err)
					}
					defer rows.Close()

					for rows.Next() {
						var i int64
						if err := rows.Scan(&i); err != nil {
							fmt.Println(err)
						}
						chatIds = append(chatIds, i)
					}

					if err = rows.Err(); err != nil {
						fmt.Println(err)
					}

					defer db.Close()

					checkTime := time.Now().Format(time.RFC3339)
					log.Println("DEBUG: inserting new release data in database")
					database.InsertReleaseData(checkTime, repo, release)

					for _, id := range chatIds {
						if reflect.TypeOf(id).Kind() == reflect.Int64 {
							log.Printf("Try to send updates. Chat ID: %d", id)
							err := sendReleased(id, repo, release.TagName, release.HtmlUrl, release.Body)
							if err != nil {
								log.Printf("Failed to send release update to chat with id %d. Reason: %s", id, err)
								return
							}
						} else {
							log.Printf("Cannot send updates. Chat ID: %d", id)
						}
					}
				}
			}
		}
	}
	return
}

func sendReleased(chatId int64, repoName, tag, url, releaseNotes string) error {
	r := helpers.SanitizeRepoName(repoName)
	n := helpers.SanitizeReleaseNotes(releaseNotes)
	msg := tgbotapi.NewMessage(chatId, "<b>"+r+"</b> released <b>"+tag+"</b>\n\n<b>Link: </b>"+url+"\n<b>Notes: </b>\n"+n)
	msg.ParseMode = tgbotapi.ModeHTML
	_, err := bot.Send(msg)
	return err
}
