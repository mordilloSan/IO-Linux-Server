package update

type UpdateGroup struct {
	Name     string   `json:"name"`
	Version  string   `json:"version"`
	Severity string   `json:"severity"`
	Packages []string `json:"packages"`
}

type UpgradeItem struct {
	Package string `json:"package"`
	Version string `json:"version,omitempty"`
}

type UpdateHistoryEntry struct {
	Date     string        `json:"date"`
	Upgrades []UpgradeItem `json:"upgrades"`
}
