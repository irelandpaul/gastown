package telegram

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/gofrs/flock"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/steveyegge/gastown/internal/mail"
	"github.com/steveyegge/gastown/internal/tui/inbox"
)

// Daemon handles the two-way bridge between Gas Town and Telegram.
type Daemon struct {
	config   *Config
	state    *State
	bot      *tgbotapi.BotAPI
	townRoot string
	logger   *log.Logger
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewDaemon creates a new Telegram daemon.
func NewDaemon(townRoot string, cfg *Config, logger *log.Logger) (*Daemon, error) {
	bot, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		return nil, fmt.Errorf("creating telegram bot: %w", err)
	}

	statePath := filepath.Join(townRoot, "daemon", "telegram.state")
	state, err := LoadState(statePath)
	if err != nil {
		return nil, fmt.Errorf("loading telegram state: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Daemon{
		config:   cfg,
		state:    state,
		bot:      bot,
		townRoot: townRoot,
		logger:   logger,
		ctx:      ctx,
		cancel:   cancel,
	},
	nil
}

// Run starts the daemon polling loops.
func (d *Daemon) Run() error {
	d.logger.Printf("Telegram daemon starting (Bot: %s, ChatID: %d, PID: %d)", d.bot.Self.UserName, d.config.ChatID, os.Getpid())

	// Acquire exclusive lock
	lockFile := filepath.Join(d.townRoot, "daemon", "telegram.lock")
	fileLock := flock.New(lockFile)
	locked, err := fileLock.TryLock()
	if err != nil {
		return fmt.Errorf("acquiring lock: %w", err)
	}
	if !locked {
		return fmt.Errorf("telegram daemon already running (lock held)")
	}
	defer func() { _ = fileLock.Unlock() }()

	// Write PID file
	pidFile := filepath.Join(d.townRoot, "daemon", "telegram.pid")
	if err := os.MkdirAll(filepath.Dir(pidFile), 0755); err != nil {
		return fmt.Errorf("creating daemon directory: %w", err)
	}
	if err := os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", os.Getpid())), 0644); err != nil {
		return fmt.Errorf("writing PID file: %w", err)
	}
	defer os.Remove(pidFile)

	// Poll intervals
	inboxPollInterval := 10 * time.Second

	inboxTicker := time.NewTicker(inboxPollInterval)
	defer inboxTicker.Stop()

	// Initial poll
	d.pollInbox()

	// Telegram updates channel
	u := tgbotapi.NewUpdate(d.state.LastUpdateID + 1)
	u.Timeout = 60
	updates := d.bot.GetUpdatesChan(u)

	for {
		select {
		case <-d.ctx.Done():
			d.logger.Println("Telegram daemon shutting down")
			return nil

		case <-inboxTicker.C:
			d.pollInbox()

		case update := <-updates:
			d.handleTelegramUpdate(update)
		}
	}
}

// pollInbox checks the overseer inbox for new messages and sends them to Telegram.
func (d *Daemon) pollInbox() {
	mailbox := mail.NewMailboxFromAddress("overseer", d.townRoot)
	messages, err := mailbox.List()
	if err != nil {
		d.logger.Printf("Error listing inbox: %v", err)
		return
	}

	for _, msg := range messages {
		// Skip if already sent to Telegram
		// We use a mapping in state to track what's been sent
		if d.isAlreadySent(msg.ID) {
			continue
		}

		d.sendToTelegram(msg)
	}
}

func (d *Daemon) isAlreadySent(beadsID string) bool {
	d.state.mu.RLock()
	defer d.state.mu.RUnlock()
	for _, id := range d.state.MsgMap {
		if id == beadsID {
			return true
		}
	}
	return false
}

// sendToTelegram sends a Gas Town message to the configured Telegram chat.
func (d *Daemon) sendToTelegram(msg *mail.Message) {
	text := FormatMessage(msg)

	tgMsg := tgbotapi.NewMessage(d.config.ChatID, text)
	tgMsg.ParseMode = tgbotapi.ModeMarkdownV2

	sent, err := d.bot.Send(tgMsg)
	if err != nil {
		d.logger.Printf("Error sending to Telegram: %v", err)
		return
	}

	// Record mapping and save state
	d.state.AddMapping(sent.MessageID, msg.ID)
	if err := d.state.Save(); err != nil {
		d.logger.Printf("Error saving state: %v", err)
	}

	d.logger.Printf("Sent message %s to Telegram (TG ID: %d)", msg.ID, sent.MessageID)
}

// handleTelegramUpdate processes an update from Telegram.
func (d *Daemon) handleTelegramUpdate(update tgbotapi.Update) {
	if update.UpdateID > d.state.LastUpdateID {
		d.state.SetLastUpdateID(update.UpdateID)
		_ = d.state.Save()
	}

	if update.Message == nil {
		return
	}

	// Only process messages from the configured chat
	if update.Message.Chat.ID != d.config.ChatID {
		return
	}

	// Handle replies
	if update.Message.ReplyToMessage != nil {
		d.handleReply(update.Message)
		return
	}

	// Handle commands if any (e.g. /status)
	if update.Message.IsCommand() {
		d.handleCommand(update.Message)
		return
	}
}

