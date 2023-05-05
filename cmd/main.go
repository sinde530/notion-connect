package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/jomei/notionapi"
)

var (
	Token       string
	NotionToken string
	DatabaseID  string
	DiscordID   string
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
		return
	}

	Token = os.Getenv("DISCORD_TOKEN")
	NotionToken = os.Getenv("NOTION_TOKEN")
	DatabaseID = os.Getenv("DATABASE_ID")
	DiscordID = os.Getenv("DISCORD_CHANNEL_ID")

	fmt.Println("Discord token:", Token)

	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	dg.AddHandler(messageCreate)

	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	defer dg.Close()

	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	<-make(chan struct{})
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	// if m.Author.ID == s.State.User.ID || m.ChannelID != ChannelID {
	// 	return
	// }

	if strings.HasPrefix(m.Content, "/help") {
		helpMessage := "Help:\n" +
			"Snowflake is a free service.\n" +
			"Use /create {statusName} {ticketName} to create a ticket in Notion."
		s.ChannelMessageSend(m.ChannelID, helpMessage)
		return
	}

	if strings.HasPrefix(m.Content, "/create") {
		args := strings.Split(m.Content, " ")
		// if len(args) != 4 {
		if len(args) < 3 {
			s.ChannelMessageSend(m.ChannelID, "Invalid command format. Please use: /create {statusName} {ticketName}")
			return
		}

		statusName := args[1]
		// ticketName := args[2] + " " + args[3]
		ticketName := strings.Join(args[2:], " ")

		err := createNotionTicket(statusName, ticketName)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error creating ticket: "+err.Error())
			return
		}

		s.ChannelMessageSend(m.ChannelID, "Ticket successfully created.")
	}
}

func createNotionTicket(statusName string, ticketName string) error {
	client := notionapi.NewClient(notionapi.Token(NotionToken))

	ctx := context.Background()

	// Retrieve the database using the GetDatabase method
	database, err := client.Database.Get(ctx, notionapi.DatabaseID(DatabaseID))
	if err != nil {
		return err
	}

	_, ok := database.Properties["Status"]
	if !ok {
		return fmt.Errorf("no status property found")
	}

	page := notionapi.Page{
		Properties: notionapi.Properties{
			"Name": &notionapi.TitleProperty{
				Title: []notionapi.RichText{
					{
						Text: &notionapi.Text{
							Content: ticketName,
						},
					},
				},
			},
			"Status": &notionapi.SelectProperty{
				Select: notionapi.Option{
					Name: statusName,
				},
			},
		},
	}

	_, err = client.Page.Create(ctx, &notionapi.PageCreateRequest{
		Parent: notionapi.Parent{
			DatabaseID: notionapi.DatabaseID(DatabaseID),
		},
		Properties: page.Properties,
	})

	if err != nil {
		return err
	}

	return nil
}
