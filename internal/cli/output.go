package cli

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/user/subscriptions-monitor/internal/provider"
)

func PrintJSON(data interface{}) error {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal data to JSON: %w", err)
	}
	fmt.Println(string(b))
	return nil
}

func PrintTable(snapshots []provider.UsageSnapshot) error {
	cellStyle := lipgloss.NewStyle().Padding(0, 1)

	t := table.New().
		Border(lipgloss.ASCIIBorder()).
		BorderRow(true).
		StyleFunc(func(row, col int) lipgloss.Style {
			return cellStyle
		}).
		Headers("NAME", "PLAN", "USAGE")

	for _, s := range snapshots {
		t.Row(
			s.Name,
			formatPlan(s.Plan),
			formatUsage(s.Metrics),
		)
	}

	header := "AI Subscriptions Usage"
	footer := fmt.Sprintf("Updated: %s", formatRefreshTime(snapshots))

	fmt.Println(header)
	fmt.Println(t)
	fmt.Println(footer)

	return nil
}

func formatRefreshTime(snapshots []provider.UsageSnapshot) string {
	if len(snapshots) == 0 {
		return "never"
	}

	var oldest time.Time
	for _, s := range snapshots {
		if oldest.IsZero() || s.Timestamp.Before(oldest) {
			oldest = s.Timestamp
		}
	}

	age := time.Since(oldest)
	relative := formatAge(age)

	return fmt.Sprintf("%s (%s)", oldest.Format("2006-01-02 15:04:05"), relative)
}

func formatAge(age time.Duration) string {
	if age < time.Second {
		return "just now"
	}
	if age < time.Minute {
		return fmt.Sprintf("%ds ago", int(age.Seconds()))
	}
	if age < time.Hour {
		return fmt.Sprintf("%dm ago", int(age.Minutes()))
	}
	if age < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(age.Hours()))
	}
	days := int(age.Hours()) / 24
	return fmt.Sprintf("%dd ago", days)
}

func formatPlan(p *provider.PlanInfo) string {
	if p == nil {
		return "N/A"
	}
	return p.Name
}

func formatUsage(metrics []provider.UsageMetric) string {
	if len(metrics) == 0 {
		return "N/A"
	}
	var parts []string
	for _, m := range metrics {
		parts = append(parts, formatMetric(m))
	}
	return strings.Join(parts, "\n")
}

func formatMetric(m provider.UsageMetric) string {
	if m.Amount.Used == nil && m.Amount.Limit == nil {
		return fmt.Sprintf("%s: N/A", m.Name)
	}

	var resetInfo string
	if m.Window.ResetsAt != nil {
		remaining := time.Until(*m.Window.ResetsAt)
		if remaining > 0 {
			resetInfo = fmt.Sprintf("  resets in %s", formatDuration(remaining))
		} else {
			resetInfo = "  resets soon"
		}
	}

	var usageLine string
	if m.Amount.Used != nil && m.Amount.Limit != nil {
		percent := (*m.Amount.Used / *m.Amount.Limit) * 100
		bar := progressBar(percent)
		usageLine = fmt.Sprintf("%s %s %.0f%%", m.Name, bar, percent)
	} else if m.Amount.Used != nil {
		usageLine = fmt.Sprintf("%s: %s %s", m.Name, formatNumber(*m.Amount.Used), m.Amount.Unit)
	} else {
		usageLine = fmt.Sprintf("%s: -/%s %s", m.Name, formatNumber(*m.Amount.Limit), m.Amount.Unit)
	}

	if resetInfo != "" {
		return usageLine + "\n" + resetInfo
	}
	return usageLine
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	if hours == 0 {
		return fmt.Sprintf("%dd", days)
	}
	return fmt.Sprintf("%dd%dh", days, hours)
}

func progressBar(percent float64) string {
	width := 10

	if percent < 0 || percent != percent {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}

	filled := int((percent / 100) * float64(width))
	empty := width - filled

	return fmt.Sprintf("[%s%s]",
		strings.Repeat("#", filled),
		strings.Repeat("-", empty),
	)
}

func formatNumber(n float64) string {
	if n >= 1000000 {
		return fmt.Sprintf("%.1fM", n/1000000)
	}
	if n >= 1000 {
		return fmt.Sprintf("%.1fK", n/1000)
	}
	return strconv.FormatFloat(n, 'f', 0, 64)
}
