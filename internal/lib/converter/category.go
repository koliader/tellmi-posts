package converter

import (
	"sync"

	pb "github.com/koliader/tellmi-posts/internal/pb"
	db "github.com/koliader/tellmi-posts/internal/store/db/sqlc"
)

func ConvertCategory(category db.Category) *pb.Category {
	return &pb.Category{
		Id:   category.ID,
		Name: category.Name,
	}
}

var categoriesPool = sync.Pool{
	New: func() interface{} {
		return &db.Category{}
	},
}

func ConvertCategories(categories []db.Category) []*pb.Category {
	converted := make([]*pb.Category, 0, len(categories))
	for _, category := range categories {
		c := categoriesPool.Get().(*db.Category)
		*c = category
		converted = append(converted, ConvertCategory(*c))
		categoriesPool.Put(c)
	}
	return converted
}
