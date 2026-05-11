package grpc

import (
	"context"
	"database/sql"
	"errors"
)

type MangaServer struct {
	UnimplementedMangaServiceServer
	DB *sql.DB
}

func NewMangaServer(db *sql.DB) *MangaServer {
	return &MangaServer{DB: db}
}

func (s *MangaServer) GetManga(ctx context.Context, req *GetMangaRequest) (*MangaResponse, error) {
	var res MangaResponse
	err := s.DB.QueryRow("SELECT id, title, author, total_chapters FROM manga WHERE id = ?", req.Id).
		Scan(&res.Id, &res.Title, &res.Author, &res.TotalChapters)

	if err != nil {
		if err == sql.ErrNoRows { return nil, errors.New("manga not found") }
		return nil, err
	}
	return &res, nil
}