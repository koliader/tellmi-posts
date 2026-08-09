package converter

import (
	"sync"

	db "github.com/koliader/tellmi-posts/internal/store/db/sqlc"
	pb "github.com/koliader/tellmi-sdk/proto/pb"
)

func ConvertPost(post db.Post) *pb.Post {
	return &pb.Post{
		Id:          post.ID,
		Title:       post.Title,
		Description: post.Description,
		UserId:      post.UserID.String(),
		CategoryId:  post.CategoryID,
	}
}

func ConvertListPostRow(post db.ListPostsRow) *pb.PostRow {
	category := pb.Category{Id: post.CategoryID, Name: post.Name}
	user := pb.User{Id: post.UserID.String(), Username: post.Username}
	return &pb.PostRow{
		Id:          post.ID,
		Title:       post.Title,
		Description: post.Description,
		User:        &user,
		Category:    &category,
	}
}

var postsRowPool = sync.Pool{
	New: func() any {
		return &db.ListPostsRow{}
	},
}

func ConvertPostRows(rows []db.ListPostsRow) []*pb.PostRow {
	converted := make([]*pb.PostRow, 0, len(rows))
	for _, row := range rows {
		r := postsRowPool.Get().(*db.ListPostsRow)
		*r = row
		converted = append(converted, ConvertListPostRow(*r))
		postsRowPool.Put(r)
	}
	return converted
}

var postsPool = sync.Pool{
	New: func() any {
		return &db.Post{}
	},
}

func ConvertPosts(posts []db.Post) []*pb.Post {
	converted := make([]*pb.Post, 0, len(posts))
	for _, post := range posts {
		p := postsPool.Get().(*db.Post)
		*p = post
		converted = append(converted, ConvertPost(*p))
		postsPool.Put(p)
	}
	return converted
}
