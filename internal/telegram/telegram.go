package telegram

import (
	"bufio"
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/s3kkt/github-releases-bot/internal"
	"github.com/s3kkt/github-releases-bot/internal/database"
	"github.com/s3kkt/github-releases-bot/internal/helpers"
	"log"
	"os"
	"strings"
)

//func Bot(connectionString string, conf internal.Config) {
//	//Создаем бота
//	bot, err := tgbotapi.NewBotAPI(conf.TelegramToken)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	bot.Debug = conf.Debug
//
//	log.Printf("Stadting bot: %s. Debug: %t", bot.Self.UserName, conf.Debug)
//	//Устанавливаем время обновления
//	u := tgbotapi.NewUpdate(0)
//	u.Timeout = 60
//	//Получаем обновления от бота
//	updates, err := bot.GetUpdatesChan(u)
//
//	for update := range updates {
//		if update.Message == nil {
//			continue
//		}
//
//		//Проверяем что от пользователья пришло именно текстовое сообщение
//		//fmt.Println(update.Message.Text)
//		if reflect.TypeOf(update.Message.Text).Kind() == reflect.String && update.Message.Text != "" {
//
//			switch update.Message.Text {
//			case "/start":
//				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Hi, i'm a GitHub releases bot! I can check Github repositories for new releases.")
//				bot.Send(msg)
//
//			case "/list":
//				reposList, err := database.GetReposList(connectionString)
//				if err != nil {
//					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Something went wrong. Cannot get repository list. Error: "+err.Error())
//					bot.Send(msg)
//				}
//				log.Printf("%v", reposList)
//
//				answer := fmt.Sprintf("%v", reposList)
//
//				msg := tgbotapi.NewMessage(update.Message.Chat.ID, answer)
//				bot.Send(msg)
//
//			case "/add":
//				//reply := tgbotapi.ForceReply{
//				//	ForceReply: true,
//				//	Selective:  false,
//				//}
//
//				if update.InlineQuery
//
//				//repoUrl, err := helpers.GetArgFromCommand(update.Message.Text)
//				//if err != nil {
//				//	log.Printf("Error: %s", err)
//				//	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Error: "+err.Error())
//				//	bot.Send(msg)
//				//} else {
//				//	log.Printf("Validating repo: %s", repoUrl)
//				//	if helpers.ValidateRepoUrl(repoUrl) == true {
//				//		log.Printf("Ubdating repository list. Add repo: %s", repoUrl)
//				//		database.AddRepo(connectionString, repoUrl)
//				//		if err != nil {
//				//			log.Printf("Error: %s", err)
//				//			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Error: "+err.Error())
//				//			bot.Send(msg)
//				//		}
//				//	} else {
//				//		log.Printf("Error: %s", err)
//				//		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Error: "+err.Error())
//				//		bot.Send(msg)
//				//	}
//				//}
//			default:
//				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Unknown command :(")
//				bot.Send(msg)
//			}
//
//		} else {
//			log.Println("Error: got not string command.")
//			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Use the words for search.")
//			bot.Send(msg)
//		}
//	}
//}

