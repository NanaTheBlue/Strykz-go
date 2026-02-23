package grpcserver

import (
	"log"
	"net"

	pb "github.com/nanagoboiler/gen"

	"github.com/nanagoboiler/internal/services/orchestrator"
	"google.golang.org/grpc"
)

type SidecarServer struct {
	pb.UnimplementedSidecarServiceServer
	orchestrator orchestrator.Service
}

func NewSidecarServer(orchestrator orchestrator.Service) *SidecarServer {
	return &SidecarServer{
		orchestrator: orchestrator,
	}
}

func StartGRPC(
	orchestrator orchestrator.Service,
	addr string,
) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("listen failed: %v", err)
	}

	server := grpc.NewServer()

	pb.RegisterSidecarServiceServer(
		server,
		NewSidecarServer(orchestrator),
	)

	go func() {
		log.Println("grpc listening on :6767")
		if err := server.Serve(lis); err != nil {
			log.Fatalf("grpc failed: %v", err)
		}
	}()
}
