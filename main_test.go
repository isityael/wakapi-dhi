package main

import (
	"os"
	"strings"
	"testing"
)

func TestMainDoesNotEmbedSQLiteWasm(t *testing.T) {
	source, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(string(source), `"github.com/ncruces/go-sqlite3/embed"`) {
		t.Fatal("main binary must not import github.com/ncruces/go-sqlite3/embed")
	}
}
