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
	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderRow(true).
		Headers("PROVIDER", "NAME", "PLAN", "USAGE", "COST")

	for _, s := range snapshots {
		t.Row(
			formatProvider(s),
			s.Name,
			formatPlan(s.Plan),
			formatUsage(s.Metrics),
			formatCost(s.Cost),
		)
	}

	header := lipgloss.NewStyle().Bold(true).Render("AI Subscriptions Usage")
	footer := lipgloss.NewStyle().Faint(true).Render(fmt.Sprintf("Updated: %s", time.Now().Format(time.RFC1123)))

	fmt.Println(header)
	fmt.Println(t)
	fmt.Println(footer)

	return nil
}

func formatProvider(s provider.UsageSnapshot) string {
	var statusIcon string
	switch s.Status {
	case provider.StatusOK:
		statusIcon = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render("✓")
	case provider.StatusError:
		statusIcon = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render("✗")
	case provider.StatusUnauthorized:
		statusIcon = lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Render("⚠")
	}
	return fmt.Sprintf("%s %s", statusIcon, s.DisplayName)
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
	// Case 1: No data at all
	if m.Amount.Used == nil && m.Amount.Limit == nil {
		return fmt.Sprintf("%s: N/A", m.Name)
	}

	// Case 2: Has both Used and Limit - show progress bar
	if m.Amount.Used != nil && m.Amount.Limit != nil {
		percent := (*m.Amount.Used / *m.Amount.Limit) * 100
		bar := progressBar(percent)
		return fmt.Sprintf("%s %s %.0f%%", m.Name, bar, percent)
	}

	// Case 3: Only Used, no Limit - show value only (e.g., Total Tokens, API Requests)
	if m.Amount.Used != nil {
		return fmt.Sprintf("%s: %s %s", m.Name, formatNumber(*m.Amount.Used), m.Amount.Unit)
	}

	// Case 4: Only Limit, no Used
	return fmt.Sprintf("%s: -/%s %s", m.Name, formatNumber(*m.Amount.Limit), m.Amount.Unit)
}

func formatCost(c *provider.CostBreakdown) string {
	if c == nil {
		return "N/A"
	}
	return fmt.Sprintf("$%.2f", c.Total)
}

func progressBar(percent float64) string {
	width := 10

	// Handle NaN and negative values
	if percent < 0 || percent != percent { // NaN check
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}

	filled := int((percent / 100) * float64(width))
	empty := width - filled

	return fmt.Sprintf("%s%s",
		lipgloss.NewStyle().Render(strings.Repeat("█", filled)),
		lipgloss.NewStyle().Faint(true).Render(strings.Repeat("░", empty)),
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
