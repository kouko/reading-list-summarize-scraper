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

// ForceQuitChrome gracefully quits Chrome on macOS, then force-kills if needed.
func ForceQuitChrome() error {
	// Step 1: Try graceful quit via AppleScript (lets Chrome save state)
	_ = exec.Command("osascript", "-e", `tell application "Google Chrome" to quit`).Run()
	time.Sleep(3 * time.Second)

	// Step 2: If still running, force kill
	if isRunning("Google Chrome") {
		slog.Warn("Chrome did not quit gracefully, sending SIGKILL")
		_ = exec.Command("pkill", "-9", "-f", "Google Chrome").Run()
		time.Sleep(2 * time.Second)
	}

	return nil
}

// isRunning checks if a process matching the name is running.
func isRunning(name string) bool {
	err := exec.Command("pgrep", "-f", name).Run()
	return err == nil
}

// removeStaleLock removes a SingletonLock if Chrome is not actually running.
func removeStaleLock(userDataDir string) bool {
	if !IsLocked(userDataDir) {
		return false
	}
	if isRunning("Google Chrome") {
		return false // Chrome is running, lock is valid
	}
	lockPath := filepath.Join(userDataDir, "SingletonLock")
	if err := os.Remove(lockPath); err != nil {
		slog.Debug("could not remove stale lock", "path", lockPath, "err", err)
		return false
	}
	slog.Info("removed stale SingletonLock", "path", lockPath)
	return true
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

	// Check for stale lock (Chrome crashed but lock remains)
	if removeStaleLock(pi.UserDataDir) {
		return pi.FolderName, pi.UserDataDir, nil
	}

	if !forceQuit {
		fmt.Fprint(os.Stderr, FormatLockedBanner(pi.UserDataDir, pi.DisplayName, pi.Email))
		return pi.FolderName, pi.UserDataDir, fmt.Errorf("user data dir %q is locked by another Chrome instance; set force_quit_chrome: true to kill Chrome automatically", pi.UserDataDir)
	}

	fmt.Fprint(os.Stderr, FormatForceQuitBanner())
	ForceQuitChrome()

	// Retry: check lock removal up to 3 times
	for i := 0; i < 3; i++ {
		if !IsLocked(pi.UserDataDir) {
			slog.Info("Chrome lock released, proceeding")
			return pi.FolderName, pi.UserDataDir, nil
		}
		removeStaleLock(pi.UserDataDir)
		time.Sleep(2 * time.Second)
	}

	return "", "", fmt.Errorf("user data dir %q still locked after force-quit (tried 3 times)", pi.UserDataDir)
}

// FormatLockedBanner returns a prominent banner when all matching profiles are locked.
func FormatLockedBanner(userDataDir, profileName, email string) string {
	return "\n" +
		"╔══════════════════════════════════════════════════════════════════╗\n" +
		"║  ⚠  Chrome Profile Locked (SingletonLock)                      ║\n" +
		"╠══════════════════════════════════════════════════════════════════╣\n" +
		"║                                                                ║\n" +
		fmt.Sprintf("║  Profile : %-50s  ║\n", profileName) +
		fmt.Sprintf("║  Account : %-50s  ║\n", email) +
		fmt.Sprintf("║  DataDir : %-50s  ║\n", truncatePath(userDataDir, 50)) +
		"║                                                                ║\n" +
		"║  Chrome is currently running with this user data directory.    ║\n" +
		"║  rlss cannot access the profile while Chrome holds the lock.   ║\n" +
		"║                                                                ║\n" +
		"║  Solutions:                                                    ║\n" +
		"║  1. Close Chrome, then re-run rlss                            ║\n" +
		"║  2. Set force_quit_chrome: true in config.yaml                ║\n" +
		"║  3. Use a dedicated rlss Chrome profile (default setup)       ║\n" +
		"║                                                                ║\n" +
		"╠══════════════════════════════════════════════════════════════════╣\n" +
		"║  Chrome が SingletonLock で使用中のため、                       ║\n" +
		"║  プロファイルにアクセスできません。                              ║\n" +
		"║  Chrome を閉じるか、force_quit_chrome: true を設定してください。║\n" +
		"╠══════════════════════════════════════════════════════════════════╣\n" +
		"║  Chrome 正在使用此 Profile（SingletonLock），                   ║\n" +
		"║  rlss 無法存取。請關閉 Chrome 或設定 force_quit_chrome: true。  ║\n" +
		"╚══════════════════════════════════════════════════════════════════╝\n"
}

// FormatForceQuitBanner returns a prominent warning when about to force-quit Chrome.
func FormatForceQuitBanner() string {
	return "\n" +
		"╔══════════════════════════════════════════════════════════════════╗\n" +
		"║  ⚠⚠⚠  Force-Quitting Chrome  ⚠⚠⚠                             ║\n" +
		"╠══════════════════════════════════════════════════════════════════╣\n" +
		"║                                                                ║\n" +
		"║  All Chrome windows will be closed to release the lock.        ║\n" +
		"║  Unsaved work in Chrome may be lost!                           ║\n" +
		"║                                                                ║\n" +
		"║  Chrome の全ウィンドウを強制終了します。                         ║\n" +
		"║  未保存の作業が失われる可能性があります。                        ║\n" +
		"║                                                                ║\n" +
		"║  即將強制關閉所有 Chrome 視窗以釋放鎖定。                       ║\n" +
		"║  Chrome 中未儲存的工作可能會遺失！                              ║\n" +
		"║                                                                ║\n" +
		"╚══════════════════════════════════════════════════════════════════╝\n"
}

func truncatePath(path string, maxLen int) string {
	if len(path) <= maxLen {
		return path
	}
	return "..." + path[len(path)-maxLen+3:]
}

func (r *ProfileResolver) availableNames() string {
	var names []string
	for uiName, folder := range r.uiNameToFolder {
		names = append(names, fmt.Sprintf("%q (%s)", uiName, folder))
	}
	return strings.Join(names, ", ")
}
