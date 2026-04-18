package stats

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"recall/internal/config"
	"recall/internal/format"
	statssvc "recall/internal/services/stats"
	"recall/internal/storage"
	"recall/internal/storage/models"

	"github.com/spf13/cobra"
)

var (
	days      int
	outputFmt string
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show usage statistics",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.LoadConfig()
		svc, err := statssvc.NewStatsService()
		if err != nil {
			log.Fatalf("error: %v", err)
		}

		var sinceTs int64
		if days > 0 {
			sinceTs = time.Now().AddDate(0, 0, -days).Unix()
		}

		data := gatherStats(cfg, svc, sinceTs)

		switch outputFmt {
		case "json":
			printJSON(data)
		case "md", "markdown":
			printMarkdown(data)
		default:
			printTerminal(data)
		}
	},
}

// statsData holds all gathered stats so renderers don't repeat queries.
type statsData struct {
	Overview    *models.OverviewStats
	DBSize      int64
	CmdGroups   []models.CommandGroup
	TopCmds     []models.CommandCount
	Failed      []models.FailedCommand
	TopDirs     []models.DirectoryCount
	DayActivity []models.DayActivity
	Hours       []models.HourActivity
}

func gatherStats(cfg *config.Config, svc *statssvc.StatsService, sinceTs int64) statsData {
	var d statsData

	d.Overview, _ = svc.Overview(sinceTs)

	if dbPath, err := storage.DBPath(cfg); err == nil {
		if info, err := os.Stat(dbPath); err == nil {
			d.DBSize = info.Size()
		}
	}

	d.CmdGroups, _ = svc.TopCommandGroups(sinceTs, 7)
	d.TopCmds, _ = svc.TopCommands(sinceTs, 10)
	d.Failed, _ = svc.MostFailed(sinceTs, 5, 5)
	d.TopDirs, _ = svc.TopDirectories(sinceTs, 5)

	weekTs := time.Now().AddDate(0, 0, -7).Unix()
	d.DayActivity, _ = svc.ActivityByDay(weekTs)

	d.Hours, _ = svc.ActivityByHour(sinceTs, 5)

	return d
}

func successRate(ov *models.OverviewStats) float64 {
	if ov.TotalCommands == 0 {
		return 0
	}
	return float64(ov.SuccessCount) / float64(ov.TotalCommands) * 100
}

// ── Terminal output ──────────────────────────────────────────────────────────

func printTerminal(d statsData) {
	ov := d.Overview

	fmt.Println()
	fmt.Println("Recall Stats")
	fmt.Println("────────────────────────────────────")

	fmt.Println()
	fmt.Println("Overview")
	fmt.Printf("  Total commands      %s\n", formatNumber(ov.TotalCommands))
	fmt.Printf("  Unique commands     %s\n", formatNumber(ov.UniqueCommands))
	fmt.Printf("  Sessions            %s\n", formatNumber(ov.TotalSessions))
	fmt.Printf("  Success rate        %.1f%%\n", successRate(ov))

	if d.DBSize > 0 {
		fmt.Printf("  Database size       %s\n", formatBytes(d.DBSize))
	}

	if ov.FirstTimestamp > 0 {
		first := time.Unix(ov.FirstTimestamp, 0)
		daysSince := int(time.Since(first).Hours() / 24)
		fmt.Printf("  Recording since     %s (%d days)\n", first.Format("2006-01-02"), daysSince)
	}

	if len(d.CmdGroups) > 0 {
		fmt.Println()
		fmt.Println("Top Command Groups")
		for i, g := range d.CmdGroups {
			line := fmt.Sprintf("  %2d.  %-14s (%d)", i+1, g.Group, g.Count)
			if g.Subcommands != "" {
				line += fmt.Sprintf("  — %s", g.Subcommands)
			}
			fmt.Println(line)
		}
	}

	if len(d.TopCmds) > 0 {
		fmt.Println()
		fmt.Println("Top Commands")
		for i, c := range d.TopCmds {
			fmt.Printf("  %2d.  %-30s (%d)\n", i+1, format.Truncate(c.Command, 30), c.Count)
		}
	}

	if len(d.Failed) > 0 {
		fmt.Println()
		fmt.Println("Most Failed Commands")
		for i, f := range d.Failed {
			var rate float64
			if f.TotalCount > 0 {
				rate = float64(f.TotalCount-f.FailureCount) / float64(f.TotalCount) * 100
			}
			fmt.Printf("  %2d.  %-30s %d failures (%.0f%% success)\n",
				i+1, format.Truncate(f.Command, 30), f.FailureCount, rate)
		}
	}

	if len(d.TopDirs) > 0 {
		fmt.Println()
		fmt.Println("Top Directories")
		for i, dir := range d.TopDirs {
			fmt.Printf("  %2d.  %-40s (%d)\n", i+1, format.Truncate(format.ShortenPath(dir.CWD), 40), dir.Count)
		}
	}

	if len(d.DayActivity) > 0 {
		fmt.Println()
		fmt.Println("Activity (last 7 days)")
		dayNames := []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
		dayMap := make(map[int]int64)
		var maxDay int64
		for _, a := range d.DayActivity {
			dayMap[a.DayOfWeek] = a.Count
			if a.Count > maxDay {
				maxDay = a.Count
			}
		}
		for i := 1; i <= 7; i++ {
			dow := i % 7
			count := dayMap[dow]
			bar := renderBar(count, maxDay, 30)
			fmt.Printf("  %s  %s  %d\n", dayNames[dow], bar, count)
		}
	}

	if len(d.Hours) > 0 {
		fmt.Println()
		fmt.Println("Busiest Hours")
		var maxHour int64
		for _, h := range d.Hours {
			if h.Count > maxHour {
				maxHour = h.Count
			}
		}
		for _, h := range d.Hours {
			bar := renderBar(h.Count, maxHour, 30)
			fmt.Printf("  %02d:00  %s  %d\n", h.Hour, bar, h.Count)
		}
	}

	fmt.Println()
}

