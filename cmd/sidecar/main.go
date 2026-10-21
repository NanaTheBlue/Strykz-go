package main

import (
	"bufio"
	"context"
	"flag"
	"io"
	"log"
	"os"
	"strings"
	"time"

	pb "github.com/nanagoboiler/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	serverID := flag.String("server-id", os.Getenv("SERVER_ID"), "The unique ID of this game server")
	orchestratorAddr := flag.String("orchestrator", os.Getenv("ORCHESTRATOR_ADDR"), "Address of the orchestrator gRPC server")
	logFile := flag.String("log-file", "cstrike/console.log", "Path to the SRCDS console.log")
	flag.Parse()

	if *serverID == "" {
		*serverID = "local-test-server-1"
	}
	if *orchestratorAddr == "" {
		*orchestratorAddr = "backend:6767"
	}

	log.Printf("Starting sidecar for server %s, connecting to %s", *serverID, *orchestratorAddr)

	conn, err := grpc.NewClient(*orchestratorAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer func() { _ = conn.Close() }()

	client := pb.NewSidecarServiceClient(conn)

	// Open stream with retry in case the backend gRPC server is still booting
	var stream pb.SidecarService_ConnectClient
	for {
		stream, err = client.Connect(context.Background())
		if err == nil {
			break
		}
		log.Printf("waiting for orchestrator stream (%v), retrying in 2s...", err)
		time.Sleep(2 * time.Second)
	}

	// 1. Send ServerStarted
	err = stream.Send(&pb.SidecarEvent{
		ServerId: *serverID,
		Payload: &pb.SidecarEvent_ServerStarted{
			ServerStarted: &pb.ServerStarted{Hostname: "css-docker-container"},
		},
	})
	if err != nil {
		log.Fatalf("failed to send ServerStarted: %v", err)
	}

	// 2. Start Heartbeat routine
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			err := stream.Send(&pb.SidecarEvent{
				ServerId: *serverID,
				Payload: &pb.SidecarEvent_Heartbeat{
					Heartbeat: &pb.Heartbeat{},
				},
			})
			if err != nil {
				log.Printf("failed to send heartbeat: %v", err)
				return // break if stream is broken
			}
		}
	}()

	// 3. Tail the log file and send LogLine events
	// Simple polling tail implementation
	for {
		file, err := os.Open(*logFile)
		if err != nil {
			log.Printf("Waiting for log file %s to be created...", *logFile)
			time.Sleep(2 * time.Second)
			continue
		}

		// Seek to end initially so we don't process old logs if sidecar restarts
		_, _ = file.Seek(0, io.SeekEnd)
		reader := bufio.NewReader(file)

		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				// EOF, wait and try again
				time.Sleep(500 * time.Millisecond)
				continue
			}

			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			err = stream.Send(&pb.SidecarEvent{
				ServerId: *serverID,
				Payload: &pb.SidecarEvent_LogLine{
					LogLine: &pb.LogLine{Raw: line},
				},
			})
			if err != nil {
				log.Printf("failed to send log line: %v", err)
			}
		}
	}
}
