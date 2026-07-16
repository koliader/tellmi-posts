package converter

import (
	"sync"

	pb "github.com/koliader/tellmi-posts/internal/pb"
	db "github.com/koliader/tellmi-posts/internal/store/db/sqlc"
)

func ConvertPost(post db.Post) *pb.Post {
	return &pb.Post{
		Id:          post.ID,
		Title:       post.Title,
		Description: post.Description,
		Author:      post.Author,
	}
}

var postsPool = sync.Pool{
	New: func() interface{} {
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
