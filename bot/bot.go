package bot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/nnazo/discord-bot/scraper"
)

type config struct {
	Token  string `json:"token"`
	BotID  string `json:"id"`
	Prefix string `json:"prefix"`
}

type Bot struct {
	config  *config
	session *discordgo.Session
	scraper scraper.Scraper
}

func (ptr *Bot) LoadConfig() error {
	ptr.config = new(config)
	fmt.Println("reading config")
	b, err := ioutil.ReadFile("./bot/config.json")
	if err != nil {
		return err
	}

	fmt.Println("loading config")
	json.Unmarshal(b, ptr.config)
	if err != nil {
		return err
	}

	fmt.Println("creating session")
	ptr.session, err = discordgo.New("Bot " + ptr.config.Token)
	if err != nil {
		return err
	}

	fmt.Println("getting user")
	user, err := ptr.session.User("@me")
	if err != nil {
		return err
	}

	ptr.config.BotID = user.ID

	ptr.scraper.Setup()

	return nil
}

func (ptr *Bot) Run() error {
	if ptr.session != nil {
		fmt.Println("adding handler")
		ptr.session.AddHandler(ptr.messageHandler)
		fmt.Println("opening session")
		err := ptr.session.Open()
		if err != nil {
			return err
		}

		fmt.Println("bot running")
		return nil
	}
	return fmt.Errorf("nil session")
}

func (ptr *Bot) messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if strings.HasPrefix(m.Content, ptr.config.Prefix) {
		command := strings.TrimPrefix(m.Content, ptr.config.Prefix)
		if m.Author.ID != ptr.config.BotID {
			switch command {
			case "test":
				s.ChannelMessageSend(m.ChannelID, "test message here")
			case "start":
				// start scraping
			case "stop":
				// stop scraping
			}
		}
	}
}
