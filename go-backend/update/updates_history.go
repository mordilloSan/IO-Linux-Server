package update

import (
	"bufio"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
)

func getUpdateHistoryHandler(c *gin.Context) {
	history := parseUpdateHistory()
	if len(history) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No update history found"})
		return
	}
	c.JSON(http.StatusOK, history)
}

func parseUpdateHistory() []UpdateHistoryEntry {
	if _, err := os.Stat("/var/log/apt/history.log"); err == nil {
		return parseAptHistory("/var/log/apt/history.log")
	}
	if _, err := os.Stat("/var/log/dnf.log"); err == nil {
		return parseDnfHistory("/var/log/dnf.log")
	}
	return []UpdateHistoryEntry{}
}

func parseAptHistory(logPath string) []UpdateHistoryEntry {
	file, err := os.Open(logPath)
	if err != nil {
		return nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	dateRe := regexp.MustCompile(`^Start-Date: (\d{4}-\d{2}-\d{2})`)
	upgradeRe := regexp.MustCompile(`^(Upgrade|Remove): (.+)`)
	pkgVerRe := regexp.MustCompile(`([a-zA-Z0-9+.\-]+)(:[^ ]+)? \(([^)]+)\)`)

	historyMap := make(map[string][]UpgradeItem)
	var currentDate string

	for scanner.Scan() {
		line := scanner.Text()

		if matches := dateRe.FindStringSubmatch(line); len(matches) == 2 {
			currentDate = matches[1]
		} else if matches := upgradeRe.FindStringSubmatch(line); len(matches) == 3 && currentDate != "" {
			entries := strings.Split(matches[2], ", ")
			for _, entry := range entries {
				if m := pkgVerRe.FindStringSubmatch(entry); len(m) == 4 {
					historyMap[currentDate] = append(historyMap[currentDate], UpgradeItem{
						Package: m[1],
						Version: m[3],
					})
				}
			}
		}
	}

	return mapToSortedHistory(historyMap)
}

func parseDnfHistory(logPath string) []UpdateHistoryEntry {
	file, err := os.Open(logPath)
	if err != nil {
		return nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	upgradeRe := regexp.MustCompile(`Upgrade:\s+([^\s-]+)-([^-]+-[^\s]+)`)

	var historyMap = make(map[string][]UpgradeItem)

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) < 1 {
			continue
		}
		date := parts[0]

		if matches := upgradeRe.FindStringSubmatch(line); len(matches) > 2 {
			historyMap[date] = append(historyMap[date], UpgradeItem{
				Package: matches[1],
				Version: matches[2],
			})
		}
	}

	return mapToSortedHistory(historyMap)
}

func mapToSortedHistory(historyMap map[string][]UpgradeItem) []UpdateHistoryEntry {
	var dates []string
	for date := range historyMap {
		dates = append(dates, date)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(dates)))

	var history []UpdateHistoryEntry
	for _, date := range dates {
		history = append(history, UpdateHistoryEntry{
			Date:     date,
			Upgrades: historyMap[date],
		})
	}
	return history
}
