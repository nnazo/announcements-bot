package main

import (
	"fmt"

	"github.com/nnazo/discord-bot/bot"
)

func main() {
	var bot bot.Bot
	err := bot.LoadConfig()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	err = bot.Run()
	if err != nil {
		fmt.Println(err.Error())
	}
	<-make(chan struct{})
	return
}
