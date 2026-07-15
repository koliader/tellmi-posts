package categories_service

import (
	"github.com/koliader/tellmi-posts/internal/lib/config"
	db "github.com/koliader/tellmi-posts/internal/store/db/sqlc"
)

type Service struct {
	config config.Config
	store  db.Store
}

func NewServer(config config.Config, store db.Store) *Service {
	return &Service{config, store}
}
