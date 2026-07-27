package main

import (
	"bytes"
	"cmd/internal/archive"
	"os"
	"path/filepath"
	"testing"
)

func TestToolFlag(t *testing.T) {
	args := []string{"-p=example.com/p", "-o", "out.a", "-shared"}
	for _, test := range []struct {
		name string
		want string
		ok   bool
	}{
		{"-p", "example.com/p", true},
		{"-o", "out.a", true},
		{"-goversion", "", false},
	} {
		got, ok := toolFlag(args, test.name)
		if got != test.want || ok != test.ok {
			t.Errorf("toolFlag(%q) = %q, %v; want %q, %v", test.name, got, ok, test.want, test.ok)
		}
	}
}

func TestAppendArchiveMember(t *testing.T) {
	dir := t.TempDir()
	archivePath := filepath.Join(dir, "p.a")
	member := filepath.Join(dir, "member.o")
	arFile, err := os.OpenFile(archivePath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	ar, err := archive.New(arFile)
	if err != nil {
		t.Fatal(err)
	}
	ar.AddEntry(archive.EntryPkgDef, "__.PKGDEF", 0, 0, 0, 0o644, 6, bytes.NewReader([]byte("pkgdef")))
	if err := arFile.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(member, []byte("native-o"), 0o600); err != nil {
		t.Fatal(err)
	}
	appendArchiveMember(archivePath, "_go_.o", member)
	arFile, err = os.Open(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	defer arFile.Close()
	parsed, err := archive.Parse(arFile, false)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(parsed.Entries), 2; got != want {
		t.Fatalf("archive entries = %d, want %d", got, want)
	}
	if got, want := parsed.Entries[0].Name, "__.PKGDEF"; got != want {
		t.Fatalf("first archive member = %q, want %q", got, want)
	}
	if got, want := parsed.Entries[1].Name, "_go_.o"; got != want {
		t.Fatalf("second archive member = %q, want %q", got, want)
	}
}
