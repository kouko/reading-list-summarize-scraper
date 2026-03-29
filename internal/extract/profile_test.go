package extract

import (
	"os"
	"path/filepath"
	"testing"
)

const testLocalState = `{
    "profile": {
        "info_cache": {
            "Default": {"name": "Personal", "user_name": "kouko.d@gmail.com"},
            "Profile 1": {"name": "Work Account", "user_name": "work@company.com"},
            "Profile 5": {"name": "ReadingList-Auto", "user_name": "kouko.d@gmail.com"}
        }
    }
}`

func setupTestDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "Local State"), []byte(testLocalState), 0644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestProfileResolver(t *testing.T) {
	dir := setupTestDir(t)

	r, err := NewProfileResolver(dir)
	if err != nil {
		t.Fatalf("NewProfileResolver error: %v", err)
	}

	tests := []struct {
		input string
		want  string
	}{
		{"ReadingList-Auto", "Profile 5"},
		{"Personal", "Default"},
		{"Work Account", "Profile 1"},
		{"Default", "Default"},
		{"Profile 1", "Profile 1"},
	}

	for _, tt := range tests {
		got, err := r.Resolve(tt.input)
		if err != nil {
			t.Errorf("Resolve(%q) error: %v", tt.input, err)
			continue
		}
		if got != tt.want {
			t.Errorf("Resolve(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}

	_, err = r.Resolve("Nonexistent")
	if err == nil {
		t.Error("Resolve(Nonexistent) should error")
	}
}

func TestNewProfileResolver_MultipleDirs(t *testing.T) {
	dir1 := setupTestDir(t)

	dir2 := t.TempDir()
	localState2 := `{
		"profile": {
			"info_cache": {
				"Default": {"name": "Other", "user_name": "other@example.com"}
			}
		}
	}`
	os.WriteFile(filepath.Join(dir2, "Local State"), []byte(localState2), 0644)

	r, err := NewProfileResolver(dir1, dir2)
	if err != nil {
		t.Fatalf("NewProfileResolver error: %v", err)
	}

	// Should find profiles from both dirs
	if len(r.profiles) != 4 {
		t.Errorf("expected 4 profiles, got %d", len(r.profiles))
	}
}

func TestNewProfileResolver_SkipEmpty(t *testing.T) {
	dir := setupTestDir(t)

	r, err := NewProfileResolver("", dir, "")
	if err != nil {
		t.Fatalf("NewProfileResolver error: %v", err)
	}
	if len(r.profiles) != 3 {
		t.Errorf("expected 3 profiles, got %d", len(r.profiles))
	}
}

func TestResolveByEmail(t *testing.T) {
	dir := setupTestDir(t)

	r, err := NewProfileResolver(dir)
	if err != nil {
		t.Fatalf("NewProfileResolver error: %v", err)
	}

	matches := r.ResolveByEmail("kouko.d@gmail.com")
	if len(matches) != 2 {
		t.Fatalf("expected 2 matches for kouko.d@gmail.com, got %d", len(matches))
	}

	folders := map[string]bool{}
	for _, m := range matches {
		folders[m.FolderName] = true
	}
	if !folders["Default"] || !folders["Profile 5"] {
		t.Errorf("expected Default and Profile 5, got %v", folders)
	}

	matches = r.ResolveByEmail("work@company.com")
	if len(matches) != 1 || matches[0].FolderName != "Profile 1" {
		t.Errorf("expected Profile 1 for work@company.com, got %v", matches)
	}

	matches = r.ResolveByEmail("nobody@example.com")
	if len(matches) != 0 {
		t.Errorf("expected 0 matches for nobody, got %d", len(matches))
	}
}

func TestResolveByEmail_CaseInsensitive(t *testing.T) {
	dir := setupTestDir(t)
	r, _ := NewProfileResolver(dir)

	matches := r.ResolveByEmail("KOUKO.D@GMAIL.COM")
	if len(matches) != 2 {
		t.Errorf("expected case-insensitive match, got %d", len(matches))
	}
}

func TestResolveByEmailAndName(t *testing.T) {
	dir := setupTestDir(t)
	r, _ := NewProfileResolver(dir)

	// Match by display name
	pi, err := r.ResolveByEmailAndName("kouko.d@gmail.com", "ReadingList-Auto")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pi.FolderName != "Profile 5" {
		t.Errorf("expected Profile 5, got %s", pi.FolderName)
	}

	// Match by folder name
	pi, err = r.ResolveByEmailAndName("kouko.d@gmail.com", "Default")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pi.DisplayName != "Personal" {
		t.Errorf("expected Personal, got %s", pi.DisplayName)
	}

	// No match: wrong email
	_, err = r.ResolveByEmailAndName("wrong@email.com", "Personal")
	if err == nil {
		t.Error("expected error for wrong email")
	}

	// No match: right email, wrong name
	_, err = r.ResolveByEmailAndName("kouko.d@gmail.com", "Nonexistent")
	if err == nil {
		t.Error("expected error for wrong name")
	}
}

func TestIsLocked(t *testing.T) {
	dir := t.TempDir()

	// No lock -> not locked
	if IsLocked(dir) {
		t.Error("expected not locked (no SingletonLock)")
	}

	// Create a symlink as Chrome does on Linux/macOS
	lockPath := filepath.Join(dir, "SingletonLock")
	if err := os.Symlink("some-hostname-12345", lockPath); err != nil {
		t.Skipf("cannot create symlink: %v", err)
	}

	if !IsLocked(dir) {
		t.Error("expected locked (SingletonLock exists)")
	}

	// Remove and verify
	os.Remove(lockPath)
	if IsLocked(dir) {
		t.Error("expected not locked after removal")
	}
}

func TestIsLocked_RegularFile(t *testing.T) {
	dir := t.TempDir()
	lockPath := filepath.Join(dir, "SingletonLock")
	os.WriteFile(lockPath, []byte("lock"), 0644)

	// Regular file should also count as locked
	if !IsLocked(dir) {
		t.Error("expected locked (regular file SingletonLock)")
	}
}

func TestSmartResolve_ProfileOnly(t *testing.T) {
	dir := setupTestDir(t)
	r, _ := NewProfileResolver(dir)

	folder, userDataDir, err := r.SmartResolve("", "ReadingList-Auto", dir, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if folder != "Profile 5" {
		t.Errorf("expected Profile 5, got %s", folder)
	}
	if userDataDir != dir {
		t.Errorf("expected %s, got %s", dir, userDataDir)
	}
}

func TestSmartResolve_Neither(t *testing.T) {
	dir := setupTestDir(t)
	r, _ := NewProfileResolver(dir)

	folder, userDataDir, err := r.SmartResolve("", "", dir, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if folder != "Default" {
		t.Errorf("expected Default, got %s", folder)
	}
	if userDataDir != dir {
		t.Errorf("expected %s, got %s", dir, userDataDir)
	}
}

func TestSmartResolve_EmailOnly_Unlocked(t *testing.T) {
	dir := setupTestDir(t)
	r, _ := NewProfileResolver(dir)

	folder, userDataDir, err := r.SmartResolve("work@company.com", "", "", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if folder != "Profile 1" {
		t.Errorf("expected Profile 1, got %s", folder)
	}
	if userDataDir != dir {
		t.Errorf("expected %s, got %s", dir, userDataDir)
	}
}

func TestSmartResolve_EmailAndProfile(t *testing.T) {
	dir := setupTestDir(t)
	r, _ := NewProfileResolver(dir)

	folder, _, err := r.SmartResolve("kouko.d@gmail.com", "ReadingList-Auto", "", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if folder != "Profile 5" {
		t.Errorf("expected Profile 5, got %s", folder)
	}
}

func TestSmartResolve_EmailOnly_PreferRlssDir(t *testing.T) {
	// Create two dirs: one with "rlss" in path, one without
	rlssDir := filepath.Join(t.TempDir(), "rlss-chrome-data")
	systemDir := filepath.Join(t.TempDir(), "Google-Chrome")
	os.MkdirAll(rlssDir, 0755)
	os.MkdirAll(systemDir, 0755)

	localStateRlss := `{
		"profile": {
			"info_cache": {
				"Default": {"name": "RlssProfile", "user_name": "test@example.com"}
			}
		}
	}`
	localStateSystem := `{
		"profile": {
			"info_cache": {
				"Default": {"name": "SystemProfile", "user_name": "test@example.com"}
			}
		}
	}`
	os.WriteFile(filepath.Join(rlssDir, "Local State"), []byte(localStateRlss), 0644)
	os.WriteFile(filepath.Join(systemDir, "Local State"), []byte(localStateSystem), 0644)

	r, err := NewProfileResolver(systemDir, rlssDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, userDataDir, err := r.SmartResolve("test@example.com", "", "", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if userDataDir != rlssDir {
		t.Errorf("expected rlss dir %s, got %s", rlssDir, userDataDir)
	}
}

func TestSmartResolve_Locked_NoClone(t *testing.T) {
	dir := setupTestDir(t)
	// Create lock
	lockPath := filepath.Join(dir, "SingletonLock")
	if err := os.Symlink("host-123", lockPath); err != nil {
		t.Skipf("cannot create symlink: %v", err)
	}

	r, _ := NewProfileResolver(dir)

	_, _, err := r.SmartResolve("work@company.com", "", "", false)
	if err == nil {
		t.Error("expected error when locked without clone_profile")
	}
}

func TestSmartResolve_Locked_WithClone(t *testing.T) {
	dir := setupTestDir(t)

	// Create profile folder with essential files
	profileDir := filepath.Join(dir, "Profile 1")
	os.MkdirAll(profileDir, 0755)
	os.WriteFile(filepath.Join(profileDir, "Preferences"), []byte(`{"test": true}`), 0644)
	os.WriteFile(filepath.Join(profileDir, "Cookies"), []byte("cookies"), 0644)
	syncDir := filepath.Join(profileDir, "Sync Data")
	os.MkdirAll(syncDir, 0755)
	os.WriteFile(filepath.Join(syncDir, "SyncData.sqlite3"), []byte("sync"), 0644)

	// Create lock
	lockPath := filepath.Join(dir, "SingletonLock")
	if err := os.Symlink("host-123", lockPath); err != nil {
		t.Skipf("cannot create symlink: %v", err)
	}

	r, _ := NewProfileResolver(dir)

	folder, cloneDir, err := r.SmartResolve("work@company.com", "", "", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if folder != "Profile 1" {
		t.Errorf("expected Profile 1, got %s", folder)
	}
	// Clone dir should be deterministic temp path
	expectedCloneDir := cloneTempDir("Profile 1")
	if cloneDir != expectedCloneDir {
		t.Errorf("expected clone dir %s, got %s", expectedCloneDir, cloneDir)
	}
	// Verify cloned files exist
	if _, err := os.Stat(filepath.Join(cloneDir, "Local State")); err != nil {
		t.Error("Local State not cloned")
	}
	if _, err := os.Stat(filepath.Join(cloneDir, "Profile 1", "Preferences")); err != nil {
		t.Error("Preferences not cloned")
	}
	if _, err := os.Stat(filepath.Join(cloneDir, "Profile 1", "Cookies")); err != nil {
		t.Error("Cookies not cloned")
	}
	if _, err := os.Stat(filepath.Join(cloneDir, "Profile 1", "Sync Data", "SyncData.sqlite3")); err != nil {
		t.Error("Sync Data not cloned")
	}
	// No SingletonLock in clone
	if IsLocked(cloneDir) {
		t.Error("clone dir should not have SingletonLock")
	}
	// Cleanup
	os.RemoveAll(cloneDir)
}

func TestSmartResolve_NoMatchEmail(t *testing.T) {
	dir := setupTestDir(t)
	r, _ := NewProfileResolver(dir)

	_, _, err := r.SmartResolve("nobody@example.com", "", "", false)
	if err == nil {
		t.Error("expected error for unmatched email")
	}
}
