package main

import (
	"log"

	"github.com/nnazo/discord-bot/bot"
)

func main() {
	var bot bot.Bot
	err := bot.LoadConfig()
	if err != nil {
		log.Fatalln(err.Error())
	}
	err = bot.Run()
	if err != nil {
		log.Fatalln(err.Error())
	}
	<-make(chan struct{})
	return
}
