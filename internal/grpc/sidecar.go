package grpcserver

import (
	"log"

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
			log.Printf("beat: %v", payload)

		}
	}
}
