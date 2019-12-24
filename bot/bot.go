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
	channels    []string
	articleType int
}

func (ptr *feed) removeChannel(channelID string) {
	ndx := -1
	for i, c := range ptr.channels {
		if channelID == c {
			ndx = i
		}
	}
	if ndx > -1 && len(ptr.channels) > 0 {
		ptr.channels[len(ptr.channels)-1], ptr.channels[ndx] = ptr.channels[ndx], ptr.channels[len(ptr.channels)-1]
		ptr.channels = ptr.channels[:len(ptr.channels)-1]
	}
}

type Bot struct {
	config  *config
	session *discordgo.Session
	serials []feed
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

	fmt.Println("adding handler")
	ptr.session.AddHandler(ptr.messageHandler)

	ptr.config.BotID = user.ID

	feeds := []struct {
		URL  string
		Type int
	}{
		{"https://natalie.mu/comic/tag/43", newSerial},
		{"https://natalie.mu/comic/tag/42", completedSerial},
	}

	ptr.serials = make([]feed, len(feeds))

	for i, f := range feeds {
		ptr.serials[i].scraper.Setup(f.URL)
		ptr.serials[i].channels = make([]string, 0)
		ptr.serials[i].articleType = f.Type
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

func (ptr *Bot) messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if strings.HasPrefix(m.Content, ptr.config.Prefix) {
		command := strings.TrimPrefix(m.Content, ptr.config.Prefix)
		if m.Author.ID != ptr.config.BotID {
			switch command {
			case "addNewSerials":
				ptr.serials[newSerial].channels = append(ptr.serials[newSerial].channels, m.ChannelID)
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("> **Added new serialization notifications for this channel**"))
			case "addCompletedSerials":
				ptr.serials[completedSerial].channels = append(ptr.serials[completedSerial].channels, m.ChannelID)
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("> **Added completed serialization notifications for this channel**"))
			case "removeNewSerials":
				ptr.serials[newSerial].removeChannel(m.ChannelID)
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("> **Removed new serialization notifications for this channel**"))
			case "removeCompletedSerials":
				ptr.serials[completedSerial].removeChannel(m.ChannelID)
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("> **Removed completed serialization notifications for this channel**"))
			}
		}
	}
}

func (ptr *Bot) scan() {
	for range time.NewTicker(time.Duration(1) * time.Minute).C {
		for _, s := range ptr.serials {
			fmt.Println("scraping..")
			articles := s.scraper.FetchNewArticles()
			for _, a := range articles {
				for _, c := range s.channels {
					var message string
					if s.articleType == newSerial {
						message = fmt.Sprintf("> **New Serial**: <%v>\n**Article Title**:%v\n**Start Date**: %v", a.URL, a.Title, a.Date)
					} else {
						message = fmt.Sprintf("> **Completed Serial**: <%v>\n**Article Title**:%v\n**End Date**: %v", a.URL, a.Title, a.Date)
					}
					ptr.session.ChannelMessageSend(c, message)
				}
			}
		}
	}
}
