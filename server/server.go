package server

import (
	"gitbeam.commit.monitor/core"
	commits "gitbeam.commit.monitor/pb"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
)

func ExecGRPCServer(address string, core *core.GitBeamService, logger *logrus.Logger) {
	defer func() {
		if err := recover(); err != nil {
			log.Fatalf("Recovered from err: %v", err)
		}
	}()
	api := NewApiService(core, logger)

	server := grpc.NewServer()
	commits.RegisterGitBeamCommitsServiceServer(server, api)
	reflection.Register(server)

	lis, err := net.Listen("tcp", address)
	if err != nil {
		logrus.Fatalf("failed to listen with the following errors: %v", err)
	}
	if err := server.Serve(lis); err != nil {
		logrus.Fatal(err)
	}
}

//// Channel to listen for signals
//	signalChan := make(chan os.Signal, 1)
//	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
//
//	// Channel to notify the server has been stopped
//	shutdownChan := make(chan bool)
//
//	// Start server in a goroutine
//	go func() {
//		logger.Info("Started Server")
//		if err := server.ListenAndServe(); err != nil {
//			logger.WithError(err).Error("failed to start server")
//			fmt.Printf("ListenAndServe(): %s\n", err)
//		}
//	}()
//
//	// Listen for shutdown signal
//	go func() {
//		<-signalChan
//		logger.Info("Shutting down server...")
//
//		// Create a deadline to wait for.
//		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//		defer cancel()
//
//		// Attempt to gracefully shut down the server
//		if err := server.Shutdown(ctx); err != nil {
//			logger.WithError(err).Error("Server forced to shutdown")
//		}
//
//		close(shutdownChan)
//	}()
//
//	// Wait for shutdown signal
//	<-shutdownChan
//	logger.Info("Server gracefully stopped...")
