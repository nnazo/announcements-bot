package main

import (
	"log"

	"github.com/nnazo/announcements-bot/bot"
)

func main() {
	var bot bot.Bot
	stop, err := bot.LoadConfig()
	if err != nil {
		log.Fatalln(err.Error())
	}
	err = bot.Run()
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer bot.Close()
	<-stop
	return
}
