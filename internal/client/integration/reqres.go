package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"__MODULE__/internal/config"
	"__MODULE__/internal/dto/client/integration"
	"__MODULE__/internal/interfaces"
)

const ReqresProvider = "reqres"

type reqresService struct {
	baseURL    string
	httpClient *http.Client
	provider   string
}

func init() {
	RegisterUserServiceFactory(ReqresProvider, func(cfg config.App, _ string) (interfaces.UserService, error) {
		svc := &reqresService{
			baseURL:    "https://reqres.in",
			httpClient: &http.Client{},
			provider:   ReqresProvider,
		}
		return svc, nil
	})
}

func (r *reqresService) GetUsers(ctx context.Context, page int) (integration.UserListResponseDTO, error) {
	u := fmt.Sprintf("%s/api/users", r.baseURL)
	values := url.Values{}
	if page <= 0 {
		page = 1
	}
	values.Set("page", strconv.Itoa(page))
	reqURL := u + "?" + values.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return integration.UserListResponseDTO{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	// Example of custom header (as in your curl)
	req.Header.Set("x-api-key", "reqres-free-v1")

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return integration.UserListResponseDTO{}, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return integration.UserListResponseDTO{Provider: r.provider, Raw: body}, fmt.Errorf("reqres: status %d", resp.StatusCode)
	}

	var parsed reqresListResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return integration.UserListResponseDTO{}, err
	}

	users := make([]integration.UserDTO, 0, len(parsed.Data))
	for _, u := range parsed.Data {
		users = append(users, integration.UserDTO{
			ID:    strconv.Itoa(u.ID),
			Name:  u.FirstName + " " + u.LastName,
			Email: u.Email,
			Extra: map[string]string{
				"avatar": u.Avatar,
			},
		})
	}

	return integration.UserListResponseDTO{
		Provider: r.provider,
		Users:    users,
		Meta: integration.MetaInfoDTO{
			Page:       parsed.Page,
			PerPage:    parsed.PerPage,
			Total:      parsed.Total,
			TotalPages: parsed.TotalPages,
		},
		Raw: body,
	}, nil
}
