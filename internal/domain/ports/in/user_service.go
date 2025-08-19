package in

import (
	"github.com/ivan-salazar14/markerTradeIa/internal/domain"
)

type UserServicePort interface {
	GetUsers() ([]domain.User, error)
}
