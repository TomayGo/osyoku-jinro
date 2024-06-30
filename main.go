package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// Variables used for command line parameters
var (
	Token        string
	state        int
	participants []*discordgo.User
)

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()
	state = 0
}

func startGame(s *discordgo.Session, m *discordgo.MessageCreate) {
	state = 1
	msg, err := s.ChannelMessageSend(m.ChannelID, "ゲームを開始します。参加者はリアクションを押してください。")
	if err != nil {
		fmt.Println("Error sending message:", err)
	}

	// Add reaction to the message
	err = s.MessageReactionAdd(m.ChannelID, msg.ID, "👍")
	if err != nil {
		fmt.Println("Error adding reaction:", err)
		return
	}

	// リアクションが追加されたイベントを捕捉
	s.AddHandler(func(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
		if r.MessageID == msg.ID && r.UserID != s.State.User.ID {
			if r.Emoji.Name == "✅" {
				// ✅リアクションが検出されたら、参加者のカウントを停止
				fmt.Println("参加者のカウントを停止します。")
				state = 2 // ゲームの状態を更新して参加者の追加を停止
				fmt.Println("参加者のリストを表示します。")
				for _, p := range participants {
					fmt.Println(p.Username)
				}
				return
			}

			if state == 1 { // 参加者のカウントが進行中の場合のみユーザを追加
				user, err := s.User(r.UserID)
				if err != nil {
					fmt.Println("Error getting user:", err)
					return
				}
				participants = append(participants, user)
			}
		}
	})
}

func main() {

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// Add event handler for message create events
	dg.AddHandler(messageCreate)

	// Open a websocket connection to Discord
	err = dg.Open()
	if err != nil {
		fmt.Println("Error opening connection:", err)
		return
	}

	// Wait here until CTRL-C or other termination signal is received
	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	<-make(chan struct{})
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	fmt.Println("Message received:", m.Content)
	// Ignore messages sent by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	if strings.Contains(m.Content, s.State.User.Mention()) && strings.Contains(m.Content, "stop") {
		state = 0
		participants = nil
		s.ChannelMessageSend(m.ChannelID, "ゲームを終了しました。")
	}

	if strings.Contains(m.Content, s.State.User.Mention()) && strings.Contains(m.Content, "start") {
		if state == 0 {
			startGame(s, m)
		} else {
			s.ChannelMessageSend(m.ChannelID, "既存のゲームが存在します。終了してから再度実行してください。")
		}
	}
}
