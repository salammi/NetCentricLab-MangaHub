package main

import (
	pb "NetCentricLab-MangaHub/internal/grpc"
	"NetCentricLab-MangaHub/pkg/database"
	"log"
	"net"

	"google.golang.org/grpc"
)

func main() {
	database.InitDB("data.db")
	lis, err := net.Listen("tcp", ":50051")
	if err != nil { log.Fatalf("Failed to listen: %v", err) }

	grpcServer := grpc.NewServer()
	pb.RegisterMangaServiceServer(grpcServer, pb.NewMangaServer(database.DB))

	log.Println("🚀 gRPC Server is running on port 50051...")
	grpcServer.Serve(lis)
}