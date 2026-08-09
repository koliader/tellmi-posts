package posts_server

import (
	"context"

	"github.com/koliader/tellmi-posts/internal/lib/converter"
	db "github.com/koliader/tellmi-posts/internal/store/db/sqlc"
	errsvc "github.com/koliader/tellmi-sdk/errors/service"
	pb "github.com/koliader/tellmi-sdk/proto/pb"
	"google.golang.org/grpc/codes"
)

// Posts

func (s *Server) CreatePost(ctx context.Context, req *pb.CreatePostReq) (*pb.Post, error) {
	payload, err := s.middleware.AuthorizeUser(ctx)
	if err != nil {
		return nil, errsvc.ErrorResponse(codes.Unauthenticated, "error to authorize user: %v", err)
	}
	post, err := s.posts_service.CreatePost(ctx, req, payload)
	if err != nil {
		return nil, err
	}
	convertedPost := converter.ConvertPost(*post)
	return convertedPost, nil
}

func (s *Server) ListPosts(ctx context.Context, req *pb.PaginationReq) (*pb.ListPostsRes, error) {
	posts, err := s.posts_service.ListPosts(ctx, req)
	if err != nil {
		return nil, err
	}
	// convertedPosts := converter.ConverPostRows(*posts)
	// res := pb.ListPostsRes{Posts: convertedPosts}
	return posts, nil
}

func (s *Server) GetPostByID(ctx context.Context, req *pb.GetByIDReq) (*pb.PostRow, error) {
	post, err := s.posts_service.GetPostByID(ctx, req)
	if err != nil {
		return nil, err
	}
	return converter.ConvertListPostRow(db.ListPostsRow{
		ID:          post.ID,
		Title:       post.Title,
		Description: post.Description,
		UserID:      post.UserID,
		Username:    post.Username,
		CategoryID:  post.CategoryID,
		Name:        post.Name,
	}), nil
}

func (s *Server) EditPost(ctx context.Context, req *pb.EditPostReq) (*pb.Post, error) {
	payload, err := s.middleware.AuthorizeUser(ctx)
	if err != nil {
		return nil, errsvc.ErrorResponse(codes.Unauthenticated, "error to authorize user: %v", err)
	}

	post, err := s.posts_service.EditPost(ctx, req, payload)
	if err != nil {
		return nil, err
	}
	return converter.ConvertPost(*post), nil
}

func (s *Server) DeletePost(ctx context.Context, req *pb.GetByIDReq) (*pb.Success, error) {
	payload, err := s.middleware.AuthorizeUser(ctx)
	if err != nil {
		return nil, errsvc.ErrorResponse(codes.Unauthenticated, "error to authorize user: %v", err)
	}

	err = s.posts_service.DeletePost(ctx, req, payload)
	if err != nil {
		return nil, err
	}
	return &pb.Success{Message: "post deleted"}, nil
}

// Categories

// TODO access only for ADMIN
func (s *Server) CreateCategory(ctx context.Context, req *pb.CreateCategoryReq) (*pb.Category, error) {
	_, err := s.middleware.AuthorizeAdmin(ctx)
	if err != nil {
		return nil, errsvc.ErrorResponse(codes.Unauthenticated, "error to authorize admin: %v", err)

	}
	category, err := s.categories_service.CreateCategory(ctx, req)
	if err != nil {
		return nil, err
	}
	return converter.ConvertCategory(*category), nil
}

func (s *Server) ListCategories(ctx context.Context, req *pb.Empty) (*pb.ListCategoriesRes, error) {
	categories, err := s.categories_service.ListCategories(ctx)
	if err != nil {
		return nil, err
	}
	return &pb.ListCategoriesRes{Categories: converter.ConvertCategories(*categories)}, nil
}

// TODO access only for ADMIN
func (s *Server) EditCategory(ctx context.Context, req *pb.EditCategoryReq) (*pb.Success, error) {
	_, err := s.middleware.AuthorizeAdmin(ctx)
	if err != nil {
		return nil, errsvc.ErrorResponse(codes.Unauthenticated, "error to authorize admin: %v", err)
	}

	err = s.categories_service.EditCategory(ctx, req)
	if err != nil {
		return nil, err
	}
	return &pb.Success{Message: "category updated"}, nil
}

// Comments

func (s *Server) CreateComment(ctx context.Context, req *pb.CreateCommentReq) (*pb.Comment, error) {
	payload, err := s.middleware.AuthorizeUser(ctx)
	if err != nil {
		return nil, errsvc.ErrorResponse(codes.Unauthenticated, "error to authorize user: %v", err)
	}

	comment, err := s.comments_service.CreateComment(ctx, req, payload)
	if err != nil {
		return nil, err
	}
	return converter.ConvertComment(*comment), nil
}

func (s *Server) ListCommentsByPost(ctx context.Context, req *pb.GetByIDReq) (*pb.ListCommentsRes, error) {
	comments, err := s.comments_service.ListCommentsByPost(ctx, req)
	if err != nil {
		return nil, err
	}
	convertedComments := converter.ConvertCommentRows(*comments)
	return &pb.ListCommentsRes{Comments: convertedComments}, nil
}

func (s *Server) EditComment(ctx context.Context, req *pb.EditCommentReq) (*pb.Comment, error) {
	payload, err := s.middleware.AuthorizeUser(ctx)
	if err != nil {
		return nil, errsvc.ErrorResponse(codes.Unauthenticated, "error to authorize user: %v", err)
	}

	comment, err := s.comments_service.EditComment(ctx, req, payload)
	if err != nil {
		return nil, err
	}
	return converter.ConvertComment(*comment), nil
}

func (s *Server) DeleteComment(ctx context.Context, req *pb.GetByIDReq) (*pb.Success, error) {
	payload, err := s.middleware.AuthorizeUser(ctx)
	if err != nil {
		return nil, errsvc.ErrorResponse(codes.Unauthenticated, "error to authorize user: %v", err)
	}

	err = s.comments_service.DeleteComment(ctx, req, payload)
	if err != nil {
		return nil, err
	}
	return &pb.Success{Message: "comment deleted"}, nil
}
