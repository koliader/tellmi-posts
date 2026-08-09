package posts_service

import (
	"github.com/koliader/tellmi-posts/internal/lib/config"
	users_service "github.com/koliader/tellmi-posts/internal/services/users"
	db "github.com/koliader/tellmi-posts/internal/store/db/sqlc"
)

type Service struct {
	store         db.Store
	config        config.Config
	users_service *users_service.Service
}

func NewService(store db.Store, config config.Config) *Service {
	usersService := users_service.NewService(store, config)
	return &Service{store, config, usersService}
}
