package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"__MODULE__/internal/config"
	"__MODULE__/internal/dto/client/integration"
	"__MODULE__/internal/entity/user"
	"__MODULE__/internal/interfaces"
	"__MODULE__/pkg"
)

const JsonPlaceholderProvider = "jsonplaceholder"

type jsonPlaceholderService struct {
	baseURL    string
	httpClient *http.Client
	provider   string
}

func init() {
	RegisterUserServiceFactory(JsonPlaceholderProvider, func(cfg config.App, _ string) (interfaces.UserService, error) {
		return &jsonPlaceholderService{
			baseURL:    "https://jsonplaceholder.typicode.com",
			httpClient: &http.Client{},
			provider:   JsonPlaceholderProvider,
		}, nil
	})
}

func (j *jsonPlaceholderService) GetUsers(ctx context.Context, page int) (res integration.UserListResponseDTO, err error) {
	u := fmt.Sprintf("%s/users", j.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return integration.UserListResponseDTO{}, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := j.httpClient.Do(req)
	if err != nil {
		return integration.UserListResponseDTO{}, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		str := "status code response is not 200 ."
		str = str + " resp.StatusCode : " + strconv.Itoa(resp.StatusCode)
		str = str + string(body)
		err = pkg.NewAppError(pkg.ErrBadRequest).AddDescription([]byte(str)).AppendStackLog()
		return res, err
	}

	var parsed []jpUser
	if err := json.Unmarshal(body, &parsed); err != nil {
		return integration.UserListResponseDTO{}, err
	}

	users := make([]integration.UserDTO, 0, len(parsed))
	for _, u := range parsed {
		users = append(users, integration.UserDTO{
			ID:       user.ID(strconv.Itoa(u.ID)),
			Name:     user.FullName(u.Name),
			Username: user.Username(u.Username),
			Email:    user.Email(u.Email),
			Phone:    user.Phone(u.Phone),
			Website:  user.Website(u.Website),
			Extra: map[string]string{
				"company": u.Company.Name,
				"city":    u.Address.City,
			},
		})
	}

	// jsonplaceholder has no pagination; meta left empty
	return integration.UserListResponseDTO{
		Provider: j.provider,
		Users:    users,
		Raw:      body,
	}, nil
}
