package cmd

import (
	"strings"
	"testing"
	"time"
)

func TestParseEventTime(t *testing.T) {
	cases := []struct {
		in   string
		want string // RFC3339 UTC; empty means "anything", caller validates instead
	}{
		{"2026-05-18", "2026-05-18T00:00:00Z"},
		{"2026-05-18T08:30", "2026-05-18T08:30:00Z"},
		{"2026-05-18T08:30:00", "2026-05-18T08:30:00Z"},
		{"2026-05-18 08:30", "2026-05-18T08:30:00Z"},
		{"2026-05-18T08:29:57Z", "2026-05-18T08:29:57Z"},
		{"2026-05-18T08:29:57+07:00", "2026-05-18T01:29:57Z"},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			got, err := parseEventTime(c.in)
			if err != nil {
				t.Fatalf("parseEventTime(%q) error: %v", c.in, err)
			}
			if got.Format(time.RFC3339) != c.want {
				t.Fatalf("parseEventTime(%q) = %s, want %s", c.in, got.Format(time.RFC3339), c.want)
			}
		})
	}

	t.Run("now", func(t *testing.T) {
		got, err := parseEventTime("now")
		if err != nil {
			t.Fatalf("parseEventTime(\"now\") error: %v", err)
		}
		if time.Since(got) > 5*time.Second {
			t.Fatalf("parseEventTime(\"now\") returned stale time: %s", got)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		if _, err := parseEventTime("not-a-date"); err == nil {
			t.Fatal("expected error for invalid input")
		}
	})
}

func TestParseEventDuration(t *testing.T) {
	cases := map[string]time.Duration{
		"48h":     48 * time.Hour,
		"2h30m":   2*time.Hour + 30*time.Minute,
		"30m":     30 * time.Minute,
		"1d":      24 * time.Hour,
		"3d":      3 * 24 * time.Hour,
		"7d":      7 * 24 * time.Hour,
	}
	for in, want := range cases {
		got, err := parseEventDuration(in)
		if err != nil {
			t.Fatalf("parseEventDuration(%q) error: %v", in, err)
		}
		if got != want {
			t.Fatalf("parseEventDuration(%q) = %s, want %s", in, got, want)
		}
	}

	for _, bad := range []string{"", "abc", "0h", "-1h", "0d"} {
		if _, err := parseEventDuration(bad); err == nil {
			t.Fatalf("parseEventDuration(%q) should fail", bad)
		}
	}
}

func TestResolveEventTimes_AllDefaults(t *testing.T) {
	startStr, endStr, err := resolveEventTimes("", "", "")
	if err != nil {
		t.Fatalf("resolveEventTimes error: %v", err)
	}
	start, err := time.Parse(time.RFC3339, startStr)
	if err != nil {
		t.Fatalf("parse start %q: %v", startStr, err)
	}
	end, err := time.Parse(time.RFC3339, endStr)
	if err != nil {
		t.Fatalf("parse end %q: %v", endStr, err)
	}
	if d := end.Sub(start); d != defaultEventDuration {
		t.Fatalf("end - start = %s, want %s", d, defaultEventDuration)
	}
	if time.Since(start) > 5*time.Second {
		t.Fatalf("start is not ~now: %s", startStr)
	}
}

func TestResolveEventTimes_StartAndDuration(t *testing.T) {
	startStr, endStr, err := resolveEventTimes("2026-05-18", "", "3d")
	if err != nil {
		t.Fatalf("resolveEventTimes error: %v", err)
	}
	if startStr != "2026-05-18T00:00:00Z" {
		t.Fatalf("start = %s, want 2026-05-18T00:00:00Z", startStr)
	}
	if endStr != "2026-05-21T00:00:00Z" {
		t.Fatalf("end = %s, want 2026-05-21T00:00:00Z", endStr)
	}
}

func TestResolveEventTimes_EndWinsOverDuration(t *testing.T) {
	_, endStr, err := resolveEventTimes("2026-05-18", "2026-05-20", "999h")
	if err != nil {
		t.Fatalf("resolveEventTimes error: %v", err)
	}
	if endStr != "2026-05-20T00:00:00Z" {
		t.Fatalf("end = %s, want 2026-05-20T00:00:00Z", endStr)
	}
}

func TestResolveEventTimes_EndBeforeStart(t *testing.T) {
	_, _, err := resolveEventTimes("2026-05-20", "2026-05-18", "")
	if err == nil || !strings.Contains(err.Error(), "must be after") {
		t.Fatalf("expected end-before-start error, got: %v", err)
	}
}

func TestResolveEventTimes_PropagatesFlagName(t *testing.T) {
	_, _, err := resolveEventTimes("nope", "", "")
	if err == nil || !strings.Contains(err.Error(), "--start") {
		t.Fatalf("expected error mentioning --start, got: %v", err)
	}
	_, _, err = resolveEventTimes("", "nope", "")
	if err == nil || !strings.Contains(err.Error(), "--end") {
		t.Fatalf("expected error mentioning --end, got: %v", err)
	}
	_, _, err = resolveEventTimes("", "", "nope")
	if err == nil || !strings.Contains(err.Error(), "--duration") {
		t.Fatalf("expected error mentioning --duration, got: %v", err)
	}
}
