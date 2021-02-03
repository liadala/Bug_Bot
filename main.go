package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aichaos/rivescript-go"
	"github.com/aichaos/rivescript-go/lang/javascript"
	"github.com/aichaos/rivescript-go/sessions/memory"
	"github.com/bwmarrin/discordgo"
)

var bot *rivescript.RiveScript

//BotID ..
var BotID string

var (
	dcBotName  string
	dcBotToken string

	config *configStruct
)

type configStruct struct {
	Token         string `json:"Token"`
	ListenChannel string `json:"ListenChannel"`
}

func init() {
	fmt.Println("Read Bot Config")
	file, err := ioutil.ReadFile("./config.json")
	if err != nil {
		fmt.Printf("[ERROR]: %s\n", err)
	}
	err = json.Unmarshal(file, &config)
	if err != nil {
		fmt.Printf("[ERROR]: %s\n", err)
		return
	}
}

func main() {
	bot = rivescript.New(&rivescript.Config{
		Debug:          false,
		Strict:         false,
		UTF8:           false,
		Depth:          50,
		Seed:           time.Now().UnixNano(),
		SessionManager: memory.New(),
	})
	//bot = rivescript.New(rivescript.WithUTF8())
	jsHandler := javascript.New(bot)
	bot.SetHandler("javascript", jsHandler)
	err := bot.LoadDirectory("brain")
	if err != nil {
		fmt.Printf("[ERROR]: %s\n", err)
	}

	bot.SortReplies()

	discord, err := discordgo.New("Bot " + config.Token)
	if err != nil {
		fmt.Printf("[ERROR]: %s\n", err)
	}

	discordUser, err := discord.User("@me")
	if err != nil {
		fmt.Printf("[ERROR]: %s\n", err)
	}

	discord.AddHandler(messageHandler)
	BotID = discordUser.ID

	err = discord.Open()
	if err != nil {
		fmt.Printf("[ERROR]: %s\n", err)
		return
	}

	fmt.Println("Bot is running.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGTERM, syscall.SIGINT, os.Interrupt, os.Kill)
	<-sc
	discord.Close()
}

// messageCreate Auto Reply from Brain
func messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {

	if m.ChannelID != config.ListenChannel {
		return
	}

	if m.Author.ID == s.State.User.ID {
		go time.AfterFunc(10*time.Minute, func() {
			err := s.ChannelMessageDelete(m.ChannelID, m.ID)
			if err != nil {
				fmt.Printf("[ERROR]: %s\n", err)
			}
		})
		return
	}

	reply, err := bot.Reply(m.Author.ID, m.Content)
	if err != nil {
		fmt.Printf("[ERROR]: %s\n", err)
	} else {
		s.ChannelMessageSend(m.ChannelID, reply)
		go time.AfterFunc(10*time.Minute, func() {
			err := s.ChannelMessageDelete(m.ChannelID, m.ID)
			if err != nil {
				fmt.Printf("[ERROR]: %s\n", err)
				return
			}
		})
	}
	fmt.Println(bot.GetUservars(m.Author.ID))
}
