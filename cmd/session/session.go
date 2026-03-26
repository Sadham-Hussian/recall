package session

import (
	"fmt"
	"log"
	"os"
	"recall/internal/config"
	"recall/internal/format"
	"recall/internal/services/session"

	"github.com/spf13/cobra"
)

var lastSessions int

var sessionCmd = &cobra.Command{
	Use:   "session",
	Short: "Show command sessions",
	Run: func(cmd *cobra.Command, args []string) {

		config.LoadConfig()

		sessionService, err := session.NewSessionService()
		if err != nil {
			log.Fatalf("failed to create session service: %v", err)
		}

		// Default behavior → show current session
		if lastSessions == 0 {

			shellPID := os.Getppid()

			sessionID, commands, err := sessionService.GetCurrentSessionByShellPID(shellPID)
			if err != nil {
				log.Fatalf("failed to get session: %v", err)
			}

			if sessionID == "" {
				fmt.Println("No active session found.")
				return
			}

			fmt.Println("Current Session")
			fmt.Println("───────────────")
			fmt.Println("Session ID:", sessionID)
			fmt.Println()

			for i, c := range commands {

				fmt.Printf(
					"%2d. %-50s %s\n",
					i+1,
					c.Command,
					format.RelativeTime(c.Timestamp),
				)
			}

			return
		}

		// --last N sessions
		sessions, err := sessionService.GetLastSessions(lastSessions)
		if err != nil {
			log.Fatalf("failed to fetch sessions: %v", err)
		}

		if len(sessions) == 0 {
			fmt.Println("No sessions found.")
			return
		}

		for i, s := range sessions {

			fmt.Printf("Session %d\n", i+1)
			fmt.Println("────────────")
			fmt.Println("Session ID:", s.SessionID)
			fmt.Println()

			for j, c := range s.Commands {

				fmt.Printf(
					"%2d. %-50s %s\n",
					j+1,
					c.Command,
					format.RelativeTime(c.Timestamp),
				)
			}

			fmt.Println()
		}
	},
}

func GetSessionCmd() *cobra.Command {

	sessionCmd.AddCommand(replayCmd)

	sessionCmd.Flags().
		IntVarP(&lastSessions, "last", "l", 0, "Show last N sessions")

	return sessionCmd
}
