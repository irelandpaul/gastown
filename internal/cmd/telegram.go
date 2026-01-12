package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/spf13/cobra"
	"github.com/steveyegge/gastown/internal/style"
	"github.com/steveyegge/gastown/internal/telegram"
	"github.com/steveyegge/gastown/internal/workspace"
)

var telegramCmd = &cobra.Command{
	Use:     "telegram",
	GroupID: GroupComm,
	Short:   "Manage the Telegram bridge daemon",
	Long: `Manage the Telegram bridge daemon for remote overseer access.

The telegram daemon provides a two-way bridge between Gas Town's overseer inbox
and a Telegram chat. It allows you to approve proposals, answer questions,
and receive alerts via Telegram.`,
	RunE: requireSubcommand,
}

var telegramSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup Telegram bot configuration",
	RunE:  runTelegramSetup,
}

var telegramStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the Telegram bridge daemon",
	RunE:  runTelegramStart,
}

var telegramStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the Telegram bridge daemon",
	RunE:  runTelegramStop,
}

var telegramStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show Telegram bridge daemon status",
	RunE:  runTelegramStatus,
}

var telegramLogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "View Telegram bridge daemon logs",
	RunE:  runTelegramLogs,
}

var telegramTestCmd = &cobra.Command{
	Use:   "test",
	Short: "Send a test message to Telegram",
	RunE:  runTelegramTest,
}

var telegramRunCmd = &cobra.Command{
	Use:    "run",
	Short:  "Run Telegram daemon in foreground (internal)",
	Hidden: true,
	RunE:   runTelegramRun,
}

func init() {
	telegramCmd.AddCommand(telegramSetupCmd)
	telegramCmd.AddCommand(telegramStartCmd)
	telegramCmd.AddCommand(telegramStopCmd)
	telegramCmd.AddCommand(telegramStatusCmd)
	telegramCmd.AddCommand(telegramLogsCmd)
	telegramCmd.AddCommand(telegramTestCmd)
	telegramCmd.AddCommand(telegramRunCmd)

	rootCmd.AddCommand(telegramCmd)
}

func runTelegramSetup(cmd *cobra.Command, args []string) error {
	townRoot, err := workspace.FindFromCwdOrError()
	if err != nil {
		return err
	}

	fmt.Print("Enter Telegram Bot Token: ")
	var token string
	fmt.Scanln(&token)

	fmt.Print("Enter Telegram Chat ID: ")
	var chatIDStr string
	fmt.Scanln(&chatIDStr)
	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid chat ID: %w", err)
	}

	cfg := &telegram.Config{
		Token:  token,
		ChatID: chatID,
	}

	if err := telegram.SaveConfig(townRoot, cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Printf("%s Telegram configuration saved to mayor/telegram.json\n", style.Bold.Render("✓"))
	return nil
}

func runTelegramStart(cmd *cobra.Command, args []string) error {
	townRoot, err := workspace.FindFromCwdOrError()
	if err != nil {
		return err
	}

	// Check if already running
	running, pid, err := telegram.IsRunning(townRoot)
	if err != nil {
		return fmt.Errorf("checking daemon status: %w", err)
	}
	if running {
		return fmt.Errorf("telegram daemon already running (PID %d)", pid)
	}

	// Verify config exists
	if _, err := telegram.LoadConfig(townRoot); err != nil {
		return fmt.Errorf("telegram not configured. Run 'gt telegram setup' first: %w", err)
	}

	// Start daemon in background
	gtPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("finding executable: %w", err)
	}

	daemonCmd := exec.Command(gtPath, "telegram", "run")
	daemonCmd.Dir = townRoot
	daemonCmd.Stdin = nil
	daemonCmd.Stdout = nil
	daemonCmd.Stderr = nil

	if err := daemonCmd.Start(); err != nil {
		return fmt.Errorf("starting telegram daemon: %w", err)
	}

	// Wait a moment for the daemon to initialize and acquire the lock
	time.Sleep(200 * time.Millisecond)

	// Verify it started
	running, pid, err = telegram.IsRunning(townRoot)
	if err != nil {
		return fmt.Errorf("checking daemon status: %w", err)
	}
	if !running {
		return fmt.Errorf("telegram daemon failed to start (check logs with 'gt telegram logs')")
	}

	fmt.Printf("%s Telegram bridge daemon started (PID %d)\n", style.Bold.Render("✓"), pid)
	return nil
}

func runTelegramStop(cmd *cobra.Command, args []string) error {
	townRoot, err := workspace.FindFromCwdOrError()
	if err != nil {
		return err
	}

	running, pid, err := telegram.IsRunning(townRoot)
	if err != nil {
		return err
	}
	if !running {
		return fmt.Errorf("telegram daemon is not running")
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	if err := process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("stopping daemon: %w", err)
	}

	// Cleanup PID file
	pidFile := filepath.Join(townRoot, "daemon", "telegram.pid")
	_ = os.Remove(pidFile)

	fmt.Printf("%s Telegram bridge daemon stopped (was PID %d)\n", style.Bold.Render("✓"), pid)
	return nil
}

func runTelegramStatus(cmd *cobra.Command, args []string) error {
	townRoot, err := workspace.FindFromCwdOrError()
	if err != nil {
		return err
	}

	running, pid, err := telegram.IsRunning(townRoot)
	if err != nil {
		return err
	}

	if running {
		fmt.Printf("%s Telegram bridge daemon is %s (PID %d)\n",
			style.Bold.Render("●"),
			style.Bold.Render("running"),
			pid)
	} else {
		fmt.Printf("%s Telegram bridge daemon is %s\n",
			style.Dim.Render("○"),
			"not running")
	}

	return nil
}

func runTelegramLogs(cmd *cobra.Command, args []string) error {
	townRoot, err := workspace.FindFromCwdOrError()
	if err != nil {
		return err
	}

	logFile := filepath.Join(townRoot, "daemon", "telegram.log")
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		return fmt.Errorf("no log file found at %s", logFile)
	}

	tailCmd := exec.Command("tail", "-n", "50", "-f", logFile)
	tailCmd.Stdout = os.Stdout
	tailCmd.Stderr = os.Stderr
	return tailCmd.Run()
}

func runTelegramTest(cmd *cobra.Command, args []string) error {
	townRoot, err := workspace.FindFromCwdOrError()
	if err != nil {
		return err
	}

	cfg, err := telegram.LoadConfig(townRoot)
	if err != nil {
		return err
	}

	bot, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		return err
	}

	text := fmt.Sprintf("Test message from Gas Town! Time: %s", time.Now().Format(time.RFC3339))
	msg := tgbotapi.NewMessage(cfg.ChatID, text)
	_, err = bot.Send(msg)
	if err != nil {
		return fmt.Errorf("sending test message: %w", err)
	}

	fmt.Printf("%s Test message sent to Telegram chat %d\n", style.Bold.Render("✓"), cfg.ChatID)
	return nil
}

func runTelegramRun(cmd *cobra.Command, args []string) error {
	townRoot, err := workspace.FindFromCwdOrError()
	if err != nil {
		return err
	}

	cfg, err := telegram.LoadConfig(townRoot)
	if err != nil {
		return err
	}

	logFile := filepath.Join(townRoot, "daemon", "telegram.log")
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	logger := log.New(f, "", log.LstdFlags)

	d, err := telegram.NewDaemon(townRoot, cfg, logger)
	if err != nil {
		return err
	}

	return d.Run()
}
