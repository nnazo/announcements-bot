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

## Setup
1. Create a discord bot and add it to server
2. Place the bot's token in bot/config.json
3. To start the bot, run ```go run main.go``` or ```go build main.go``` and run the executable.
