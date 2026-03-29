package extract

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ProfileInfo holds resolved information about a single Chrome profile.
type ProfileInfo struct {
	FolderName  string // "Default", "Profile 1"
	DisplayName string // UI name from Local State
	Email       string // user_name from Local State
	UserDataDir string // which userDataDir this belongs to
}

// ProfileResolver scans one or more Chrome user data directories and resolves
// profile names/emails to folder names.
type ProfileResolver struct {
	profiles       []ProfileInfo
	uiNameToFolder map[string]string
	folderNames    map[string]bool
}

// NewProfileResolver scans one or more userDataDirs for Chrome Local State files.
// Empty strings in the variadic list are skipped.
func NewProfileResolver(userDataDirs ...string) (*ProfileResolver, error) {
	r := &ProfileResolver{
		uiNameToFolder: make(map[string]string),
		folderNames:    make(map[string]bool),
	}

	if len(userDataDirs) == 0 {
		home, _ := os.UserHomeDir()
		userDataDirs = []string{filepath.Join(home, "Library", "Application Support", "Google", "Chrome")}
	}

	var lastErr error
	loaded := 0
	for _, dir := range userDataDirs {
		if dir == "" {
			continue
		}
		if err := r.loadLocalState(dir); err != nil {
			slog.Debug("skip chrome user data dir", "dir", dir, "err", err)
			lastErr = err
			continue
		}
		loaded++
	}

	if loaded == 0 {
		if lastErr != nil {
			return nil, fmt.Errorf("no Local State found: %w", lastErr)
		}
		return nil, fmt.Errorf("no chrome user data dirs provided")
	}

	return r, nil
}

func (r *ProfileResolver) loadLocalState(userDataDir string) error {
	localStatePath := filepath.Join(userDataDir, "Local State")
	data, err := os.ReadFile(localStatePath)
	if err != nil {
		return fmt.Errorf("read Local State: %w", err)
	}

	var state struct {
		Profile struct {
			InfoCache map[string]struct {
				Name     string `json:"name"`
				UserName string `json:"user_name"`
			} `json:"info_cache"`
		} `json:"profile"`
	}
	if err := json.Unmarshal(data, &state); err != nil {
		return fmt.Errorf("parse Local State: %w", err)
	}

	for folder, info := range state.Profile.InfoCache {
		pi := ProfileInfo{
			FolderName:  folder,
			DisplayName: info.Name,
			Email:       info.UserName,
			UserDataDir: userDataDir,
		}
		r.profiles = append(r.profiles, pi)
		r.uiNameToFolder[info.Name] = folder
		r.folderNames[folder] = true
	}
	return nil
}

// Resolve looks up a profile by display name or folder name (backward compatible).
func (r *ProfileResolver) Resolve(name string) (string, error) {
	if r.folderNames[name] {
		return name, nil
	}
	if folder, ok := r.uiNameToFolder[name]; ok {
		return folder, nil
	}
	return "", fmt.Errorf("Chrome profile %q not found. Available: %s", name, r.availableNames())
}

// ResolveByEmail returns all profiles matching the given email address.
func (r *ProfileResolver) ResolveByEmail(email string) []ProfileInfo {
	var matches []ProfileInfo
	email = strings.ToLower(email)
	for _, p := range r.profiles {
		if strings.ToLower(p.Email) == email {
			matches = append(matches, p)
		}
	}
	return matches
}

// ResolveByEmailAndName finds a profile matching both email and display/folder name.
func (r *ProfileResolver) ResolveByEmailAndName(email, name string) (*ProfileInfo, error) {
	email = strings.ToLower(email)
	for _, p := range r.profiles {
		if strings.ToLower(p.Email) != email {
			continue
		}
		if p.FolderName == name || p.DisplayName == name {
			return &p, nil
		}
	}
	return nil, fmt.Errorf("no profile matching email=%q and name=%q", email, name)
}

