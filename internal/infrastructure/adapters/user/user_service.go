package user

import (
	"github.com/ivan-salazar14/markerTradeIa/internal/domain"
)

// mockUsers is a package-level variable to store mock users for easy add/remove
var mockUsers = []domain.User{
	{UID: "1", Strategy: "conservative", ApiKey: "|mockApiKey|"},
	{UID: "2", Strategy: "aggressive", ApiKey: "|mockApiKey1|"},
	{UID: "3", Strategy: "conservative", ApiKey: "|mockApiKey2|"},
	{UID: "4", Strategy: "conservative", ApiKey: "|mockApiKey3|"},
	{UID: "5", Strategy: "aggressive", ApiKey: "|mockApiKey4|"},
	{UID: "6", Strategy: "conservative", ApiKey: "|mockApiKey5|"},
	{UID: "7", Strategy: "aggressive", ApiKey: "|mockApiKey6|"},
	{UID: "8", Strategy: "conservative", ApiKey: "|mockApiKey7|"},
	{UID: "9", Strategy: "aggressive", ApiKey: "|mockApiKey8|"},
	{UID: "10", Strategy: "conservative", ApiKey: "|mockApiKey9|"},
	{UID: "11", Strategy: "aggressive", ApiKey: "|mockApiKey10|"},
	{UID: "12", Strategy: "conservative", ApiKey: "|mockApiKey11|"},
	{UID: "13", Strategy: "aggressive", ApiKey: "|mockApiKey12|"},
	{UID: "14", Strategy: "conservative", ApiKey: "|mockApiKey13|"},
	{UID: "15", Strategy: "aggressive", ApiKey: "|mockApiKey14|"},
	{UID: "16", Strategy: "conservative", ApiKey: "|mockApiKey15|"},
	{UID: "17", Strategy: "aggressive", ApiKey: "|mockApiKey16|"},
	{UID: "18", Strategy: "conservative", ApiKey: "|mockApiKey17|"},
	{UID: "19", Strategy: "aggressive", ApiKey: "|mockApiKey18|"},
	{UID: "20", Strategy: "conservative", ApiKey: "|mockApiKey19|"},
}

/*
// Ejemplo de uso para agregar un usuario:
AddMockUser(domain.User{UID: "21", Strategy: "aggressive", ApiKey: "|mockApiKey20|"})

// Ejemplo de uso para eliminar un usuario:
RemoveMockUser("5")
*/

// AddMockUser adds a new user to the mockUsers slice
func AddMockUser(user domain.User) {
	mockUsers = append(mockUsers, user)
}

// RemoveMockUser removes a user by UID from the mockUsers slice
func RemoveMockUser(uid string) {
	for i, u := range mockUsers {
		if u.UID == uid {
			mockUsers = append(mockUsers[:i], mockUsers[i+1:]...)
			return
		}
	}
}

type HttpUserService struct {
	Endpoint string
}

func NewHttpUserService(endpoint string) *HttpUserService {
	return &HttpUserService{Endpoint: endpoint}
}

func (s *HttpUserService) GetUsers() ([]domain.User, error) {
	// Return the current mock users
	return mockUsers, nil
}
