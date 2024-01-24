package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

const (
	maxPower  uint8 = 240
	regenRate uint8 = 6
)

func main() {
	loadErr := godotenv.Load()
	if loadErr != nil {
		log.Fatal("Error loading .env file")
	}
	discordToken := os.Getenv("TOKEN")
	session, sessionErr := discordgo.New("Bot " + discordToken)
	if sessionErr != nil {
		log.Fatal("error creating Discord session,", sessionErr)
	}
	session.AddHandler(messageCreate)
	session.Identify.Intents = discordgo.IntentsGuildMessages
	openErr := session.Open()
	if openErr != nil {
		fmt.Println("error opening connection,", openErr)
		return
	}
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-signalChan
	session.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	content := strings.TrimSpace(m.Content)
	fmt.Println("Received message:", content)

	if strings.Contains(content, "на дабл") {
		handleDoubleCommand(s, m, content)
	} else if strings.Contains(content, "клара ресет") {
		handleResetCommand(s, m, content)
	} else if match, _ := regexp.MatchString(`^ех(\.+)?$`, content); match {
		handleExhCommand(s, m, content)
	} else if strings.HasPrefix(strings.ToLower(content), "клара напомни") {
		handleReminderCommand(s, m, content)
	}
}

func handleDoubleCommand(s *discordgo.Session, m *discordgo.MessageCreate, content string) {
	phrase := strings.TrimSpace(strings.Replace(content, "на дабл", "", 1))

	randomNumber := rand.Intn(101)

	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("🎲 **%d** %s", randomNumber, phrase))
	if err != nil {
		fmt.Println("Error sending message:", err)
	}
}

func handleResetCommand(s *discordgo.Session, m *discordgo.MessageCreate, content string) {
	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Ресет на европе <t:%d:R>", GetNextDayTimestamp()))
	if err != nil {
		fmt.Println("Error sending message:", err)
	}
}

func handleExhCommand(s *discordgo.Session, m *discordgo.MessageCreate, content string) {
	match, _ := regexp.MatchString(`^ех(\.+)?$`, m.Content)
	if match {
		_, err := s.ChannelMessageSend(m.ChannelID, "тяжело... тяжело...")
		if err != nil {
			fmt.Println("Error sending message:", err)
		}
	}
}

func handleReminderCommand(s *discordgo.Session, m *discordgo.MessageCreate, content string) {
	command := strings.ReplaceAll(content, "клара напомни", "")
	currentStamina, err := parseStaminaNumber(command)
	if err != nil {
		response := fmt.Sprintf("Сестричка, формат команды должен быть такой: `клара напомни <число 0-%d>`", maxPower)
		_, err = s.ChannelMessageSend(m.ChannelID, response)
		if err != nil {
			log.Printf("Error sending message: %s", err)
		}
		return
	}

	timestamp := regenTimestamp(remainingPower(currentStamina))
	response := fmt.Sprintf("Сестричка, твоя стамина полностью заполнится <t:%d:R>\n", timestamp)
	_, err = s.ChannelMessageSend(m.ChannelID, response)
	if err != nil {
		log.Printf("Error sending message: %s", err)
		return
	}

	duration := time.Until(time.Unix(timestamp, 0))

	time.AfterFunc(duration, func() {
		reminderMessage := fmt.Sprintf("Сестричка, <@%s>, твоя стамина полностью восстановлена!", m.Author.ID)
		_, err = s.ChannelMessageSendReply(m.ChannelID, reminderMessage, m.Reference())
		if err != nil {
			log.Printf("Error sending reminder message: %s", err)
		}
	})
}

func remainingPower(currentPower uint8) uint8 {
	return maxPower - currentPower
}

func regenTimestamp(remainingPower uint8) int64 {
	regenDuration := time.Duration(remainingPower*regenRate) * time.Minute
	timestamp := time.Now().Add(regenDuration).Unix()
	return timestamp
}

func GetNextDayTimestamp() int64 {
	currentTime := time.Now().UTC()

	// Set the location to Moscow time zone (UTC+3)
	loc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		fmt.Println("Error loading location:", err)
		return 0
	}

	// Set the current time to the desired time in Moscow
	currentTime = currentTime.In(loc)

	// Check if it's between 12:00 PM and 5:59 AM
	if currentTime.Hour() >= 12 || currentTime.Hour() < 6 {
		desiredTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 6, 0, 0, 0, loc)
		timestamp := desiredTime.Unix()
		return timestamp
	}

	currentTime = currentTime.AddDate(0, 0, 1)
	desiredTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 6, 0, 0, 0, loc)

	timestamp := desiredTime.Unix()

	return timestamp
}

func parseStaminaNumber(command string) (uint8, error) {
	re := regexp.MustCompile(`\d+`)
	match := re.FindString(command)

	stamina, err := strconv.Atoi(match)
	if err != nil {
		return 0, err
	}

	if stamina < 0 || stamina > int(maxPower) {
		return 0, fmt.Errorf(`stamina number out of range (0-%d)`, maxPower)
	}

	return uint8(stamina), nil
}
