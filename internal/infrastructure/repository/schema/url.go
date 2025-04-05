package schema

import (
	"github.com/Rafiur/go-url-shortener/internal/domain/entity"
	"github.com/uptrace/bun"
	"time"
)

type URL struct {
	bun.BaseModel `bun:"table:urls"`
	ID            int64     `bun:"id,pk,autoincrement"`
	ShortCode     string    `bun:"short_code,unique,notnull"`
	OriginalURL   string    `bun:"original_url,notnull"`
	CreatedAt     time.Time `bun:"created_at,default:current_timestamp"`
}

func (s *URL) SchemaToEntity() *entity.URL {
	return &entity.URL{
		ID:          s.ID,
		ShortCode:   s.ShortCode,
		OriginalURL: s.OriginalURL,
		CreatedAt:   s.CreatedAt,
	}
}
