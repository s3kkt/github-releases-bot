package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/s3kkt/github-releases-bot/internal"
	"github.com/s3kkt/github-releases-bot/internal/database"
	"github.com/s3kkt/github-releases-bot/internal/helpers"
	"github.com/s3kkt/github-releases-bot/internal/transport"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var (
	// Menu texts
	firstMenu = "<b>Select GitHub bot action</b>"

	// Button texts
	listButton   = "List repos"
	addButton    = "Add repo"
	deleteButton = "Delete repo"
	latestButton = "Show latest tags"
	helpButton   = "Halp!"
	docsButton   = "Documentation"

	// Service messages
	addMessageText    = "Adding repository. Reply on this message and send GitHub link (format: https://author/repository)"
	deleteMessageText = "Deleting repository. Reply on this message and send GitHub link (format: https://author/repository)"
	helpMessageText   = `
<b>Commands short description and functionality:</b>

/menu   - Show bot menu buttons.
/list   - List added repositories.
/add    - Add repo to list. Bot will send you a message, just answer on it and send repository link.
Example: https://github.com/s3kkt/github-releases-bot
/delete - Delete repo from list. Action is the same as add command.
/latest - Show latest tags and release dates for repos
/help   - Show this message.

<b>For more information press "Docs" button in menu</b>`

	docsText = "https://github.com/s3kkt/github-releases-bot"

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
			tgbotapi.NewInlineKeyboardButtonData(latestButton, latestButton),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(helpButton, helpButton),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL(docsButton, docsText),
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

	closed := make(chan struct{})

	// `updates` is a golang channel which receives telegram updates
	updates := bot.GetUpdatesChan(u)

	// Pass cancellable context to goroutine
	go receiveUpdates(closed, updates)

	// Tell the user the bot is online
	log.Println("Start listening for updates.")

	// wait for os.Signal, then stop handling updates
	interrupt := make(chan os.Signal)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	<-interrupt
	close(closed)
}

