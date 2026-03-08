package hookscmd

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"
)

// NotificationCmd represents the notification hook command
var NotificationCmd = &cobra.Command{
	Use:   "notification",
	Short: "Handle Claude Code notifications with desktop alerts",
	Long: `Notification hook that receives Claude Code notification events.
Sends desktop notifications on macOS (osascript) and Linux (notify-send).`,
	RunE: runNotification,
}

type notificationInput struct {
	SessionID        string `json:"session_id"`
	Title            string `json:"title"`
	Message          string `json:"message"`
	NotificationType string `json:"notification_type"`
}

func runNotification(cmd *cobra.Command, args []string) error {
	var input notificationInput
	decoder := json.NewDecoder(os.Stdin)
	_ = decoder.Decode(&input)

	title := input.Title
	if title == "" {
		title = "JikiME-ADK"
	}
	message := input.Message

	// sendDesktopNotification is defined in session_end_cleanup.go (shared)
	if message != "" {
		sendDesktopNotification(title, message)
	}

	response := HookResponse{Continue: true}
	return writeResponse(response)
}
