package grpcserver

import (
	"context"
	"log"
	"time"

	pb "github.com/nanagoboiler/gen"
	"github.com/nanagoboiler/models"
)

func (s *SidecarServer) Connect(stream pb.SidecarService_ConnectServer) error {
	log.Println("sidecar connected")

	firstEvt, err := stream.Recv()
	if err != nil {
		log.Printf("failed to receive first event: %v", err)
		return err
	}

	serverID := firstEvt.GetServerId()

	s.orchestrator.RegisterStream(serverID, stream)
	defer s.orchestrator.UnregisterStream(serverID)

	if payload, ok := firstEvt.Payload.(*pb.SidecarEvent_ServerStarted); ok {
		ctx, cancel := context.WithTimeout(stream.Context(), 1000*time.Millisecond)
		defer cancel()
		if err := s.orchestrator.UpdateServerStatus(ctx, serverID, models.ServerReady); err != nil {
			log.Println(err)
		}
		log.Printf("server %s started: hostname=%s", serverID, payload.ServerStarted.Hostname)
	} else {
		log.Printf("expected ServerStarted as first event, got %T", firstEvt.Payload)
	}

	for {
		evt, err := stream.Recv()
		if err != nil {
			log.Printf("sidecar disconnected: %v", err)
			return err
		}

		switch payload := evt.Payload.(type) {
		case *pb.SidecarEvent_Heartbeat:

			ctx, cancel := context.WithTimeout(stream.Context(), 500*time.Millisecond)
			defer cancel()

			err := s.orchestrator.UpdateHeartbeat(ctx, serverID)
			if err != nil {
				log.Println(err)
			}
			log.Printf("heartbeat from %s", serverID)
		case *pb.SidecarEvent_ServerStarted:
			log.Println("Bing Bong")

			ctx, cancel := context.WithTimeout(stream.Context(), 1000*time.Millisecond)
			defer cancel()
			err := s.orchestrator.UpdateServerStatus(ctx, serverID, models.ServerReady)
			if err != nil {
				log.Println(err)
			}
		case *pb.SidecarEvent_ServerStopped:
			log.Println("Here We Would Delete The Server")
		default:
			log.Printf("unhandled event type %T from %s", payload, evt.GetServerId())
		}
	}
}
