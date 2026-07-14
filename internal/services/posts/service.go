package posts_service

import (
	"github.com/koliadertellmi-posts/internal/lib/config"
	db "github.com/koliadertellmi-posts/internal/store/db/sqlc"
)

type Service struct {
	store  db.Store
	config config.Config
}

func NewService(store db.Store, config config.Config) *Service {
	return &Service{store, config}
}
