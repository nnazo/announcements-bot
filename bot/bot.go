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
	Scraper     scraper.Scraper `json:"scraper"`
	Channels    []string        `json:"channels"`
	ArticleType int             `json:"type"`
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
	Stop    chan struct{}
	Serials []*feed `json:"feeds"`
}

func (ptr *Bot) LoadConfig() (chan struct{}, error) {
	ptr.Stop = make(chan struct{}, 0)

	ptr.config = new(config)
	fmt.Println("reading config")
	b, err := ioutil.ReadFile("./bot/config.json")
	if err != nil {
		return nil, err
	}

	fmt.Println("loading config")
	err = json.Unmarshal(b, ptr.config)
	if err != nil {
		return nil, err
	}

	fmt.Println("creating session")
	ptr.session, err = discordgo.New("Bot " + ptr.config.Token)
	if err != nil {
		return nil, err
	}

	fmt.Println("getting user")
	user, err := ptr.session.User("@me")
	if err != nil {
		return nil, err
	}

	fmt.Println("adding handler")
	ptr.session.AddHandler(ptr.messageHandler)
	ptr.config.BotID = user.ID

	fmt.Println("getting feeds")
	b, err = ioutil.ReadFile("./bot/feeds.json")
	if err != nil {
		return nil, err
	}

	feeds := make([]feed, 0)
	err = json.Unmarshal(b, &feeds)
	if err != nil {
		return nil, err
	}

	ptr.Serials = make([]*feed, len(feeds))
	for i, f := range feeds {
		fmt.Println("setup", f.Scraper.URL)
		ptr.Serials[i] = &feed{
			Scraper:     f.Scraper,
			Channels:    f.Channels,
			ArticleType: f.ArticleType,
		}
		// ptr.Serials[i].Scraper.URL = f.Scraper.URL
		// ptr.Serials[i].Scraper.MaxArticles = f.Scraper.MaxArticles
		// ptr.Serials[i].Scraper.BufferSize = f.Scraper.BufferSize
		ptr.Serials[i].Scraper.Setup()
		// ptr.Serials[i].Channels = f.Channels
		// ptr.Serials[i].ArticleType = f.ArticleType
	}

	return ptr.Stop, nil
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
			case "off":
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("> **Turning off**"))
				ptr.Stop <- struct{}{}
			}
			ptr.saveFeeds()
		}
	}
}

func (ptr *Bot) scan() {
	for range time.NewTicker(time.Duration(1) * time.Minute).C {
		for _, s := range ptr.Serials {
			s.Scraper.UpdateArticles()
			fmt.Println("post", len(s.Scraper.Articles), s.Scraper.Articles)
			for _, c := range s.Channels {
				for _, a := range s.Scraper.Articles {
					if !a.Sent {
						var message string

						loc, _ := time.LoadLocation("Japan")
						date := strings.Split(time.Now().In(loc).String(), " ")[0]

						switch s.ArticleType {
						case newSerial:
							message = fmt.Sprintf("> **New Serial**: <%v>\n> **Article Title**: %v\n> **Start Date**: %v", a.URL, a.Title, date)
						case completedSerial:
							message = fmt.Sprintf("> **Completed Serial**: <%v>\n> **Article Title**: %v\n> **End Date**: %v", a.URL, a.Title, date)
						default:
							panic("invalid article type")
						}

						fmt.Println("sending message:", message)
						ptr.session.ChannelMessageSend(c, message)
						a.Sent = true
					}
				}
			}
			for _, a := range s.Scraper.Articles {
				if !a.Sent {
					panic(fmt.Sprintf("article sent field is still false %v", a.URL))
				}
			}
		}
	}
}
