# announcements-bot
##### A discord bot for being notified of new and completed manga serializations

## Commands
* Commands have a '!' prefix, but this can be changed in bot/config.json

* `!notifyNewSerials`
  - The channel the message is received in begins receiving new serialization messages

* `!notifyCompletedSerials`
  - The channel the message is received in begins receiving completed serialization messages

* `!removeNewSerials`
  - The channel the message is received in stops received new serialization messages

* `!removeCompletedSerials`
  - The channel the message is received in stops receiving completed serialization messages

* `!off`
  - Turns the bot off

## Installation
1. Have Go installed on your machine
2. Run `go get github.com/nnazo/announcements-bot`
3. Create a discord application and bot [here](https://discordapp.com/developers/applications/) and add it to your server
4. Place the bot's token in bot/config.json
5. To start the bot, run ```go run main.go``` or ```go build main.go``` and run the executable.
