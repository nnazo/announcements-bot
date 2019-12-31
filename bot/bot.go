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

func (ptr *Bot) LoadConfig() error {
	ptr.Stop = make(chan struct{}, 0)

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

	ptr.Serials = make([]*feed, len(feeds))
	for i, f := range feeds {
		fmt.Println("setup", f.Scraper.URL)
		ptr.Serials[i] = &feed{
			Scraper:     f.Scraper,
			Channels:    f.Channels,
			ArticleType: f.ArticleType,
		}
		ptr.Serials[i].Scraper.Setup()
	}

	return nil
}

func (ptr *Bot) Run() (chan struct{}, error) {
	if ptr.session != nil {
		fmt.Println("opening session")
		err := ptr.session.Open()
		if err != nil {
			return nil, err
		}
		go ptr.scan()
		fmt.Println("bot running")
		return ptr.Stop, nil
	}
	return nil, fmt.Errorf("nil session")
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
			embed := &discordgo.MessageEmbed{
				Color: 174591, // #02a9ff
			}
			switch command {
			case "notifyNewSerials":
				ptr.Serials[newSerial].Channels = append(ptr.Serials[newSerial].Channels, m.ChannelID)
				embed.Title = "Now notifying this channel with new serializations"
				s.ChannelMessageSendEmbed(m.ChannelID, embed)
			case "notifyCompletedSerials":
				ptr.Serials[completedSerial].Channels = append(ptr.Serials[completedSerial].Channels, m.ChannelID)
				embed.Title = "Now notifying this channel with completed serializations"
				s.ChannelMessageSendEmbed(m.ChannelID, embed)
			case "removeNewSerials":
				ptr.Serials[newSerial].removeChannel(m.ChannelID)
				embed.Title = "No longer notifying this channel with new serializations"
				s.ChannelMessageSendEmbed(m.ChannelID, embed)
			case "removeCompletedSerials":
				ptr.Serials[completedSerial].removeChannel(m.ChannelID)
				embed.Title = "No longer notifying this channel with completed serializations"
				s.ChannelMessageSendEmbed(m.ChannelID, embed)
			case "off":
				embed.Title = "Turning off"
				s.ChannelMessageSendEmbed(m.ChannelID, embed)
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
			for _, c := range s.Channels {
				for _, a := range s.Scraper.Articles {
					if !a.Sent {
						m := ptr.findMessage(a, c)
						if m == nil {
							embed := &discordgo.MessageEmbed{
								Color:       174591, // #02a9ff
								URL:         a.URL,
								Title:       a.Title,
								Description: a.Summary,
								Thumbnail: &discordgo.MessageEmbedThumbnail{
									URL: a.Image,
								},
							}

							fmt.Println("\tsending message for", a.URL)
							ptr.session.ChannelMessageSendEmbed(c, embed)
						}
						a.Sent = true
					}
				}
			}
		}
	}
}

func (ptr *Bot) findMessage(a *scraper.Article, channel string) *discordgo.Message {
	messages, err := ptr.session.ChannelMessages(channel, 100, "", "", "")
	if err != nil {
		panic(err)
	}

	for _, m := range messages {
		for _, e := range m.Embeds {
			if e.URL == a.URL {
				return m
			}
		}
	}
	return nil
}
