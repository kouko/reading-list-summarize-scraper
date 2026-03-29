package output

import (
	"crypto/sha256"
	"fmt"
	"net/url"
	"strings"
	"time"
)

func SHA8(s string) string {
	h := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", h[:4])
}

func DomainDir(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil || u.Host == "" {
		return "unknown"
	}
	host := u.Hostname()
	return strings.ReplaceAll(host, ".", "_")
}

func SummaryFilename(date time.Time, sha8 string) string {
	return fmt.Sprintf("%s__%s__summary.md", date.Format("2006-01-02"), sha8)
}

func ContentFilename(date time.Time, sha8 string) string {
	return fmt.Sprintf("%s__%s__content.md", date.Format("2006-01-02"), sha8)
}
