package converter

import (
	"sync"

	pb "github.com/koliader/tellmi-sdk/proto/pb"
	db "github.com/koliader/tellmi-posts/internal/store/db/sqlc"
)

func ConvertComment(comment db.Comment) *pb.Comment {
	return &pb.Comment{
		Id:      comment.ID,
		Comment: comment.Comment,
		PostId:  comment.PostID,
		UserId:  comment.UserID,
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

func ConvertCommentRow(row db.ListCommentsByPostRow) *pb.CommentRow {
	user := pb.User{Id: row.UserID, Username: row.CommenterUsername}
	category := pb.Category{Id: row.CategoryID, Name: row.CategoryName}
	postAuthor := pb.User{Id: row.PostAuthorID, Username: row.PostAuthorUsername}
	post := pb.PostRow{
		Id:          row.PostID,
		Title:       row.PostTitle,
		Description: row.PostDescription,
		User:        &postAuthor,
		Category:    &category,
	}
	return &pb.CommentRow{
		Id:      row.ID,
		Comment: row.Comment,
		PostId:  row.PostID,
		User:    &user,
		Post:    &post,
	}
}

var commentRowPool = sync.Pool{
	New: func() any {
		return &db.ListCommentsByPostRow{}
	},
}

func ConvertCommentRows(rows []db.ListCommentsByPostRow) []*pb.CommentRow {
	converted := make([]*pb.CommentRow, 0, len(rows))
	for _, row := range rows {
		r := commentRowPool.Get().(*db.ListCommentsByPostRow)
		*r = row
		converted = append(converted, ConvertCommentRow(*r))
		commentRowPool.Put(r)
	}
	return converted
}
