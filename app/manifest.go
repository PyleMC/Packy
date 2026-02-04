package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type resourcePackManifest struct {
	FormatVersion any              `json:"format_version"`
	Header        resourcePackHead `json:"header"`
	Modules       []resourceModule `json:"modules"`
}

type resourcePackHead struct {
	Name             string `json:"name"`
	Description      string `json:"description"`
	UUID             string `json:"uuid"`
	Version          []int  `json:"version"`
	MinEngineVersion []int  `json:"min_engine_version"`
}

type resourceModule struct {
	Type    string `json:"type"`
	UUID    string `json:"uuid"`
	Version []int  `json:"version"`
}

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

func validateResourcePackFolder(folderPath string) error {
	manifestPath := filepath.Join(folderPath, "manifest.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("manifest.json not found in folder")
	}

	var manifest resourcePackManifest
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	if err := dec.Decode(&manifest); err != nil {
		return fmt.Errorf("manifest.json is not valid JSON: %w", err)
	}

	return validateManifest(manifest)
}

func validateManifest(manifest resourcePackManifest) error {
	var issues []string

	formatVersion, ok := parseFormatVersion(manifest.FormatVersion)
	if !ok || (formatVersion != 1 && formatVersion != 2) {
		issues = append(issues, "format_version must be 1 or 2")
	}

	if strings.TrimSpace(manifest.Header.Name) == "" {
		issues = append(issues, "header.name is required")
	}
	// description may be empty
	if !isValidUUID(manifest.Header.UUID) {
		issues = append(issues, "header.uuid must be a valid UUID")
	}
	if !isValidVersion(manifest.Header.Version) {
		issues = append(issues, "header.version must be 3 integers")
	}
	if len(manifest.Header.MinEngineVersion) > 0 && !isValidVersion(manifest.Header.MinEngineVersion) {
		issues = append(issues, "header.min_engine_version must be 3 integers")
	}

	if len(manifest.Modules) == 0 {
		issues = append(issues, "modules must include at least one entry")
	} else {
		for i, module := range manifest.Modules {
			if strings.TrimSpace(module.Type) == "" {
				issues = append(issues, fmt.Sprintf("modules[%d].type is required", i))
			}
			if !isValidUUID(module.UUID) {
				issues = append(issues, fmt.Sprintf("modules[%d].uuid must be a valid UUID", i))
			}
			if !isValidVersion(module.Version) {
				issues = append(issues, fmt.Sprintf("modules[%d].version must be 3 integers", i))
			}
		}
	}

	if len(issues) > 0 {
		return fmt.Errorf("%s", strings.Join(issues, "; "))
	}
	return nil
}

func parseFormatVersion(value any) (int, bool) {
	switch v := value.(type) {
	case json.Number:
		i, err := v.Int64()
		return int(i), err == nil
	case float64:
		return int(v), true
	case int:
		return v, true
	case string:
		if strings.TrimSpace(v) == "" {
			return 0, false
		}
		var num json.Number = json.Number(v)
		i, err := num.Int64()
		return int(i), err == nil
	default:
		return 0, false
	}
}

func isValidUUID(value string) bool {
	return uuidPattern.MatchString(strings.TrimSpace(value))
}

func isValidVersion(version []int) bool {
	if len(version) != 3 {
		return false
	}
	for _, v := range version {
		if v < 0 {
			return false
		}
	}
	return true
}
