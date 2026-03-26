package doctor

import (
	"recall/internal/config"
	"recall/internal/doctor"

	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check system health",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.LoadConfig()
		doctor.RunDoctor(cfg)
	},
}

func GetDoctorCmd() *cobra.Command {
	return doctorCmd
}
