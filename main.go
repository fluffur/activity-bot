package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/celestix/gotgproto"
	"github.com/celestix/gotgproto/dispatcher"
	"github.com/celestix/gotgproto/dispatcher/handlers"
	"github.com/celestix/gotgproto/dispatcher/handlers/filters"
	"github.com/celestix/gotgproto/ext"
	"github.com/celestix/gotgproto/sessionMaker"
	"github.com/glebarez/sqlite"
	"github.com/gotd/td/telegram/message/entity"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	appIDStr := os.Getenv("APP_ID")
	appID, err := strconv.Atoi(appIDStr)
	if err != nil {
		log.Fatal(err)
	}

	client, err := gotgproto.NewClient(
		// Get AppID from https://my.telegram.org/apps
		appID,
		// Get ApiHash from https://my.telegram.org/apps
		os.Getenv("APP_HASH"),
		// ClientType, as we defined above
		gotgproto.ClientTypeBot(os.Getenv("BOT_TOKEN")),
		// Optional parameters of client
		&gotgproto.ClientOpts{
			Session: sessionMaker.SqlSession(sqlite.Open("flood_cm")),
		},
	)
	if err != nil {
		log.Fatalln("failed to start client:", err)
	}

	dp := client.Dispatcher

	// Command Handler for /start
	dp.AddHandler(handlers.NewCommand("start", start))
	// This Message Handler will call our echo function on new messages
	dp.AddHandlerToGroup(handlers.NewMessage(filters.Message.Text, echo), 1)

	fmt.Printf("client (@%s) has been started...\n", client.Self.Username)

	if err := client.Idle(); err != nil {
		log.Fatal(err)
	}
}

// callback function for /start command
func start(ctx *ext.Context, update *ext.Update) error {
	_, _ = ctx.Reply(update, ext.ReplyTextStyledText(styling.Custom(func(eb *entity.Builder) error {
		eb.Plain("Hello World!\n")
		eb.CustomEmoji("❤", 5411197345968701560)

		return nil
	})), nil)

	// End dispatcher groups so that bot doesn't echo /start command usage
	return dispatcher.EndGroups
}

func echo(ctx *ext.Context, update *ext.Update) error {
	msg := update.EffectiveMessage
	_, err := ctx.Reply(update, ext.ReplyTextString(msg.Text), nil)
	return err
}
