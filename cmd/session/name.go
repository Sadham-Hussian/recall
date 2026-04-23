package session

import (
	"fmt"
	"strings"

	"recall/internal/config"
	"recall/internal/services/session"

	"github.com/spf13/cobra"
)

var nameCmd = &cobra.Command{
	Use:               "name <session_id> [label]",
	Short:             "Name a session or show its name",
	Args:              cobra.RangeArgs(1, 2),
	ValidArgsFunction: sessionCompletion,
	RunE: func(cmd *cobra.Command, args []string) error {
		config.LoadConfig()

		svc, err := session.NewSessionService()
		if err != nil {
			return err
		}

		sessionID := args[0]

		if len(args) == 1 {
			// Show name
			name, err := svc.GetSessionName(sessionID)
			if err != nil {
				return err
			}
			if name == "" {
				fmt.Println("Session has no name.")
			} else {
				fmt.Println(name)
			}
			return nil
		}

		// Set name
		label := strings.TrimSpace(args[1])
		if label == "" {
			return fmt.Errorf("name cannot be empty")
		}
		if err := svc.SetSessionName(sessionID, label); err != nil {
			return err
		}
		fmt.Printf("✔ Session named: %s\n", label)
		return nil
	},
}
