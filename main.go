package main

import (
	"github.com/culnaen/ss-tulpn-go/cmd"
	"log/slog"
	"os"
)

func main() {
	if err := cmd.Execute(); err != nil {
		slog.Error("failed to run cmd", "err", err)
		os.Exit(1)
	}
}
