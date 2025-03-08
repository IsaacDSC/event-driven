package sqlc

import (
	"database/sql"
	genrepo "github.com/IsaacDSC/event-driven/internal/sqlc/generated/repository"
)

func NewRepository(db *sql.DB) *genrepo.Queries {
	return genrepo.New(db)
}