// IsLocked checks if a userDataDir has a SingletonLock (Chrome is running with it).
func IsLocked(userDataDir string) bool {
	lockPath := filepath.Join(userDataDir, "SingletonLock")
	_, err := os.Lstat(lockPath) // Lstat doesn't follow symlinks
	return err == nil
}

// ForceQuitChrome kills all Chrome processes (macOS).
func ForceQuitChrome() error {
	return exec.Command("pkill", "-f", "Google Chrome").Run()
}

// SmartResolve implements the full resolution logic:
//  1. email + profile -> match both, check lock
//  2. email only -> find unlocked matching profile (prefer rlss dir over system Chrome)
//  3. profile only -> current Resolve logic
//  4. neither -> default
func (r *ProfileResolver) SmartResolve(email, profile, defaultUserDataDir string, forceQuit bool) (folderName string, userDataDir string, err error) {
	// Case 3: profile only
	if email == "" && profile != "" {
		folder, err := r.Resolve(profile)
		if err != nil {
			return "", defaultUserDataDir, err
		}
		return folder, defaultUserDataDir, nil
	}

	// Case 4: neither
	if email == "" && profile == "" {
		return "Default", defaultUserDataDir, nil
	}

	// Cases 1 & 2: email is set
	if email != "" && profile != "" {
		// Case 1: both email and profile specified
		pi, err := r.ResolveByEmailAndName(email, profile)
		if err != nil {
			return "", "", fmt.Errorf("smart resolve: %w", err)
		}
		return r.ensureUnlocked(pi, forceQuit)
	}

	// Case 2: email only
	matches := r.ResolveByEmail(email)
	if len(matches) == 0 {
		return "", "", fmt.Errorf("no profile found for email %q", email)
	}

	// Prefer unlocked profiles; prefer rlss-dedicated dir over system Chrome dir.
	// Sort: unlocked first, then prefer paths containing "rlss".
	var unlocked []ProfileInfo
	var locked []ProfileInfo
	for _, m := range matches {
		if IsLocked(m.UserDataDir) {
			locked = append(locked, m)
		} else {
			unlocked = append(unlocked, m)
		}
	}

	candidates := append(unlocked, locked...)
	// Prefer rlss dir
	for _, c := range candidates {
		if strings.Contains(c.UserDataDir, "rlss") {
			return r.ensureUnlocked(&c, forceQuit)
		}
	}
	// Fallback to first candidate
	return r.ensureUnlocked(&candidates[0], forceQuit)
}

// ensureUnlocked checks if the profile's userDataDir is locked and optionally
// force-quits Chrome before retrying.
func (r *ProfileResolver) ensureUnlocked(pi *ProfileInfo, forceQuit bool) (string, string, error) {
	if !IsLocked(pi.UserDataDir) {
		return pi.FolderName, pi.UserDataDir, nil
	}

	slog.Warn("Chrome user data dir is locked (Chrome may be running)",
		"dir", pi.UserDataDir, "profile", pi.DisplayName)

	if !forceQuit {
		return pi.FolderName, pi.UserDataDir, fmt.Errorf("user data dir %q is locked by another Chrome instance; set force_quit_chrome=true to kill Chrome automatically", pi.UserDataDir)
	}

	slog.Warn("force-quitting Chrome to release lock")
	if err := ForceQuitChrome(); err != nil {
		slog.Debug("pkill returned non-zero (may be expected)", "err", err)
	}
	time.Sleep(2 * time.Second)

	if IsLocked(pi.UserDataDir) {
		return "", "", fmt.Errorf("user data dir %q still locked after force-quit", pi.UserDataDir)
	}

	return pi.FolderName, pi.UserDataDir, nil
}

func (r *ProfileResolver) availableNames() string {
	var names []string
	for uiName, folder := range r.uiNameToFolder {
		names = append(names, fmt.Sprintf("%q (%s)", uiName, folder))
	}
	return strings.Join(names, ", ")
}
