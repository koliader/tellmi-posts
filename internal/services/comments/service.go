package comments_service

import (
	"github.com/koliader/tellmi-posts/internal/lib/config"
	db "github.com/koliader/tellmi-posts/internal/store/db/sqlc"
)

type Service struct {
	store  db.Store
	config config.Config
}

func NewService(store db.Store, config config.Config) *Service {
	return &Service{store, config}
}
