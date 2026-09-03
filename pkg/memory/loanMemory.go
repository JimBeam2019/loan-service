package memory

import (
	"loan-service/internal/domain"
	"sync"

	"github.com/google/uuid"
)

type MemoryDB struct {
	Loans map[uuid.UUID]*domain.Loan
	Mtx   sync.RWMutex
}

func NewMemoryDB() *MemoryDB {
	return &MemoryDB{Loans: make(map[uuid.UUID]*domain.Loan)}
}
