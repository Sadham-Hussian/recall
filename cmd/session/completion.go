package session

import (
	"fmt"
	"strings"
	"time"

	"recall/internal/config"
	"recall/internal/services/session"

	"github.com/spf13/cobra"
)

// sessionCompletion returns both session IDs and session names as
// completion candidates. Since `replay` accepts either a session ID
// or a name via ResolveSession, both are valid inputs.
func sessionCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	config.LoadConfig()

	svc, err := session.NewSessionService()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	limit := 100
	if config.AppConfig != nil && config.AppConfig.Session.AutocompleteLimit > 0 {
		limit = config.AppConfig.Session.AutocompleteLimit
	}
	sessions, err := svc.GetLastSessions(limit)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	out := make([]string, 0, len(sessions)*2)
	for _, s := range sessions {
		// Build description from first command + timestamp
		var firstCmdDesc string
		if len(s.Commands) > 0 {
			first := s.Commands[0]
			c := first.Command
			if len(c) > 40 {
				c = c[:37] + "..."
			}
			t := time.Unix(first.Timestamp, 0)
			firstCmdDesc = fmt.Sprintf("%s (%s)", c, t.Format("Jan 2 15:04"))
		}

		sid := strings.ReplaceAll(s.SessionID, ":", "\\:")

		// Always suggest the session ID
		if toComplete == "" || strings.HasPrefix(s.SessionID, toComplete) {
			desc := s.Name
			if desc == "" {
				desc = firstCmdDesc
			}
			desc = strings.ReplaceAll(desc, ":", "\\:")
			if desc != "" {
				out = append(out, sid+"\t"+desc)
			} else {
				out = append(out, sid)
			}
		}

		// Also suggest the name (if it has one) as a separate completion candidate
		if s.Name != "" && (toComplete == "" || strings.HasPrefix(s.Name, toComplete)) {
			name := strings.ReplaceAll(s.Name, ":", "\\:")
			descForName := strings.ReplaceAll(firstCmdDesc, ":", "\\:")
			if descForName != "" {
				out = append(out, name+"\t"+descForName)
			} else {
				out = append(out, name)
			}
		}
	}

	return out, cobra.ShellCompDirectiveNoFileComp
}
