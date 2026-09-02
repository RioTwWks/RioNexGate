package telegram

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"rionexgate/internal/config"
	"rionexgate/internal/core"
	"rionexgate/internal/db"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	api  *tgbotapi.BotAPI
	db   *db.DB
	core core.Manager
	cfg  *config.Config
}

func New(cfg *config.Config, database *db.DB, coreMgr core.Manager) (*Bot, error) {
	if cfg.Telegram.BotToken == "" || cfg.Telegram.BotToken == "YOUR_BOT_TOKEN" {
		return nil, fmt.Errorf("telegram bot token not configured")
	}
	api, err := tgbotapi.NewBotAPI(cfg.Telegram.BotToken)
	if err != nil {
		return nil, err
	}
	return &Bot{api: api, db: database, core: coreMgr, cfg: cfg}, nil
}

func (b *Bot) Start() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.api.GetUpdatesChan(u)
	log.Println("telegram bot started")
	for update := range updates {
		if update.Message == nil || !update.Message.IsCommand() {
			continue
		}
		if !b.isAdmin(update.Message.From.ID) {
			b.send(update.Message.Chat.ID, "Access denied.")
			continue
		}
		b.handleCommand(update.Message)
	}
}

func (b *Bot) isAdmin(id int64) bool {
	for _, admin := range b.cfg.Telegram.AdminIDs {
		if admin == id {
			return true
		}
	}
	return false
}

func (b *Bot) handleCommand(msg *tgbotapi.Message) {
	args := strings.Fields(msg.CommandArguments())
	switch msg.Command() {
	case "start":
		b.handleStart(msg.Chat.ID)
	case "users":
		b.handleUsers(msg.Chat.ID)
	case "add":
		b.handleAddUser(msg.Chat.ID, args)
	case "link":
		b.handleLink(msg.Chat.ID, args)
	case "traffic":
		b.handleTraffic(msg.Chat.ID, args)
	case "reload":
		b.handleReload(msg.Chat.ID)
	default:
		b.send(msg.Chat.ID, "Unknown command. Try /start")
	}
}

func (b *Bot) handleStart(chatID int64) {
	msg := "Proxy Manager Bot\n\nCommands:\n/users - list users\n/add <email> [traffic_gb] [expire_days]\n/link <user_id>\n/traffic <user_id>\n/reload - reload core config"
	b.send(chatID, msg)
}

func (b *Bot) handleUsers(chatID int64) {
	users, err := b.db.ListUsers()
	if err != nil {
		b.send(chatID, "Error: "+err.Error())
		return
	}
	if len(users) == 0 {
		b.send(chatID, "No users yet.")
		return
	}
	var sb strings.Builder
	sb.WriteString("Users:\n")
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, u := range users {
		usedGB := float64(u.UsedBytes) / (1024 * 1024 * 1024)
		active := "yes"
		if !u.Active {
			active = "no"
		}
		sb.WriteString(fmt.Sprintf("#%d %s — %.2f/%d GB, active: %s\n", u.ID, u.Email, usedGB, u.TrafficGB, active))
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("#%d %s", u.ID, u.Email), fmt.Sprintf("user:%d", u.ID)),
		))
	}
	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
	msg := tgbotapi.NewMessage(chatID, sb.String())
	msg.ReplyMarkup = keyboard
	if _, err := b.api.Send(msg); err != nil {
		log.Printf("telegram send: %v", err)
	}
}

func (b *Bot) handleAddUser(chatID int64, args []string) {
	if len(args) < 1 {
		b.send(chatID, "Usage: /add <email> [traffic_gb] [expire_days]")
		return
	}
	trafficGB := b.cfg.Limits.DefaultTrafficGB
	expireDays := b.cfg.Limits.DefaultExpireDays
	if len(args) >= 2 {
		if v, err := strconv.ParseInt(args[1], 10, 64); err == nil {
			trafficGB = v
		}
	}
	if len(args) >= 3 {
		if v, err := strconv.Atoi(args[2]); err == nil {
			expireDays = v
		}
	}
	user, err := b.db.CreateUser(db.CreateUserInput{
		Email:      args[0],
		TrafficGB:  trafficGB,
		ExpireDays: expireDays,
	})
	if err != nil {
		b.send(chatID, "Error: "+err.Error())
		return
	}
	if err := b.core.Reload(); err != nil {
		b.send(chatID, "User created but reload failed: "+err.Error())
		return
	}
	link, _ := b.core.GetClientLink(strconv.FormatUint(uint64(user.ID), 10), "vless")
	b.send(chatID, fmt.Sprintf("User #%d created.\n%s", user.ID, link))
}

func (b *Bot) handleLink(chatID int64, args []string) {
	if len(args) < 1 {
		b.send(chatID, "Usage: /link <user_id>")
		return
	}
	link, err := b.core.GetClientLink(args[0], "vless")
	if err != nil {
		b.send(chatID, "Error: "+err.Error())
		return
	}
	b.send(chatID, link)
}

func (b *Bot) handleTraffic(chatID int64, args []string) {
	if len(args) < 1 {
		b.send(chatID, "Usage: /traffic <user_id>")
		return
	}
	id, err := strconv.ParseUint(args[0], 10, 64)
	if err != nil {
		b.send(chatID, "Invalid user id")
		return
	}
	user, err := b.db.GetUser(uint(id))
	if err != nil {
		b.send(chatID, "User not found")
		return
	}
	usedGB := float64(user.UsedBytes) / (1024 * 1024 * 1024)
	b.send(chatID, fmt.Sprintf("User #%d %s: %.2f / %d GB used", user.ID, user.Email, usedGB, user.TrafficGB))
}

func (b *Bot) handleReload(chatID int64) {
	if err := b.core.Reload(); err != nil {
		b.send(chatID, "Reload failed: "+err.Error())
		return
	}
	b.send(chatID, "Core config reloaded.")
}

func (b *Bot) send(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := b.api.Send(msg); err != nil {
		log.Printf("telegram send: %v", err)
	}
}
