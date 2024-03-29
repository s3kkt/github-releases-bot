package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	zlog "github.com/rs/zerolog/log"
	"github.com/s3kkt/github-releases-bot/internal"
	"github.com/s3kkt/github-releases-bot/internal/database"
	"github.com/s3kkt/github-releases-bot/internal/helpers"
	"github.com/s3kkt/github-releases-bot/internal/transport"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type TGStruct struct {
	githubRepo *database.GithubRepos
}

func NewTG(githubRepo *database.GithubRepos) *TGStruct {
	return &TGStruct{
		githubRepo: githubRepo,
	}
}

var (
	// Menu texts
	firstMenu = "<b>Select GitHub bot action</b>"

	// Button texts
	listButton   = "List repos"
	addButton    = "Add repo"
	deleteButton = "Delete repo"
	latestButton = "Show latest tags"
	helpButton   = "Halp!"
	docsButton   = "GitHub"

	// Service messages
	addMessageText    = "Adding repository. Reply on this message and send GitHub link (format: https://author/repository)"
	deleteMessageText = "Deleting repository. Reply on this message and send GitHub link (format: https://author/repository)"
	helpMessageText   = `
<b>Bot commands short description:</b>

/menu   - Show bot menu buttons
/list   - List added repositories
/add    - Add repo to list. Bot will send you a message, just answer on it and send repository link.
Example: https://github.com/s3kkt/github-releases-bot
/delete - Delete repo from list. Action is the same as 'add' command
/latest - Show latest tags and release dates for repos
/help   - Show this message

<b>For more information press "GitHub" button in menu</b>`

	docsText = "https://github.com/s3kkt/github-releases-bot/blob/master/README.md"

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

func (r *TGStruct) Bot(conf internal.Config) {
	var err error
	bot, err = tgbotapi.NewBotAPI(conf.TelegramToken)
	if err != nil {
		// Abort if something is wrong
		zlog.Panic().Msgf("Cannot initialize new bot: %s", err)
	}

	// Set this to true to log all interactions with telegram servers
	bot.Debug = conf.Debug

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	closed := make(chan struct{})

	// `updates` is a golang channel which receives telegram updates
	updates := bot.GetUpdatesChan(u)

	// Pass cancellable context to goroutine
	go r.receiveUpdates(closed, updates)

	// Tell the user the bot is online
	zlog.Info().Msg("Start listening for updates.")

	// wait for os.Signal, then stop handling updates
	interrupt := make(chan os.Signal)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	<-interrupt
	close(closed)
}

func (r *TGStruct) receiveUpdates(closed <-chan struct{}, updates tgbotapi.UpdatesChannel) {
	// `for {` means the loop is infinite until we manually stop it
	for {
		select {
		// wait for os.Signal to stop
		case <-closed:
			return
		// receive update from channel and then handle it
		case update := <-updates:
			r.handleUpdate(update)
		}
	}
}

func (r *TGStruct) handleUpdate(update tgbotapi.Update) {
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

		r.handleMessage(update.Message, update.Message.ReplyToMessage)
		break
	// Handle button clicks
	case update.CallbackQuery != nil:
		r.handleButton(update.CallbackQuery)

		break
	}
}

