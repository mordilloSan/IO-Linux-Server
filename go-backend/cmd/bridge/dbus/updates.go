package dbus

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/godbus/dbus/v5"
)

type UpdateDetail struct {
	PackageID    string
	Updates      []string
	Obsoletes    []string
	VendorURLs   []string
	BugzillaURLs []string
	CVEURLs      []string
	Restart      uint32
	UpdateText   string
	Changelog    string
	State        uint32
	Issued       string
	Updated      string
	Summary      string
}

// --- Helpers ---

func extractCVEs(text string) []string {
	re := regexp.MustCompile(`CVE-\d{4}-\d+`)
	return re.FindAllString(text, -1)
}

func formatTextForHTML(text string) string {
	return strings.ReplaceAll(text, "\n", "<br>")
}

func extractIssued(changelog string) string {
	re := regexp.MustCompile(`(\w{3},\s*\d{1,2}\s*\w+\s*\d{4}\s*\d{2}:\d{2}:\d{2}\s*[-+]\d{4})`)
	match := re.FindStringSubmatch(changelog)
	if len(match) > 1 {
		t, err := time.Parse("Mon, 2 Jan 2006 15:04:05 -0700", match[1])
		if err == nil {
			return t.Format(time.RFC3339)
		}
		return match[1]
	}
	return ""
}

func extractNameVersion(packageID string) (name, version string) {
	parts := strings.Split(packageID, ";")
	if len(parts) >= 2 {
		return parts[0], parts[1]
	}
	return packageID, ""
}

// --- Clean Output ---

func cleanUpdateDetail(detail map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{})
	// Remove all empty fields and convert nil slices to omission
	for k, v := range detail {
		switch vv := v.(type) {
		case string:
			if vv != "" {
				out[k] = vv
			}
		case []string:
			if len(vv) > 0 {
				out[k] = vv
			}
		case []interface{}:
			if len(vv) > 0 {
				out[k] = vv
			}
		default:
			if v != nil {
				out[k] = v
			}
		}
	}
	// Add name/version if not present
	if pid, ok := out["package_id"].(string); ok {
		name, version := extractNameVersion(pid)
		out["name"] = name
		out["version"] = version
	}
	// Optionally extract issued date from changelog
	if changelog, ok := out["changelog"].(string); ok {
		issued := extractIssued(changelog)
		if issued != "" {
			out["issued"] = issued
		}
	}
	return out
}

// --- D-Bus Core ---