func (d *Daemon) handleReply(tgMsg *tgbotapi.Message) {
	originalID := d.state.GetBeadsID(tgMsg.ReplyToMessage.MessageID)
	if originalID == "" {
		d.logger.Printf("Received reply to unknown Telegram message: %d", tgMsg.ReplyToMessage.MessageID)
		return
	}

	mailbox := mail.NewMailboxFromAddress("overseer", d.townRoot)
	original, err := mailbox.Get(originalID)
	if err != nil {
		d.logger.Printf("Error getting original message %s: %v", originalID, err)
		return
	}

	text := strings.TrimSpace(tgMsg.Text)
	msgType := inbox.InferMessageType(original)

	if msgType == inbox.TypeProposal {
		lower := strings.ToLower(text)
		if lower == "y" || lower == "yes" || lower == "approve" || lower == "ok" {
			d.logger.Printf("Approving proposal %s via Telegram", originalID)
			d.approveProposal(original)
			return
		} else if lower == "n" || lower == "no" || lower == "reject" {
			d.logger.Printf("Rejecting proposal %s via Telegram", originalID)
			d.rejectProposal(original)
			return
		}
	}

	// Default: send as a reply message
	d.logger.Printf("Sending reply to %s via Telegram: %s", originalID, text)
	d.sendReply(original, text)
}

func (d *Daemon) approveProposal(original *mail.Message) {
	router := mail.NewRouter(d.townRoot)
	mailbox := mail.NewMailboxFromAddress("overseer", d.townRoot)

	reply := mail.NewReplyMessage(
		"overseer",
		original.From,
		"Re: "+original.Subject,
		"[APPROVED] ✓ (via Telegram)",
		original,
	)

	if err := router.Send(reply); err != nil {
		d.logger.Printf("Error sending approval reply: %v", err)
		return
	}

	if err := mailbox.MarkRead(original.ID); err != nil {
		d.logger.Printf("Error marking message %s as read: %v", original.ID, err)
	}
}

func (d *Daemon) rejectProposal(original *mail.Message) {
	router := mail.NewRouter(d.townRoot)
	mailbox := mail.NewMailboxFromAddress("overseer", d.townRoot)

	reply := mail.NewReplyMessage(
		"overseer",
		original.From,
		"Re: "+original.Subject,
		"[REJECTED] ✗ (via Telegram)",
		original,
	)

	if err := router.Send(reply); err != nil {
		d.logger.Printf("Error sending rejection reply: %v", err)
		return
	}

	if err := mailbox.MarkRead(original.ID); err != nil {
		d.logger.Printf("Error marking message %s as read: %v", original.ID, err)
	}
}

func (d *Daemon) sendReply(original *mail.Message, text string) {
	router := mail.NewRouter(d.townRoot)
	mailbox := mail.NewMailboxFromAddress("overseer", d.townRoot)

	reply := mail.NewReplyMessage(
		"overseer",
		original.From,
		"Re: "+original.Subject,
		text,
		original,
	)

	if err := router.Send(reply); err != nil {
		d.logger.Printf("Error sending reply: %v", err)
		return
	}

	// For non-proposals, we might not want to close the original message immediately?
	// But in inbox TUI, replying typically marks it as read if it was a QUESTION.
	// Let's mark as read to keep inbox clean.
	if err := mailbox.MarkRead(original.ID); err != nil {
		d.logger.Printf("Error marking message %s as read: %v", original.ID, err)
	}
}

func (d *Daemon) handleCommand(tgMsg *tgbotapi.Message) {
	switch tgMsg.Command() {
	case "status":
		d.bot.Send(tgbotapi.NewMessage(d.config.ChatID, "Gas Town Telegram Bridge is running."))
	case "help":
		d.bot.Send(tgbotapi.NewMessage(d.config.ChatID, "/status - Check bridge status\nReply to any message to send a reply back to Gas Town."))
	}
}

// Stop stops the daemon.
func (d *Daemon) Stop() {
	d.cancel()
}

// IsRunning checks if a telegram daemon is running.
func IsRunning(townRoot string) (bool, int, error) {
	pidFile := filepath.Join(townRoot, "daemon", "telegram.pid")
	data, err := os.ReadFile(pidFile)
	if err != nil {
		if os.IsNotExist(err) {
			return false, 0, nil
		}
		return false, 0, err
	}

	var pid int
	_, err = fmt.Sscanf(string(data), "%d", &pid)
	if err != nil {
		return false, 0, nil
	}

	// Check if process is running
	process, err := os.FindProcess(pid)
	if err != nil {
		return false, 0, nil
	}

	// On Unix, FindProcess always succeeds. Send signal 0 to check if alive.
	err = process.Signal(syscall.Signal(0))
	if err != nil {
		return false, 0, nil
	}

	return true, pid, nil
}
