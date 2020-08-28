//go:generate go run assets-generator.go

package main

import (
	"prayer-display-web/internal/backend"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	// Include vfsgen to prevent it removed by go mod tidy
	_ "github.com/shurcooL/httpfs/filter"
	_ "github.com/shurcooL/vfsgen"
)

func main() {
	// Format logrus
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	// Create root cmd
	rootCmd := &cobra.Command{
		Use:   "prayer-display-web",
		Short: "Prayer display for Raspberry Pi, created using web technology",
		Run:   rootCmdHandler,
	}

	rootCmd.Flags().IntP("port", "p", 9001, "port used for the GUI")

	// Execute cmd
	err := rootCmd.Execute()
	if err != nil {
		logrus.Fatalln(err)
	}
}

func rootCmdHandler(cmd *cobra.Command, args []string) {
	// Get flags value
	port, _ := cmd.Flags().GetInt("port")

	// Serve app
	err := backend.ServeApp(port)
	if err != nil {
		logrus.Fatalf("Failed to serve app: %v", err)
	}
}
