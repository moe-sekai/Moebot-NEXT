package logbuffer

import (
	"fmt"
	"strings"
	"testing"
)

func writeLine(t *testing.T, b *Buffer, level, message string, extra ...string) {
	t.Helper()
	parts := []string{
		fmt.Sprintf("\"level\":\"%s\"", level),
		fmt.Sprintf("\"message\":\"%s\"", message),
	}
	parts = append(parts, extra...)
	line := "{" + strings.Join(parts, ",") + "}\n"
	if _, err := b.Write([]byte(line)); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func TestBufferRingOverwrite(t *testing.T) {
	b := New(minCapacity)
	for i := 0; i < minCapacity+50; i++ {
		writeLine(t, b, "info", fmt.Sprintf("msg-%d", i))
	}
	total, dropped, nextSeq := b.Stats()
	if total != minCapacity {
		t.Fatalf("expected size=%d, got %d", minCapacity, total)
	}
	if dropped != 50 {
		t.Fatalf("expected dropped=50, got %d", dropped)
	}
	if nextSeq != uint64(minCapacity+50) {
		t.Fatalf("unexpected nextSeq %d", nextSeq)
	}

	entries, _ := b.Snapshot(FilterOpts{Limit: 5})
	if len(entries) != 5 {
		t.Fatalf("expected 5 entries, got %d", len(entries))
	}
	if entries[0].Message != fmt.Sprintf("msg-%d", minCapacity+49) {
		t.Fatalf("expected newest first, got %q", entries[0].Message)
	}
	if entries[0].Seq <= entries[1].Seq {
		t.Fatalf("seq must descend: %d vs %d", entries[0].Seq, entries[1].Seq)
	}
}

func TestBufferFilters(t *testing.T) {
	b := New(minCapacity)
	writeLine(t, b, "info", "hello world")
	writeLine(t, b, "warn", "something off", "\"region\":\"jp\"")
	writeLine(t, b, "error", "boom!", "\"err\":\"failed\"")

	entries, total := b.Snapshot(FilterOpts{Levels: []string{"warn", "error"}})
	if total != 3 {
		t.Fatalf("total=%d", total)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 filtered entries, got %d", len(entries))
	}

	entries, _ = b.Snapshot(FilterOpts{Query: "WORLD"})
	if len(entries) != 1 || entries[0].Message != "hello world" {
		t.Fatalf("unexpected query result: %+v", entries)
	}

	entries, _ = b.Snapshot(FilterOpts{Query: "jp"})
	if len(entries) != 1 || entries[0].Level != "warn" {
		t.Fatalf("expected fields query to hit warn entry, got %+v", entries)
	}
}

func TestBufferIncremental(t *testing.T) {
	b := New(minCapacity)
	for i := 0; i < 5; i++ {
		writeLine(t, b, "info", fmt.Sprintf("m%d", i))
	}
	first, _ := b.Snapshot(FilterOpts{})
	if len(first) != 5 {
		t.Fatalf("expected 5, got %d", len(first))
	}
	since := first[0].Seq // newest
	writeLine(t, b, "info", "after")
	entries, _ := b.Snapshot(FilterOpts{SinceSeq: since})
	if len(entries) != 1 || entries[0].Message != "after" {
		t.Fatalf("expected only new entry, got %+v", entries)
	}
}

func TestBufferRawFallback(t *testing.T) {
	b := New(minCapacity)
	if _, err := b.Write([]byte("not json line\n")); err != nil {
		t.Fatalf("write: %v", err)
	}
	entries, _ := b.Snapshot(FilterOpts{})
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry")
	}
	if entries[0].Message != "not json line" || entries[0].Level != "info" {
		t.Fatalf("unexpected raw fallback: %+v", entries[0])
	}
}
