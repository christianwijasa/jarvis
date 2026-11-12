package update

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// DefaultCheckIntervalDays matches oh-my-zsh's UPDATE_ZSH_DAYS default.
const DefaultCheckIntervalDays = 13

// StampName is the file under JARVIS_HOME that records the last network check.
const StampName = "update-check"

// StampPath is $JARVIS_HOME/update-check.
func StampPath(home string) string {
	return filepath.Join(home, StampName)
}

// ReadStamp returns the last check time. A missing file is a zero Time, not an error.
func ReadStamp(path string) (time.Time, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return time.Time{}, nil
		}
		return time.Time{}, err
	}
	line := strings.TrimSpace(string(body))
	if i := strings.IndexByte(line, '\n'); i >= 0 {
		line = strings.TrimSpace(line[:i])
	}
	if line == "" {
		return time.Time{}, nil
	}
	sec, err := strconv.ParseInt(line, 10, 64)
	if err != nil {
		return time.Time{}, nil
	}
	return time.Unix(sec, 0), nil
}

// WriteStamp records now (unix seconds) so the next nag waits the interval.
func WriteStamp(path string, now time.Time) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(strconv.FormatInt(now.Unix(), 10)+"\n"), 0o644)
}

// CheckDue is true when no stamp exists, days <= 0 (every launch), or the
// interval has elapsed.
func CheckDue(last, now time.Time, days int) bool {
	if days <= 0 || last.IsZero() {
		return true
	}
	return !last.After(now) && now.Sub(last) >= time.Duration(days)*24*time.Hour
}

// OfferUpdate is true when a tagged binary is behind latest. A "dev" build
// is a local compile and is never offered as an auto-update.
func OfferUpdate(current, latest string) bool {
	current = strings.TrimSpace(current)
	latest = strings.TrimSpace(latest)
	if current == "" || current == "dev" || latest == "" {
		return false
	}
	return current != latest
}
