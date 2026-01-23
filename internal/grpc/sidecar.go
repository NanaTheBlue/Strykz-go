package grpcserver

import (
	"context"
	"log"
	"time"

	pb "github.com/nanagoboiler/gen"
)

func (s *SidecarServer) Connect(stream pb.SidecarService_ConnectServer) error {
	log.Println("sidecar connected")

	for {
		evt, err := stream.Recv()
		if err != nil {
			log.Printf("sidecar disconnected: %v", err)
			return err
		}

		switch payload := evt.Payload.(type) {
		case *pb.SidecarEvent_Heartbeat:
			serverID := evt.GetServerId()
			ctx, cancel := context.WithTimeout(stream.Context(), 500*time.Millisecond)
			defer cancel()

			err := s.orchestrator.UpdateHeartbeat(serverID, ctx)
			if err != nil {
				log.Println(err)
			}
			log.Printf("heartbeat from %s", serverID)
		case *pb.SidecarEvent_ServerStarted:
			log.Println("Bing Bong")

		default:
			log.Printf("unhandled event type %T from %s", payload, evt.GetServerId())
		}
	}
}
