package main

import (
	pb "NetCentricLab-MangaHub/internal/grpc"
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.Dial("127.0.0.1:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil { log.Fatalf("did not connect: %v", err) }
	defer conn.Close()

	client := pb.NewMangaServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	fmt.Println("🛰️ Sending gRPC request for Manga ID: one-piece...")
	r, err := client.GetManga(ctx, &pb.GetMangaRequest{Id: "one-piece"})
	if err != nil { log.Fatalf("Error: %v", err) }

	fmt.Printf("✅ Success! Title: %s | Chapters: %d\n", r.GetTitle(), r.GetTotalChapters())
}