func receiveUpdates(closed <-chan struct{}, updates tgbotapi.UpdatesChannel) {
	// `for {` means the loop is infinite until we manually stop it
	for {
		select {
		// wait for os.Signal to stop
		case <-closed:
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
		//fmt.Printf("DEBUG MESSAGE TEXT:             %+v\n", update.Message.Text)
		//fmt.Printf("DEBUG MESSAGE ID:               %+v\n", update.Message.MessageID)
		//fmt.Printf("DEBUG MESSAGE FROM:             %+v\n", update.Message.From)
		//fmt.Printf("DEBUG MESSAGE FRORWARD FROM ID: %+v\n", update.Message.ForwardFromMessageID)
		//fmt.Printf("DEBUG MESSAGE REPLY TO:         %+v\n", update.Message.ReplyToMessage)
		//fmt.Printf("DEBUG MESSAGE USERNAME:         %+v\n", update.Message.Chat.UserName)
		//fmt.Printf("DEBUG MESSAGE LAST NAME:        %+v\n", update.Message.Chat.LastName)
		//fmt.Printf("DEBUG MESSAGE FIRST NAME:       %+v\n", update.Message.Chat.FirstName)
		//fmt.Printf("DEBUG MESSAGE TYPE:             %+v\n", update.Message.Chat.Type)

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
	var err error

	err = updateChat(message)
	if err != nil {
		log.Println(err)
		return
	}

	if user == nil {
		return
	}

	// Print user input to console
	log.Printf("%s(@%s) wrote %s", user.FirstName, user.UserName, text)

	if strings.HasPrefix(text, "/") {
		//err = handleCommand(message.Chat.ID, text)
		err = handleCommand(message, text)
	} else if reply != nil {
		if reply.Text == addMessageText {
			if helpers.ValidateRepoUrl(message.Text) == true {
				database.AddRepo(message.Text, message.Chat.ID)
				msg := tgbotapi.NewMessage(message.Chat.ID, "Adding repo: "+message.Text)
				msg.DisableWebPagePreview = true
				_, err = bot.Send(msg)
			} else {
				msg := tgbotapi.NewMessage(message.Chat.ID, "It is not a GitHub repository URL")
				_, err = bot.Send(msg)
			}
		} else if reply.Text == deleteMessageText {
			if helpers.ValidateRepoUrl(message.Text) == true {
				database.DeleteRepo(message.Text, message.Chat.ID)
				msg := tgbotapi.NewMessage(message.Chat.ID, "Deleting repo: "+message.Text)
				msg.DisableWebPagePreview = true
				_, err = bot.Send(msg)
			} else {
				msg := tgbotapi.NewMessage(message.Chat.ID, "It is not a GitHub repository URL.")
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

func handleCommand(message *tgbotapi.Message, command string) error {
	var err error

	switch command {
	case "/menu":
		err = sendMenu(message.Chat.ID)
		break
	case "/list":
		err = listRepos(message.Chat.ID)
		break
	case "/add":
		err = addRepo(message.Chat.ID)
		break
	case "/delete":
		err = deleteRepo(message.Chat.ID)
		break
	case "/latest":
		err = listLatest(message.Chat.ID)
		break
	case "/help":
		err = sendHelp(message.Chat.ID)
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
	} else if query.Data == latestButton {
		text = latestButton
		err := listLatest(message.Chat.ID)
		if err != nil {
			return
		}
	} else if query.Data == addButton {
		text = addButton
		msg := tgbotapi.NewMessage(message.Chat.ID, addMessageText)
		bot.Send(msg)
	} else if query.Data == deleteButton {
		text = deleteButton
		msg := tgbotapi.NewMessage(message.Chat.ID, deleteMessageText)
		bot.Send(msg)
	} else if query.Data == helpButton {
		text = helpButton
		msg := tgbotapi.NewMessage(message.Chat.ID, helpMessageText)
		msg.ParseMode = tgbotapi.ModeHTML
		msg.DisableWebPagePreview = true
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

func updateChat(message *tgbotapi.Message) error {
	date := time.Now().Unix()
	err := database.UpdateChat(message.Chat.ID, message.Chat.UserName, message.Chat.FirstName, message.Chat.LastName, message.Chat.Type, message.From.IsBot, date)
	return err
}

func sendMenu(chatId int64) error {
	msg := tgbotapi.NewMessage(chatId, firstMenu)
	msg.ParseMode = tgbotapi.ModeHTML
	msg.ReplyMarkup = firstMenuMarkup
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

func listLatest(chatId int64) error {
	latestList, err := database.GetChatLatestList(chatId)
	if err != nil {
		log.Fatal("Failed to get latest list: %w", err)
	} else {
		msg := tgbotapi.NewMessage(chatId, helpers.LatestListOutput(latestList))
		msg.ParseMode = tgbotapi.ModeHTML
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

func sendHelp(chatId int64) error {
	msg := tgbotapi.NewMessage(chatId, helpMessageText)
	msg.ParseMode = tgbotapi.ModeHTML
	msg.DisableWebPagePreview = true
	_, err := bot.Send(msg)
	return err
}

func Notifier(conf internal.Config, chatId int64) {
	duration, _ := time.ParseDuration(conf.UpdateInterval)
	for range time.Tick(duration) {
		log.Print("Bot notifier: check for updates...")
		_, reposList := database.GetChatReposList(chatId)
		if len(reposList) == 0 {
			log.Printf("There is to chats to interact with.")
			return
		}
		for _, repo := range reposList {
			release, err := transport.GetReleases(helpers.GetApiURL(repo), conf.GitHubToken)
			if err != nil {
				log.Println(err)
			} else {
				log.Printf("Checking if %s is new release for %s", release.TagName, repo)
				ifNew, err := database.CheckIfNew(repo, release.TagName)
				if err != nil {
					continue
				} else if ifNew == true {

					log.Printf("%s is new release for %s, inserting data to database.", release.TagName, repo)
					checkTime := time.Now().Format(time.RFC3339)
					database.InsertReleaseData(checkTime, repo, release, true)

					log.Printf("Try to send updates. Chat ID: %d", chatId)
					err := sendReleased(chatId, repo, release.TagName, release.HtmlUrl, release.Body)
					if err != nil {
						log.Printf("Failed to send release update. ChatId: %d. Reason: %s", chatId, err)
						return
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
	msg := tgbotapi.NewMessage(chatId, "<b>"+r+"</b> released <b>"+tag+"</b>\n\n<b>Link: </b>"+url+"\n\n<b>Notes: </b>\n"+n+"\n<a href='"+url+"'>Read more</a>\n")
	msg.ParseMode = tgbotapi.ModeHTML
	_, err := bot.Send(msg)
	return err
}
