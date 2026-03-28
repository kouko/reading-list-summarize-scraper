package output

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/kouko/reading-list-summarize-scraper/internal/config"
)

type CopyToVars struct {
	OutputDir string
	Date      string
	DateAdded string
	Title     string
	SHA8      string
	Source    string
	Domain    string
	DomainDir string
	Type      string
}

func ExpandTemplate(tmpl string, vars CopyToVars) string {
	r := strings.NewReplacer(
		"{output_dir}", vars.OutputDir,
		"{date}", vars.Date,
		"{date_added}", vars.DateAdded,
		"{title}", vars.Title,
		"{sha8}", vars.SHA8,
		"{source}", vars.Source,
		"{domain}", vars.Domain,
		"{domain_dir}", vars.DomainDir,
		"{type}", vars.Type,
	)
	return r.Replace(tmpl)
}

var unsafeChars = regexp.MustCompile(`[\\/:*?"<>|]`)

var multiSpace = regexp.MustCompile(`\s{2,}`)

func SanitizeTitleForDisplay(title string) string {
	clean := unsafeChars.ReplaceAllString(title, " ")
	clean = multiSpace.ReplaceAllString(clean, " ")
	clean = strings.TrimSpace(clean)
	if utf8.RuneCountInString(clean) > 80 {
		runes := []rune(clean)
		clean = string(runes[:80])
	}
	return clean
}

func ExecuteCopyTo(cfg config.CopyToConfig, sourceDir string, sha8 string, vars CopyToVars) error {
	if !cfg.Enabled {
		return nil
	}

	for _, fileType := range cfg.Files {
		vars.Type = fileType
		srcPattern := fmt.Sprintf("*__%s__%s.md", sha8, fileType)

		matches, _ := filepath.Glob(filepath.Join(sourceDir, srcPattern))
		if len(matches) == 0 {
			slog.Warn("copy_to: source file not found", "pattern", srcPattern)
			continue
		}
		srcPath := matches[0]

		targetDir := ExpandTemplate(cfg.Path, vars)
		targetFile := ExpandTemplate(cfg.Filename, vars)
		targetPath := filepath.Join(targetDir, targetFile)

		if !cfg.Overwrite {
			if _, err := os.Stat(targetPath); err == nil {
				slog.Debug("copy_to: skip existing", "path", targetPath)
				continue
			}
		}

		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return fmt.Errorf("create dir %s: %w", targetDir, err)
		}

		if err := copyFile(srcPath, targetPath); err != nil {
			return fmt.Errorf("copy %s → %s: %w", srcPath, targetPath, err)
		}
		slog.Info("copy_to", "src", srcPath, "dst", targetPath)
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
