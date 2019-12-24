package main

import (
	"log"

	"github.com/nnazo/serialization-bot/bot"
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
	defer bot.Close()
	<-make(chan struct{})
	return
}
