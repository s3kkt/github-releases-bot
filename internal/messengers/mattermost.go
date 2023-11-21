package messengers

import (
	"encoding/json"
	"fmt"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/s3kkt/github-releases-bot/internal"
	"github.com/s3kkt/github-releases-bot/internal/database"
	"github.com/s3kkt/github-releases-bot/internal/helpers"
	"github.com/s3kkt/github-releases-bot/internal/transport"
	"log"
	"regexp"
	"time"
)

type MmStruct struct {
	githubRepo                *database.GithubRepos
	mattermostClient          *model.Client4
	mattermostWebSocketClient *model.WebSocketClient
	mattermostUser            *model.User
	mattermostChannel         *model.Channel
	mattermostTeam            *model.Team
}

func NewMM(githubRepo *database.GithubRepos, mattermostClient *model.Client4, mattermostWebSocketClient *model.WebSocketClient, mattermostUser *model.User, mattermostChannel *model.Channel, mattermostTeam *model.Team) *MmStruct {
	return &MmStruct{
		githubRepo:                githubRepo,
		mattermostClient:          mattermostClient,
		mattermostWebSocketClient: mattermostWebSocketClient,
		mattermostUser:            mattermostUser,
		mattermostChannel:         mattermostChannel,
		mattermostTeam:            mattermostTeam,
	}
}

func (r *MmStruct) MmBot(conf internal.Config) {

	// Create a new mattermost client.
	r.mattermostClient = model.NewAPIv4Client(conf.MattermostServer.String())
	// Login.
	r.mattermostClient.SetToken(conf.MattermostToken)

	if user, resp, err := r.mattermostClient.GetUser("me", ""); err != nil {
		log.Fatal("Mattermost: Could not log in")
	} else {
		log.Printf("Mattermost. %s", resp)
		log.Printf("Logged in to mattermost")
		r.mattermostUser = user
	}

	// Find and save the bot's team to app struct.
	if team, resp, err := r.mattermostClient.GetTeamByName(conf.MattermostTeam, ""); err != nil {
		log.Fatal("Could not find team. Is this bot a member ?")
	} else {
		log.Printf("Mattermost. %s", resp)
		r.mattermostTeam = team
	}

	// Find and save the talking channel to app struct.
	if channel, resp, err := r.mattermostClient.GetChannelByName(
		conf.MattermostChannel, r.mattermostTeam.Id, "",
	); err != nil {
		log.Fatal("Could not find channel. Is this bot added to that channel ?")
	} else {
		log.Printf("Mattermost. %s", resp)
		r.mattermostChannel = channel
	}

	// Send a message (new post).
	r.sendMsgToTalkingChannel("Hi! I am a bot.", "")

	// Listen to live events coming in via websocket.
	r.listenToEvents(conf)
}

func (r *MmStruct) sendMsgToTalkingChannel(msg string, replyToId string) {
	// Note that replyToId should be empty for a new post.
	// All replies in a thread should reply to root.

	post := &model.Post{}
	post.ChannelId = r.mattermostChannel.Id
	post.Message = msg

	post.RootId = replyToId

	if _, _, err := r.mattermostClient.CreatePost(post); err != nil {
		log.Printf("Failed to create post. RootID %s", replyToId)
	}
}

func (r *MmStruct) listenToEvents(conf internal.Config) {
	var err error
	failCount := 0
	for {
		r.mattermostWebSocketClient, err = model.NewWebSocketClient4(
			fmt.Sprintf("ws://%s", conf.MattermostServer.Host+conf.MattermostServer.Path),
			r.mattermostClient.AuthToken,
		)
		if err != nil {
			log.Printf("Mattermost websocket disconnected, retrying")
			failCount += 1
			// TODO: backoff based on failCount and sleep for a while.
			continue
		}
		log.Printf("Mattermost websocket connected")

		r.mattermostWebSocketClient.Listen()

		for event := range r.mattermostWebSocketClient.EventChannel {
			// Launch new goroutine for handling the actual event.
			// If required, you can limit the number of events beng processed at a time.
			go r.handleWebSocketEvent(conf, event)
		}
	}
}

func (r *MmStruct) handleWebSocketEvent(conf internal.Config, event *model.WebSocketEvent) {

	// Ignore other channels.
	if event.GetBroadcast().ChannelId != r.mattermostChannel.Id {
		return
	}

	// Ignore other types of events.
	if event.EventType() != model.WebsocketEventPosted {
		return
	}

	// Since this event is a post, unmarshal it to (*model.Post)
	post := &model.Post{}
	err := json.Unmarshal([]byte(event.GetData()["post"].(string)), &post)
	if err != nil {
		log.Printf("Could not cast event to *model.Post")
	}

	// Ignore messages sent by this bot itself.
	if post.UserId == r.mattermostUser.Id {
		return
	}

	// Handle however you want.
	r.handlePost(conf, post)
}

func (r *MmStruct) handlePost(conf internal.Config, post *model.Post) {
	log.Printf("message: %s", post.Message)
	log.Printf("post: %s", post)

	if matched, _ := regexp.MatchString(`(?:^|\W)hello(?:$|\W)`, post.Message); matched {

		// If post has a root ID then its part of thread, so reply there.
		// If not, then post is independent, so reply to the post.
		if post.RootId != "" {
			r.sendMsgToTalkingChannel("I replied in an existing thread.", post.RootId)
		} else {
			r.sendMsgToTalkingChannel("I just replied to a new post, starting a chain.", post.Id)
		}
		return
	}
}

func (r *MmStruct) NotifierMm(conf internal.Config, mattermostChannel int64) {
	duration, _ := time.ParseDuration(conf.UpdateInterval)
	for range time.Tick(duration) {
		log.Print("Bot notifier: check for updates...")
		_, reposList := r.githubRepo.GetChatReposList(mattermostChannel)
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
				ifNew, err := r.githubRepo.CheckIfNew(repo, release.TagName)
				if err != nil {
					continue
				} else if ifNew == true {

					log.Printf("%s is new release for %s, inserting data to database.", release.TagName, repo)
					checkTime := time.Now().Format(time.RFC3339)
					r.githubRepo.InsertReleaseData(checkTime, repo, release, true)

					log.Printf("Try to send updates. Channel: %d", mattermostChannel)
					err := r.sendMm(mattermostChannel, repo, release.TagName, release.HtmlUrl, release.Body)
					if err != nil {
						log.Printf("Failed to send release update. Channel: %d. Reason: %s", mattermostChannel, err)
						return
					}
				}
			}
		}
	}
	return
}

func (r *MmStruct) sendMm(chatId int64, repoName, tag, url, releaseNotes string) error {
	return nil
}