func GetUpdatesWithDetails() (string, error) {
	details, err := getUpdatesWithDetails()
	if err != nil {
		return "", err
	}
	// Clean up all updates before returning
	cleaned := make([]map[string]interface{}, 0, len(details))
	for _, d := range details {
		cleaned = append(cleaned, cleanUpdateDetail(d))
	}
	jsonBytes, err := json.MarshalIndent(cleaned, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

func getUpdatesWithDetails() ([]map[string]interface{}, error) {
	conn, err := dbus.SystemBus()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to system bus: %w", err)
	}
	defer conn.Close()

	const (
		pkBusName      = "org.freedesktop.PackageKit"
		pkObjPath      = "/org/freedesktop/PackageKit"
		transactionIfc = "org.freedesktop.PackageKit.Transaction"
	)

	// 1. First transaction: GetUpdates
	obj := conn.Object(pkBusName, dbus.ObjectPath(pkObjPath))
	var updatesTransPath dbus.ObjectPath
	if err := obj.Call("org.freedesktop.PackageKit.CreateTransaction", 0).Store(&updatesTransPath); err != nil {
		return nil, fmt.Errorf("CreateTransaction failed: %w", err)
	}
	updatesTrans := conn.Object(pkBusName, updatesTransPath)

	updatesCh := make(chan *dbus.Signal, 20)
	conn.Signal(updatesCh)
	conn.AddMatchSignal(
		dbus.WithMatchObjectPath(updatesTransPath),
	)

	getUpdatesCall := updatesTrans.Call(transactionIfc+".GetUpdates", 0, uint64(0))
	if getUpdatesCall.Err != nil {
		return nil, fmt.Errorf("GetUpdates failed: %w", getUpdatesCall.Err)
	}

	var pkgIDs []string
	var summaries []string
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

collectPackages:
	for {
		select {
		case sig := <-updatesCh:
			if sig == nil {
				break collectPackages
			}
			if sig.Name == transactionIfc+".Package" {
				if len(sig.Body) > 2 {
					pkgID, _ := sig.Body[1].(string)
					summary, _ := sig.Body[2].(string)
					pkgIDs = append(pkgIDs, pkgID)
					summaries = append(summaries, summary)
				}
			} else if sig.Name == transactionIfc+".Finished" {
				break collectPackages
			}
		case <-ctx.Done():
			break collectPackages
		}
	}

	if len(pkgIDs) == 0 {
		return nil, nil
	}

	// 2. New transaction: GetUpdateDetail
	var detailsTransPath dbus.ObjectPath
	if err := obj.Call("org.freedesktop.PackageKit.CreateTransaction", 0).Store(&detailsTransPath); err != nil {
		return nil, fmt.Errorf("CreateTransaction (for details) failed: %w", err)
	}
	detailsTrans := conn.Object(pkBusName, detailsTransPath)

	detailsCh := make(chan *dbus.Signal, 20)
	conn.Signal(detailsCh)
	conn.AddMatchSignal(
		dbus.WithMatchObjectPath(detailsTransPath),
	)

	detailCall := detailsTrans.Call(transactionIfc+".GetUpdateDetail", 0, pkgIDs)
	if detailCall.Err != nil {
		return nil, fmt.Errorf("GetUpdateDetail failed: %w", detailCall.Err)
	}

	var details []map[string]interface{}
	ctx2, cancel2 := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel2()

	summaryByPkg := map[string]string{}
	for i, id := range pkgIDs {
		if i < len(summaries) {
			summaryByPkg[id] = summaries[i]
		}
	}

collectDetails:
	for {
		select {
		case sig := <-detailsCh:
			if sig == nil {
				break collectDetails
			}
			if sig.Name == transactionIfc+".UpdateDetail" {
				detail := UpdateDetail{
					PackageID:  sig.Body[0].(string),
					CVEURLs:    toStringSlice(sig.Body[5]),
					Restart:    sig.Body[6].(uint32),
					UpdateText: sig.Body[7].(string),
					Changelog:  sig.Body[8].(string),
					State:      sig.Body[9].(uint32),
					Issued:     sig.Body[10].(string),
					Summary:    summaryByPkg[sig.Body[0].(string)],
				}

				// Combine CVEs from detail.CVEURLs and parsed from changelog/update_text
				cveSet := make(map[string]struct{})
				for _, cve := range detail.CVEURLs {
					cveSet[cve] = struct{}{}
				}
				for _, cve := range extractCVEs(detail.Changelog) {
					cveSet[cve] = struct{}{}
				}
				for _, cve := range extractCVEs(detail.UpdateText) {
					cveSet[cve] = struct{}{}
				}
				combinedCVEs := make([]string, 0, len(cveSet))
				for cve := range cveSet {
					combinedCVEs = append(combinedCVEs, cve)
				}

				// Only keep needed fields for JSON output
				detailMap := map[string]interface{}{
					"package_id": detail.PackageID,
					"summary":    detail.Summary,
					"restart":    detail.Restart,
					"state":      detail.State,
					"changelog":  formatTextForHTML(detail.Changelog),
					"cve_urls":   combinedCVEs,
				}
				// Optional: add issued if extracted
				if issued := extractIssued(detail.Changelog); issued != "" {
					detailMap["issued"] = issued
				}
				details = append(details, detailMap)
			} else if sig.Name == transactionIfc+".Finished" {
				break collectDetails
			}
		case <-ctx2.Done():
			break collectDetails
		}
	}

	return details, nil
}

func toStringSlice(iface any) []string {
	arr, ok := iface.([]interface{})
	if !ok {
		return []string{}
	}
	strs := make([]string, 0, len(arr))
	for _, v := range arr {
		if s, ok := v.(string); ok {
			strs = append(strs, s)
		}
	}
	return strs
}