// ── Markdown output ──────────────────────────────────────────────────────────

func printMarkdown(d statsData) {
	ov := d.Overview

	fmt.Println("### Terminal Stats")
	fmt.Println()
	fmt.Println("| Metric | Value |")
	fmt.Println("|--------|-------|")
	fmt.Printf("| Total commands | %s |\n", formatNumber(ov.TotalCommands))
	fmt.Printf("| Unique commands | %s |\n", formatNumber(ov.UniqueCommands))
	fmt.Printf("| Sessions | %s |\n", formatNumber(ov.TotalSessions))
	fmt.Printf("| Success rate | %.1f%% |\n", successRate(ov))
	if d.DBSize > 0 {
		fmt.Printf("| Database size | %s |\n", formatBytes(d.DBSize))
	}
	if ov.FirstTimestamp > 0 {
		first := time.Unix(ov.FirstTimestamp, 0)
		daysSince := int(time.Since(first).Hours() / 24)
		fmt.Printf("| Recording since | %s (%d days) |\n", first.Format("2006-01-02"), daysSince)
	}

	if len(d.CmdGroups) > 0 {
		fmt.Println()
		parts := make([]string, len(d.CmdGroups))
		for i, g := range d.CmdGroups {
			if g.Subcommands != "" {
				parts[i] = fmt.Sprintf("`%s` (%d) — %s", g.Group, g.Count, g.Subcommands)
			} else {
				parts[i] = fmt.Sprintf("`%s` (%d)", g.Group, g.Count)
			}
		}
		fmt.Println("**Top tool groups:** " + strings.Join(parts, ", "))
	}

	if len(d.TopCmds) > 0 {
		fmt.Println()
		fmt.Println("**Top commands:** ", end(""))
		parts := make([]string, len(d.TopCmds))
		for i, c := range d.TopCmds {
			parts[i] = fmt.Sprintf("`%s` (%d)", c.Command, c.Count)
		}
		fmt.Println(strings.Join(parts, ", "))
	}

	if len(d.Failed) > 0 {
		fmt.Println()
		fmt.Println("**Most failed:** ", end(""))
		parts := make([]string, len(d.Failed))
		for i, f := range d.Failed {
			var rate float64
			if f.TotalCount > 0 {
				rate = float64(f.TotalCount-f.FailureCount) / float64(f.TotalCount) * 100
			}
			parts[i] = fmt.Sprintf("`%s` — %d failures (%.0f%% success)", f.Command, f.FailureCount, rate)
		}
		fmt.Println(strings.Join(parts, ", "))
	}

	if len(d.TopDirs) > 0 {
		fmt.Println()
		fmt.Println("**Top directories:** ", end(""))
		parts := make([]string, len(d.TopDirs))
		for i, dir := range d.TopDirs {
			parts[i] = fmt.Sprintf("`%s` (%d)", format.ShortenPath(dir.CWD), dir.Count)
		}
		fmt.Println(strings.Join(parts, ", "))
	}

	fmt.Println()
	fmt.Println("*Powered by [Recall](https://github.com/Sadham-Hussian/recall)*")
}

