package repository

type BasePaginationRequest struct {
	Limit int
	Page  int
	Sort  string
}

type BasePaginationResponse struct {
	Limit   int
	Total   int64
	Page    int
	HasMore bool
}

type ListRepositoryResponseDTO[T any] struct {
	List []T
	BasePaginationResponse
}

type ListRepositoryRequestDTO[T any] struct {
	Filter T
	BasePaginationRequest
}
