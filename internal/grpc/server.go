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

func StartGRPC(
	orchestrator orchestrator.Service,

) {
	lis, err := net.Listen("tcp", ":6767")
	if err != nil {
		log.Fatalf("listen failed: %v", err)
	}

	server := grpc.NewServer()

	pb.RegisterSidecarServiceServer(
		server,
		&SidecarServer{
			orchestrator: orchestrator,
		},
	)

	go func() {
		log.Println("grpc listening on :6767")
		if err := server.Serve(lis); err != nil {
			log.Fatalf("grpc failed: %v", err)
		}
	}()
}
