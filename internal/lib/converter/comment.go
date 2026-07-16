package converter

import (
	"sync"

	pb "github.com/koliader/tellmi-posts/internal/pb"
	db "github.com/koliader/tellmi-posts/internal/store/db/sqlc"
)

func ConvertComment(comment db.Comment) *pb.Comment {
	return &pb.Comment{
		Id:      comment.ID,
		Comment: comment.Comment,
		PostId:  comment.PostID,
		Author:  comment.Author,
	}
}

var commentsPool = sync.Pool{
	New: func() interface{} {
		return &db.Comment{}
	},
}

func ConvertComments(comments []db.Comment) []*pb.Comment {
	converted := make([]*pb.Comment, 0, len(comments))
	for _, comment := range comments {
		c := commentsPool.Get().(*db.Comment)
		*c = comment
		converted = append(converted, ConvertComment(*c))
		commentsPool.Put(c)
	}
	return converted
}
