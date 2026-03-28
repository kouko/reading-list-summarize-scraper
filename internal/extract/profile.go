package extract

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ProfileResolver struct {
	uiNameToFolder map[string]string
	folderNames    map[string]bool
}

func NewProfileResolver(chromeUserDataDir string) (*ProfileResolver, error) {
	if chromeUserDataDir == "" {
		home, _ := os.UserHomeDir()
		chromeUserDataDir = filepath.Join(home, "Library", "Application Support", "Google", "Chrome")
	}

	localStatePath := filepath.Join(chromeUserDataDir, "Local State")
	data, err := os.ReadFile(localStatePath)
	if err != nil {
		return nil, fmt.Errorf("read Local State: %w", err)
	}

	var state struct {
		Profile struct {
			InfoCache map[string]struct {
				Name string `json:"name"`
			} `json:"info_cache"`
		} `json:"profile"`
	}
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("parse Local State: %w", err)
	}

	r := &ProfileResolver{
		uiNameToFolder: make(map[string]string),
		folderNames:    make(map[string]bool),
	}
	for folder, info := range state.Profile.InfoCache {
		r.uiNameToFolder[info.Name] = folder
		r.folderNames[folder] = true
	}
	return r, nil
}

func (r *ProfileResolver) Resolve(name string) (string, error) {
	if r.folderNames[name] {
		return name, nil
	}
	if folder, ok := r.uiNameToFolder[name]; ok {
		return folder, nil
	}
	return "", fmt.Errorf("Chrome profile %q not found. Available: %s", name, r.availableNames())
}

func (r *ProfileResolver) availableNames() string {
	var names []string
	for uiName, folder := range r.uiNameToFolder {
		names = append(names, fmt.Sprintf("%q (%s)", uiName, folder))
	}
	return strings.Join(names, ", ")
}
