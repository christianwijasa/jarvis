package update

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCheckDue(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	if !CheckDue(time.Time{}, now, DefaultCheckIntervalDays) {
		t.Fatal("missing stamp should be due")
	}
	if !CheckDue(now, now, 0) {
		t.Fatal("days=0 should be due every launch")
	}
	last := now.Add(-14 * 24 * time.Hour)
	if !CheckDue(last, now, 13) {
		t.Fatal("14 days later should be due")
	}
	fresh := now.Add(-time.Hour)
	if CheckDue(fresh, now, 13) {
		t.Fatal("fresh stamp should wait")
	}
}

func TestOfferUpdate(t *testing.T) {
	if OfferUpdate("dev", "v0.2.0") {
		t.Fatal("dev must not nag")
	}
	if OfferUpdate("v0.2.0", "v0.2.0") {
		t.Fatal("same tag must not nag")
	}
	if !OfferUpdate("v0.1.0", "v0.2.0") {
		t.Fatal("older tag should nag")
	}
	if OfferUpdate("v0.1.0", "") {
		t.Fatal("empty latest must not nag")
	}
}

func TestStampRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := StampPath(dir)
	if filepath.Base(path) != StampName {
		t.Fatalf("path %s", path)
	}
	got, err := ReadStamp(path)
	if err != nil || !got.IsZero() {
		t.Fatalf("missing stamp: %v %v", got, err)
	}
	now := time.Unix(1_700_000_000, 0)
	if err := WriteStamp(path, now); err != nil {
		t.Fatal(err)
	}
	got, err = ReadStamp(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.Unix() != now.Unix() {
		t.Fatalf("got %v want %v", got, now)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
}
