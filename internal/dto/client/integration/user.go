package integration

// Unified DTOs used by the application.

type UserDTO struct {
	ID       string            `json:"id"`
	Name     string            `json:"name,omitempty"`
	Username string            `json:"username,omitempty"`
	Email    string            `json:"email,omitempty"`
	Phone    string            `json:"phone,omitempty"`
	Website  string            `json:"website,omitempty"`
	Extra    map[string]string `json:"extra,omitempty"` // provider-specific fields
}

type MetaInfoDTO struct {
	Page       int `json:"page,omitempty"`
	PerPage    int `json:"per_page,omitempty"`
	Total      int `json:"total,omitempty"`
	TotalPages int `json:"total_pages,omitempty"`
}

type UserListResponseDTO struct {
	Provider string      `json:"provider"`
	Users    []UserDTO   `json:"users"`
	Meta     MetaInfoDTO `json:"meta,omitempty"`
	Raw      []byte      `json:"-"`
}