// end is a no-op helper to suppress the trailing newline from Println
// when we want inline content on the next line.
func end(s string) string { return s }

// ── JSON output ──────────────────────────────────────────────────────────────

func printJSON(d statsData) {

	type jsonCommandGroup struct {
		Group       string `json:"group"`
		Count       int64  `json:"count"`
		Subcommands string `json:"subcommands,omitempty"`
	}

	type jsonOutput struct {
		Overview struct {
			TotalCommands  int64   `json:"total_commands"`
			UniqueCommands int64   `json:"unique_commands"`
			Sessions       int64   `json:"sessions"`
			SuccessRate    float64 `json:"success_rate"`
			DBSizeBytes    int64   `json:"db_size_bytes"`
			RecordingSince string  `json:"recording_since"`
		} `json:"overview"`
		CommandGroups []jsonCommandGroup `json:"command_groups"`
		TopCommands   []struct {
			Command string `json:"command"`
			Count   int64  `json:"count"`
		} `json:"top_commands"`
		MostFailed []struct {
			Command     string  `json:"command"`
			Failures    int64   `json:"failures"`
			SuccessRate float64 `json:"success_rate"`
		} `json:"most_failed"`
		TopDirectories []struct {
			Directory string `json:"directory"`
			Count     int64  `json:"count"`
		} `json:"top_directories"`
	}

	var out jsonOutput
	ov := d.Overview
	out.Overview.TotalCommands = ov.TotalCommands
	out.Overview.UniqueCommands = ov.UniqueCommands
	out.Overview.Sessions = ov.TotalSessions
	out.Overview.SuccessRate = successRate(ov)
	out.Overview.DBSizeBytes = d.DBSize
	if ov.FirstTimestamp > 0 {
		out.Overview.RecordingSince = time.Unix(ov.FirstTimestamp, 0).Format("2006-01-02")
	}

	for _, g := range d.CmdGroups {
		out.CommandGroups = append(out.CommandGroups, jsonCommandGroup{
			Group: g.Group, Count: g.Count, Subcommands: g.Subcommands,
		})
	}

	for _, c := range d.TopCmds {
		out.TopCommands = append(out.TopCommands, struct {
			Command string `json:"command"`
			Count   int64  `json:"count"`
		}{c.Command, c.Count})
	}

	for _, f := range d.Failed {
		var rate float64
		if f.TotalCount > 0 {
			rate = float64(f.TotalCount-f.FailureCount) / float64(f.TotalCount) * 100
		}
		out.MostFailed = append(out.MostFailed, struct {
			Command     string  `json:"command"`
			Failures    int64   `json:"failures"`
			SuccessRate float64 `json:"success_rate"`
		}{f.Command, f.FailureCount, rate})
	}

	for _, dir := range d.TopDirs {
		out.TopDirectories = append(out.TopDirectories, struct {
			Directory string `json:"directory"`
			Count     int64  `json:"count"`
		}{format.ShortenPath(dir.CWD), dir.Count})
	}

	b, _ := json.MarshalIndent(out, "", "  ")
	fmt.Println(string(b))
}

func renderBar(value, max int64, width int) string {
	if max == 0 {
		return ""
	}
	filled := int(float64(value) / float64(max) * float64(width))
	if filled == 0 && value > 0 {
		filled = 1
	}
	return strings.Repeat("█", filled)
}

func formatNumber(n int64) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	if n < 1_000_000 {
		return fmt.Sprintf("%d,%03d", n/1000, n%1000)
	}
	return fmt.Sprintf("%d,%03d,%03d", n/1_000_000, (n%1_000_000)/1000, n%1000)
}

func formatBytes(b int64) string {
	switch {
	case b >= 1<<30:
		return fmt.Sprintf("%.1f GB", float64(b)/float64(1<<30))
	case b >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(b)/float64(1<<20))
	case b >= 1<<10:
		return fmt.Sprintf("%.1f KB", float64(b)/float64(1<<10))
	default:
		return fmt.Sprintf("%d B", b)
	}
}

func GetStatsCmd() *cobra.Command {
	statsCmd.Flags().IntVar(&days, "days", 0, "limit stats to last N days (default: all time)")
	statsCmd.Flags().StringVar(&outputFmt, "format", "", "output format: md, json (default: terminal)")
	return statsCmd
}
