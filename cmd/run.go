package cmd

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"

	"scheduler/internal/service"
	"scheduler/pkg/containerd"
	pb "scheduler/proto/gen"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Start the scheduler server",
	Long: `Start the gRPC scheduler server that manages containerD containers.
The server will listen for incoming requests and handle container deployment operations.`,
	Run: runServer,
}

func init() {
	rootCmd.AddCommand(runCmd)
}

func runServer(cmd *cobra.Command, args []string) {
	host := viper.GetString("host")
	port := viper.GetInt("port")
	address := fmt.Sprintf("%s:%d", host, port)

	// Setup logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Setup containerD configuration
	containerdConfig := containerd.Config{
		Socket:    viper.GetString("containerd.socket"),
		Namespace: viper.GetString("containerd.namespace"),
	}

	// Create containerD manager
	containerdManager := containerd.NewManager(containerdConfig, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start containerD manager
	if err := containerdManager.Start(ctx); err != nil {
		log.Fatalf("Failed to start containerd manager: %v", err)
	}
	defer containerdManager.Stop()

	fmt.Printf("Starting scheduler server on %s\n", address)

	// Create gRPC server
	server := grpc.NewServer()

	// Create and register the scheduler service with containerD manager
	schedulerService := service.NewSchedulerService(containerdManager)
	pb.RegisterSchedulerServiceServer(server, schedulerService)

	// Create listener
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", address, err)
	}

	// Handle graceful shutdown
	// TODO: Use context for graceful shutdown of services

	// Start server in goroutine
	go func() {
		fmt.Printf("Server listening on %s\n", address)
		if err := server.Serve(listener); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	fmt.Println("\nShutting down server...")

	// Graceful shutdown
	server.GracefulStop()
	fmt.Println("Server stopped")
}
