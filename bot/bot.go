package bot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/nnazo/announcements-bot/scraper"
)

type config struct {
	Token  string `json:"token"`
	BotID  string `json:"id"`
	Prefix string `json:"prefix"`
}

const (
	newSerial       = 0
	completedSerial = iota
)

type feed struct {
	scraper     scraper.Scraper
	Channels    []string `json:"channels"`
	ArticleType int      `json:"type"`
	URL         string   `json:"URL"`
}

func (ptr *feed) removeChannel(channelID string) {
	ndx := -1
	for i, c := range ptr.Channels {
		if channelID == c {
			ndx = i
		}
	}
	if ndx > -1 && len(ptr.Channels) > 0 {
		ptr.Channels[len(ptr.Channels)-1], ptr.Channels[ndx] = ptr.Channels[ndx], ptr.Channels[len(ptr.Channels)-1]
		ptr.Channels = ptr.Channels[:len(ptr.Channels)-1]
	}
}

type Bot struct {
	config  *config
	session *discordgo.Session
	Serials []feed `json:"feeds"`
}

func (ptr *Bot) LoadConfig() error {
	ptr.config = new(config)
	fmt.Println("reading config")
	b, err := ioutil.ReadFile("./bot/config.json")
	if err != nil {
		return err
	}

	fmt.Println("loading config")
	err = json.Unmarshal(b, ptr.config)
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

	fmt.Println("adding handler")
	ptr.session.AddHandler(ptr.messageHandler)
	ptr.config.BotID = user.ID

	fmt.Println("getting feeds")
	b, err = ioutil.ReadFile("./bot/feeds.json")
	if err != nil {
		return err
	}

	feeds := make([]feed, 0)
	err = json.Unmarshal(b, &feeds)
	if err != nil {
		return err
	}

	ptr.Serials = make([]feed, len(feeds))
	for i, f := range feeds {
		fmt.Println("setup", f.URL)
		ptr.Serials[i].scraper.Setup(f.URL)
		ptr.Serials[i].URL = f.URL
		ptr.Serials[i].Channels = f.Channels
		ptr.Serials[i].ArticleType = f.ArticleType
	}

	return nil
}

func (ptr *Bot) Run() error {
	if ptr.session != nil {
		fmt.Println("opening session")
		err := ptr.session.Open()
		if err != nil {
			return err
		}
		go ptr.scan()
		fmt.Println("bot running")
		return nil
	}
	return fmt.Errorf("nil session")
}

func (ptr *Bot) Close() error {
	return ptr.session.Close()
}

func (ptr *Bot) saveFeeds() {
	b, err := json.MarshalIndent(ptr.Serials, "", "    ")
	if err != nil {
		fmt.Println("error marshalling feeds")
		panic(err)
	}
	err = ioutil.WriteFile("./bot/feeds.json", b, 0777)
	if err != nil {
		fmt.Println("error writing to feeds.json")
		panic(err)
	}
}

func (ptr *Bot) messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if strings.HasPrefix(m.Content, ptr.config.Prefix) {
		command := strings.TrimPrefix(m.Content, ptr.config.Prefix)
		if m.Author.ID != ptr.config.BotID {
			switch command {
			case "notifyNewSerials":
				ptr.Serials[newSerial].Channels = append(ptr.Serials[newSerial].Channels, m.ChannelID)
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("> **Added new serialization notifications for this channel**"))
			case "notifyCompletedSerials":
				ptr.Serials[completedSerial].Channels = append(ptr.Serials[completedSerial].Channels, m.ChannelID)
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("> **Added completed serialization notifications for this channel**"))
			case "removeNewSerials":
				ptr.Serials[newSerial].removeChannel(m.ChannelID)
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("> **Removed new serialization notifications for this channel**"))
			case "removeCompletedSerials":
				ptr.Serials[completedSerial].removeChannel(m.ChannelID)
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("> **Removed completed serialization notifications for this channel**"))
			}
			ptr.saveFeeds()
		}
	}
}

func (ptr *Bot) scan() {
	for range time.NewTicker(time.Duration(1) * time.Minute).C {
		for _, s := range ptr.Serials {
			s.scraper.FetchNewArticles()
			for i := 0; i < len(s.scraper.Articles); i++ {
				if !s.scraper.Articles[i].Sent {
					for _, c := range s.Channels {
						var message string

						loc, _ := time.LoadLocation("Japan")
						date := strings.Split(time.Now().In(loc).String(), " ")[0]

						if s.ArticleType == newSerial {
							message = fmt.Sprintf("> **New Serial**: <%v>\n> **Article Title**:%v\n> **Start Date**: %v", s.scraper.Articles[i].URL, s.scraper.Articles[i].Title, date)
						} else {
							message = fmt.Sprintf("> **Completed Serial**: <%v>\n> **Article Title**:%v\n> **End Date**: %v", s.scraper.Articles[i].URL, s.scraper.Articles[i].Title, date)
						}

						fmt.Println("sending message:", message)
						ptr.session.ChannelMessageSend(c, message)
					}
					s.scraper.Articles[i].Sent = true
				}
			}
		}
	}
}