//func Bot(conf internal.Config) {
//	// Get token from the environment variable
//	token := conf.TelegramToken
//	if token == "" {
//		panic("TOKEN environment variable is empty")
//	}
//
//	// Create bot.
//	b, err := gotgbot.NewBot(token, &gotgbot.BotOpts{
//		Client: http.Client{},
//		DefaultRequestOpts: &gotgbot.RequestOpts{
//			Timeout: gotgbot.DefaultTimeout,
//			APIURL:  gotgbot.DefaultAPIURL,
//		},
//	})
//	if err != nil {
//		panic("failed to create new bot: " + err.Error())
//	}
//
//	// Create updater and dispatcher.
//	updater := ext.NewUpdater(&ext.UpdaterOpts{
//		Dispatcher: ext.NewDispatcher(&ext.DispatcherOpts{
//			// If an error is returned by a handler, log it and continue going.
//			Error: func(b *gotgbot.Bot, ctx *ext.Context, err error) ext.DispatcherAction {
//				log.Println("an error occurred while handling update:", err.Error())
//				return ext.DispatcherActionNoop
//			},
//			MaxRoutines: ext.DefaultMaxRoutines,
//		}),
//	})
//	dispatcher := updater.Dispatcher
//
//	// /start command to introduce the bot
//	dispatcher.AddHandler(handlers.NewCommand("start", start))
//	//// /list command to list repositories
//	//dispatcher.AddHandler(handlers.NewCommand("list", list))
//	// Callback button to list repositories
//	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Equal("list_repo"), list))
//	// Callback button to add repositories
//	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Equal("add_repo"), add))
//
//	// Start receiving updates.
//	err = updater.StartPolling(b, &ext.PollingOpts{
//		DropPendingUpdates: true,
//		GetUpdatesOpts: gotgbot.GetUpdatesOpts{
//			Timeout: 9,
//			RequestOpts: &gotgbot.RequestOpts{
//				Timeout: time.Second * 10,
//			},
//		},
//	})
//	if err != nil {
//		panic("failed to start polling: " + err.Error())
//	}
//	log.Printf("%s has been started...\n", b.User.Username)
//
//	// Idle, to keep updates coming in, and avoid bot stopping.
//	updater.Idle()
//}
//
//// start introduces the bot.
//func start(b *gotgbot.Bot, ctx *ext.Context) error {
//	_, err := ctx.EffectiveMessage.Reply(b, fmt.Sprintf("Hello, I'm @%s. I <b>repeat</b> all your messages.", b.User.Username), &gotgbot.SendMessageOpts{
//		ParseMode: "html",
//		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
//			InlineKeyboard: [][]gotgbot.InlineKeyboardButton{{
//				{Text: "List repositories", CallbackData: "list_repo"},
//			}},
//		},
//	})
//	if err != nil {
//		return fmt.Errorf("failed to send start message: %w", err)
//	}
//	return nil
//}
//
//// startCB edits the start message.
//func startCB(b *gotgbot.Bot, ctx *ext.Context) error {
//	cb := ctx.Update.CallbackQuery
//
//	_, err := cb.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
//		Text: "Done!",
//	})
//	if err != nil {
//		return fmt.Errorf("failed to answer start callback query: %w", err)
//	}
//
//	_, _, err = cb.Message.EditText(b, "You edited the start message.", nil)
//	if err != nil {
//		return fmt.Errorf("failed to edit start message text: %w", err)
//	}
//	return nil
//}
//
//func list(b *gotgbot.Bot, ctx *ext.Context) error {
//	reposList, err := database.GetReposList()
//	if err != nil {
//		log.Fatal("failed to send start message: %w", err)
//	} else {
//		_, err := ctx.EffectiveMessage.Reply(b, fmt.Sprintf("%v", reposList), &gotgbot.SendMessageOpts{
//			ParseMode: "html",
//		})
//		if err != nil {
//			return fmt.Errorf("failed to send start message: %w", err)
//		}
//	}
//	return nil
//}
//
//func add(b *gotgbot.Bot, ctx *ext.Context) error {
//	reposList, err := database.GetReposList()
//	if err != nil {
//		log.Fatal("failed to send start message: %w", err)
//	} else {
//		_, err := ctx.EffectiveMessage.Reply(b, fmt.Sprintf("%v", reposList), &gotgbot.SendMessageOpts{
//			ParseMode: "html",
//		})
//		if err != nil {
//			return fmt.Errorf("failed to send start message: %w", err)
//		}
//	}
//	return nil
//}

var (
	// Menu texts
	firstMenu = "<b>Bot menu:</b>"

	// Button texts
	listButton   = "List repos"
	addButton    = "Add repo"
	deleteButton = "Delete repo"
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

	// Wait for a newline symbol, then cancel handling updates
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
		handleMessage(update.Message)
		break

	// Handle messages
	case update.InlineQuery != nil:
		handleInline(update.Message)
		break

	// Handle button clicks
	case update.CallbackQuery != nil:
		handleButton(update.CallbackQuery)
		break
	}
}

func handleMessage(message *tgbotapi.Message) {
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
	} else {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Unknown command! Send /menu")
		_, err = bot.Send(msg)
	}

	if err != nil {
		log.Printf("An error occured: %s", err.Error())
	}
}

func handleInline(message *tgbotapi.Message) {
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
	} else {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Unknown command! Send /menu")
		_, err = bot.Send(msg)
	}

	if err != nil {
		log.Printf("An error occured: %s", err.Error())
	}
}

// When we get a command, we react accordingly
func handleCommand(chatId int64, command string) error {
	var err error

	switch command {
	case "/menu":
		err = sendMenu(chatId)
		break
	}

	return err
}

func handleButton(query *tgbotapi.CallbackQuery) {
	var text string

	markup := tgbotapi.NewInlineKeyboardMarkup()
	message := query.Message

	if query.Data == listButton {
		text = "List repositories"
		//markup = firstMenuMarkup
		reposList, err := database.GetReposList()
		if err != nil {
			log.Fatal("Failed to get repos list: %w", err)
		} else {
			msg := tgbotapi.NewMessage(message.Chat.ID, helpers.ReposListOutput(reposList))
			msg.DisableWebPagePreview = true
			bot.Send(msg)
		}
	} else if query.Data == addButton {
		text = addButton
		msg := tgbotapi.NewMessage(message.Chat.ID, "Send repo link you want to add. Example: https://author/repository")
		bot.Send(msg)
		//markup = firstMenuMarkup
	} else if query.Data == deleteButton {
		text = deleteButton
		msg := tgbotapi.NewMessage(message.Chat.ID, "Send repo link you want to delete. Example: https://author/repository")
		bot.Send(msg)
		//markup = firstMenuMarkup
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
