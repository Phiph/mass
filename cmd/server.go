package cmd

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"github.com/massdriver-cloud/mass/pkg/server"
	"github.com/spf13/cobra"
)

func NewCmdServer() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",
		Short: "Start the bundle development server",
		Long:  "Start the bundle development server. If no port is supplied an ephemeral port will be used",
		Run: func(cmd *cobra.Command, args []string) {
			runServer(cmd)
		},
		Args: cobra.NoArgs,
	}

	cmd.Flags().StringP("port", "p", "", "port for the server to listen on")
	cmd.Flags().StringP("directory", "d", "", "directory for the massdriver bundle, will default to the directory the server is ran from")

	return cmd
}

func runServer(cmd *cobra.Command) {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		for sig := range c {
			slog.Info("Shutting down", "signal", sig)
			// TODO: Add cleanup work here, that could be flushing current work or just shutting down
			// the server gracefully
			os.Exit(0)
		}
	}()

	dir, err := cmd.Flags().GetString("directory")
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	// Check we have a massdriver.yaml file available, if not error out.
	_, err = os.Stat(path.Join(dir, "massdriver.yaml"))
	if err != nil {
		slog.Error("massdriver.yaml file is not available in the specified directory")
		os.Exit(1)
	}

	port, err := cmd.Flags().GetString("port")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// If no port is supplied grab an ephemeral port
	if port == "" {
		port = "0"
	}

	// Setup our single handler
	server.RegisterServerHandler(dir)

	ln, err := net.Listen("tcp", "127.0.0.1:"+port)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	slog.Info(fmt.Sprintf("Visit http://%s/hello-agent in your browser", ln.Addr().String()))
	server := http.Server{ReadHeaderTimeout: 60 * time.Second}
	err = server.Serve(ln)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}