func (r *TGStruct) handleMessage(message *tgbotapi.Message, reply *tgbotapi.Message) {
	user := message.From
	text := message.Text
	var err error

	err = r.updateChat(message)
	if err != nil {
		zlog.Error().Msgf("%s", err)
		return
	}

	if user == nil {
		return
	}

	// Print user input to console
	zlog.Info().Msgf("%s(@%s) wrote %s", user.FirstName, user.UserName, text)

	if strings.HasPrefix(text, "/") {
		err = r.handleCommand(message, text)
	} else if reply != nil {
		if reply.Text == addMessageText {
			if helpers.ValidateRepoUrl(message.Text) == true {
				r.githubRepo.AddRepo(message.Text, message.Chat.ID)
				msg := tgbotapi.NewMessage(message.Chat.ID, "Adding repo: "+message.Text)
				msg.DisableWebPagePreview = true
				_, err = bot.Send(msg)
			} else {
				msg := tgbotapi.NewMessage(message.Chat.ID, "It is not a GitHub repository URL")
				_, err = bot.Send(msg)
			}
		} else if reply.Text == deleteMessageText {
			if helpers.ValidateRepoUrl(message.Text) == true {
				r.githubRepo.DeleteRepo(message.Text, message.Chat.ID)
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
		zlog.Error().Msgf("An error occured: %s", err.Error())
	}
}

func (r *TGStruct) handleCommand(message *tgbotapi.Message, command string) error {
	var err error

	switch command {
	case "/menu":
		err = r.sendMenu(message.Chat.ID)
		break
	case "/list":
		err = r.listRepos(message.Chat.ID)
		break
	case "/add":
		err = r.addRepo(message.Chat.ID)
		break
	case "/delete":
		err = r.deleteRepo(message.Chat.ID)
		break
	case "/latest":
		err = r.listLatest(message.Chat.ID)
		break
	case "/help":
		err = r.sendHelp(message.Chat.ID)
		break
	}

	return err
}

func (r *TGStruct) handleButton(query *tgbotapi.CallbackQuery) {
	var text string

	markup := tgbotapi.NewInlineKeyboardMarkup()
	message := query.Message

	if query.Data == listButton {
		text = firstMenu
		markup = firstMenuMarkup
		err, reposList := r.githubRepo.GetChatReposList(message.Chat.ID)
		if err != nil {
			zlog.Error().Msgf("Failed to get repos list: %w", err)
		} else {
			msg := tgbotapi.NewMessage(message.Chat.ID, helpers.ReposListOutput(reposList))
			msg.DisableWebPagePreview = true
			bot.Send(msg)
		}
	} else if query.Data == latestButton {
		text = latestButton
		err := r.listLatest(message.Chat.ID)
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
	bot.Send(msg)
}

func (r *TGStruct) updateChat(message *tgbotapi.Message) error {
	date := time.Now().Unix()
	err := r.githubRepo.UpdateChat(message.Chat.ID, message.Chat.UserName, message.Chat.FirstName, message.Chat.LastName, message.Chat.Type, message.From.IsBot, date)
	return err
}

func (r *TGStruct) sendMenu(chatId int64) error {
	msg := tgbotapi.NewMessage(chatId, firstMenu)
	msg.ParseMode = tgbotapi.ModeHTML
	msg.ReplyMarkup = firstMenuMarkup
	_, err := bot.Send(msg)
	return err
}

func (r *TGStruct) listRepos(chatId int64) error {
	err, reposList := r.githubRepo.GetChatReposList(chatId)
	if err != nil {
		zlog.Error().Msgf("Failed to get repos list: %w", err)
	} else {
		msg := tgbotapi.NewMessage(chatId, helpers.ReposListOutput(reposList))
		msg.DisableWebPagePreview = true
		bot.Send(msg)
	}
	return err
}

func (r *TGStruct) listLatest(chatId int64) error {
	latestList, err := r.githubRepo.GetChatLatestList(chatId)
	if err != nil {
		zlog.Error().Msgf("Failed to get latest list: %w", err)
	} else {
		msg := tgbotapi.NewMessage(chatId, helpers.LatestListOutput(latestList))
		msg.ParseMode = tgbotapi.ModeHTML
		msg.DisableWebPagePreview = true
		bot.Send(msg)
	}
	return err
}

func (r *TGStruct) addRepo(chatId int64) error {
	msg := tgbotapi.NewMessage(chatId, addMessageText)
	_, err := bot.Send(msg)
	return err
}

func (r *TGStruct) deleteRepo(chatId int64) error {
	msg := tgbotapi.NewMessage(chatId, deleteMessageText)
	_, err := bot.Send(msg)
	return err
}

func (r *TGStruct) sendHelp(chatId int64) error {
	msg := tgbotapi.NewMessage(chatId, helpMessageText)
	msg.ParseMode = tgbotapi.ModeHTML
	msg.DisableWebPagePreview = true
	_, err := bot.Send(msg)
	return err
}

func (r *TGStruct) Notifier(conf internal.Config, chatId int64) {
	duration, _ := time.ParseDuration(conf.UpdateInterval)
	for range time.Tick(duration) {
		zlog.Info().Msg("Bot notifier: check for updates...")
		_, reposList := r.githubRepo.GetChatReposList(chatId)
		if len(reposList) == 0 {
			zlog.Info().Msg("There is to chats to interact with.")
			return
		}
		for _, repo := range reposList {
			release, err := transport.GetReleases(helpers.GetApiURL(repo), conf.GitHubToken)
			if err != nil {
				zlog.Error().Msgf("Cannot get releases", err)
			} else {
				zlog.Info().Msgf("Checking if %s is new release for %s", release.TagName, repo)
				ifNew, err := r.githubRepo.CheckIfNew(repo, release.TagName)
				if err != nil {
					continue
				} else if ifNew == true {

					zlog.Info().Msgf("%s is new release for %s, inserting data to database.", release.TagName, repo)
					checkTime := time.Now().Format(time.RFC3339)
					r.githubRepo.InsertReleaseData(checkTime, repo, release, true)

					zlog.Info().Msgf("Try to send updates. Chat ID: %d", chatId)
					err := r.sendReleased(chatId, repo, release.TagName, release.HtmlUrl, release.Body)
					if err != nil {
						zlog.Info().Msgf("Failed to send release update. ChatId: %d. Reason: %s", chatId, err)
						return
					}
				}
			}
		}
	}
	return
}

func (r *TGStruct) sendReleased(chatId int64, repoName, tag, url, releaseNotes string) error {
	repo := helpers.SanitizeRepoName(repoName)
	note := helpers.SanitizeReleaseNotes(releaseNotes)
	msg := tgbotapi.NewMessage(chatId, "<b>"+repo+"</b> released <b>"+tag+"</b>\n\n<b>Link: </b>"+url+"\n\n<b>Notes: </b>\n"+note+"\n<a href='"+url+"'>Read more</a>\n")
	msg.ParseMode = tgbotapi.ModeHTML
	_, err := bot.Send(msg)
	return err
}
