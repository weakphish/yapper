package cmd

import (
	"context"
	"os"

	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
)

func Execute() {
	cmd := &cobra.Command{
		Use:   "yapper",
		Short: "Yapper is a terminal-based task management application",
	}
	if err := fang.Execute(context.TODO(), cmd); err != nil {
		os.Exit(1)
	}
}
