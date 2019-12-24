package bot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/nnazo/discord-bot/scraper"
)

type config struct {
	Token  string `json:"token"`
	BotID  string `json:"id"`
	Prefix string `json:"prefix"`
}

type Bot struct {
	config   *config
	session  *discordgo.Session
	scraper  scraper.Scraper
	channels []string
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
	ptr.channels = make([]string, 0)

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
		go ptr.scrape()
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
			case "add":
				ptr.channels = append(ptr.channels, m.ChannelID)
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("> **Added new serialization notifications for this channel**"))
			case "remove":
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("> **Removed new serialization notifications for this channel**"))
				ndx := -1
				for i, c := range ptr.channels {
					if m.ChannelID == c {
						ndx = i
					}
				}
				if ndx > -1 && len(ptr.channels) > 0 {
					ptr.channels[len(ptr.channels)-1], ptr.channels[ndx] = ptr.channels[ndx], ptr.channels[len(ptr.channels)-1]
					ptr.channels = ptr.channels[:len(ptr.channels)-1]
				}
			}
		}
	}
}

func (ptr *Bot) scrape() {
	for range time.NewTicker(time.Duration(1) * time.Minute).C {
		fmt.Println("scraping..")
		articles := ptr.scraper.FetchNewArticles()
		for _, a := range articles {
			for _, c := range ptr.channels {
				ptr.session.ChannelMessageSend(c, a.URL)
			}
		}
	}
}
