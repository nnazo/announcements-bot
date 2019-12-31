package main

import (
	"log"

	"github.com/nnazo/announcements-bot/bot"
)

func main() {
	var bot bot.Bot
	err := bot.LoadConfig()
	if err != nil {
		log.Fatalln(err.Error())
	}
	stop, err := bot.Run()
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer bot.Close()
	<-stop
	return
}
