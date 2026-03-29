package extract

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

// isRunning checks if a process matching the name is running.
func isRunning(name string) bool {
	err := exec.Command("pgrep", "-f", name).Run()
	return err == nil
}

// removeStaleLock removes a SingletonLock ONLY from rlss-managed directories.
// It NEVER touches the system Chrome directory to prevent session corruption.
func removeStaleLock(userDataDir string) bool {
	if !IsLocked(userDataDir) {
		return false
	}
	// Safety: only remove locks from rlss-managed directories
	if !strings.Contains(userDataDir, "rlss") {
		slog.Debug("skip stale lock removal for non-rlss dir", "dir", userDataDir)
		return false
	}
	if isRunning("Google Chrome") {
		return false
	}
	lockPath := filepath.Join(userDataDir, "SingletonLock")
	if err := os.Remove(lockPath); err != nil {
		slog.Debug("could not remove stale lock", "path", lockPath, "err", err)
		return false
	}
	slog.Info("removed stale SingletonLock (rlss-managed dir)", "path", lockPath)
	return true
}

// SmartResolve implements the full resolution logic:
//  1. email + profile -> match both, check lock
//  2. email only -> find unlocked matching profile (prefer rlss dir over system Chrome)
//  3. profile only -> current Resolve logic
//  4. neither -> default
//
// When cloneProfile is true and the resolved profile is locked, it clones the
// essential profile files to a temp directory and returns that instead.
func (r *ProfileResolver) SmartResolve(email, profile, defaultUserDataDir string, cloneProfile bool) (folderName string, userDataDir string, err error) {
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
		return r.ensureAccessible(pi, cloneProfile)
	}

	// Case 2: email only
	matches := r.ResolveByEmail(email)
	if len(matches) == 0 {
		return "", "", fmt.Errorf("no profile found for email %q", email)
	}

	// Prefer unlocked profiles; prefer rlss-dedicated dir over system Chrome dir.
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
			return r.ensureAccessible(&c, cloneProfile)
		}
	}
	// Fallback to first candidate
	return r.ensureAccessible(&candidates[0], cloneProfile)
}

// ensureAccessible checks if the profile's userDataDir is locked.
// If locked and cloneProfile is true, it clones essential profile files
// to a deterministic temp directory.
func (r *ProfileResolver) ensureAccessible(pi *ProfileInfo, cloneProfile bool) (string, string, error) {
	if !IsLocked(pi.UserDataDir) {
		return pi.FolderName, pi.UserDataDir, nil
	}

	// Check for stale lock (Chrome crashed but lock remains)
	if removeStaleLock(pi.UserDataDir) {
		return pi.FolderName, pi.UserDataDir, nil
	}

	if !cloneProfile {
		fmt.Fprint(os.Stderr, FormatLockedBanner(pi.UserDataDir, pi.DisplayName, pi.Email))
		return "", "", fmt.Errorf("user data dir %q is locked; set clone_profile: true to clone profile data automatically", pi.UserDataDir)
	}

	// Clone profile to deterministic temp path
	cloneDir, err := CloneProfile(pi.UserDataDir, pi.FolderName)
	if err != nil {
		return "", "", fmt.Errorf("clone profile: %w", err)
	}

	slog.Info("cloned Chrome profile to bypass SingletonLock",
		"source", pi.UserDataDir,
		"profile", pi.FolderName,
		"clone", cloneDir,
	)

	return pi.FolderName, cloneDir, nil
}

// cloneTempDir returns a deterministic temp directory path for a cloned profile.
// Format: /tmp/rlss-chrome-{profileFolder}/ (e.g., /tmp/rlss-chrome-Default/)
func cloneTempDir(profileFolder string) string {
	safe := strings.ReplaceAll(profileFolder, " ", "-")
	safe = strings.ToLower(safe)
	return filepath.Join(os.TempDir(), fmt.Sprintf("rlss-chrome-%s", safe))
}

// essentialFiles are the profile files/dirs needed for Reading List + login state.
// Large cache directories are excluded.
var essentialFiles = []string{
	"Preferences",
	"Secure Preferences",
	"Cookies",
	"Cookies-journal",
	"Login Data",
	"Login Data-journal",
	"Web Data",
	"Web Data-journal",
	"Sync Data",
	"Sync Data Backup",
}

// CloneProfile copies essential profile files to a deterministic temp directory.
// The temp dir is reused across runs (cleaned and re-created each time).
func CloneProfile(srcUserDataDir, profileFolder string) (string, error) {
	cloneDir := cloneTempDir(profileFolder)

	// Clean previous clone
	os.RemoveAll(cloneDir)

	// Create clone directory structure
	cloneProfileDir := filepath.Join(cloneDir, profileFolder)
	if err := os.MkdirAll(cloneProfileDir, 0755); err != nil {
		return "", fmt.Errorf("create clone dir: %w", err)
	}

	// Copy Local State
	srcLocalState := filepath.Join(srcUserDataDir, "Local State")
	dstLocalState := filepath.Join(cloneDir, "Local State")
	if err := copyFile(srcLocalState, dstLocalState); err != nil {
		return "", fmt.Errorf("copy Local State: %w", err)
	}

	// Copy essential profile files
	srcProfileDir := filepath.Join(srcUserDataDir, profileFolder)
	copied := 0
	for _, name := range essentialFiles {
		src := filepath.Join(srcProfileDir, name)
		dst := filepath.Join(cloneProfileDir, name)

		info, err := os.Stat(src)
		if err != nil {
			continue // file doesn't exist, skip
		}

		if info.IsDir() {
			if err := copyDir(src, dst); err != nil {
				slog.Debug("skip dir copy", "name", name, "err", err)
				continue
			}
		} else {
			if err := copyFile(src, dst); err != nil {
				slog.Debug("skip file copy", "name", name, "err", err)
				continue
			}
		}
		copied++
	}

	slog.Info("profile clone complete",
		"files_copied", copied,
		"clone_dir", cloneDir,
	)

	return cloneDir, nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)

		if info.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		return copyFile(path, target)
	})
}

// FormatLockedBanner returns a prominent banner when a profile is locked.
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
		"║  Set clone_profile: true to auto-clone profile data.          ║\n" +
		"║                                                                ║\n" +
		"╠══════════════════════════════════════════════════════════════════╣\n" +
		"║  Chrome が SingletonLock で使用中です。                         ║\n" +
		"║  clone_profile: true を設定するとプロファイルを複製して使用します║\n" +
		"╠══════════════════════════════════════════════════════════════════╣\n" +
		"║  Chrome 正在使用此 Profile（SingletonLock）。                   ║\n" +
		"║  設定 clone_profile: true 可自動複製 Profile 資料使用。         ║\n" +
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
