package record

import (
	"fmt"
	"recall/internal/config"
	"recall/internal/services/command_execution"
	"time"

	"github.com/spf13/cobra"
)

var (
	cmdStr    string
	exitCode  int
	cwd       string
	ts        int64
	shellPID  int
	sessionID string
)

var recordCmd = &cobra.Command{
	Use:    "record",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {

		config.LoadConfig()

		if cmdStr == "" {
			return
		}

		timestamp := time.Unix(ts, 0)

		commandExecutionService, err := command_execution.NewCommandExecutionService()
		if err != nil {
			fmt.Println("Error creating command execution service:", err)
			return
		}

		_, err = commandExecutionService.RecordLiveCommandExecution(cmdStr, ts, cwd, exitCode, shellPID, sessionID)
		if err != nil {
			fmt.Println("Error recording command execution:", err)
			return
		}

		fmt.Println("----- RECALL RECORD -----")
		fmt.Println("Command:", cmdStr)
		fmt.Println("Exit Code:", exitCode)
		fmt.Println("CWD:", cwd)
		fmt.Println("Timestamp:", timestamp)
		fmt.Println("--------------------------")
	},
}

func GetRecordCmd() *cobra.Command {
	recordCmd.Flags().StringVar(&cmdStr, "cmd", "", "Executed command")
	recordCmd.Flags().IntVar(&exitCode, "exit", 0, "Exit code")
	recordCmd.Flags().StringVar(&cwd, "cwd", "", "Working directory")
	recordCmd.Flags().Int64Var(&ts, "ts", time.Now().Unix(), "Timestamp")
	recordCmd.Flags().IntVar(&shellPID, "shell-pid", 0, "Shell PID")
	recordCmd.Flags().StringVar(&sessionID, "session-id", "", "Session ID")
	return recordCmd
}
