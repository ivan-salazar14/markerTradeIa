package user

import (
	"github.com/ivan-salazar14/markerTradeIa/internal/domain"
)

type HttpUserService struct {
	Endpoint string
}

func NewHttpUserService(endpoint string) *HttpUserService {
	return &HttpUserService{Endpoint: endpoint}
}

func (s *HttpUserService) GetUsers() ([]domain.User, error) {
	/*resp, _ := http.Get(s.Endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	*/
	var users []domain.User
	/*if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, err
	}*/
	//mock de usuarios para simular una respuesta
	users = append(users, domain.User{UID: "1", Strategy: "conservative", ApiKey: "|mockApiKey|"})
	users = append(users, domain.User{UID: "2", Strategy: "aggressive", ApiKey: "|mockApiKey1|"})
	users = append(users, domain.User{UID: "3", Strategy: "conservative", ApiKey: "|mockApiKey2|"})
	users = append(users, domain.User{UID: "4", Strategy: "conservative", ApiKey: "|mockApiKey2|"})
	users = append(users, domain.User{UID: "5", Strategy: "conservative", ApiKey: "|mockApiKey2|"})
	return users, nil
}
