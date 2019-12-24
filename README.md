# serialization-bot
##### A discord bot for being notified of new and completed manga serializations

## Commands
* Commands have a '!' prefix, but this can be changed in bot/config.json

* `!addNewSerials`
  - The channel the message is received in begins receiving new serialization messages

* `!addCompletedSerials`
  - The channel the message is received in begins receiving completed serialization messages

* `!removeNewSerials`
  - The channel the message is received in stops received new serialization messages

* `!removeCompletedSerials`
  - The channel the message is received in stops receiving completed serialization messages

## Prerequisites
* Must have Go installed
* [discordgo](https://github.com/bwmarrin/discordgo)
* [colly](https://github.com/gocolly/colly)

## Setup
1. Have Go installed on your machine
2. Run `go get github.com/bwmarrin/discordgo`
3. Run `go get -u github.com/gocolly/colly/...`
4. Run `go get github.com/nnazo/serialization-bot`
5. Create a discord application and bot [here](https://discordapp.com/developers/applications/) and add it to server
6. Place the bot's token in bot/config.json
7. To start the bot, run ```go run main.go``` or ```go build main.go``` and run the executable.
