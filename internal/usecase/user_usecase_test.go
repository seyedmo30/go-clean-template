package usecase

import (
	"__MODULE__/internal/dto/client/integration"
	"__MODULE__/internal/dto/repository"
	"__MODULE__/internal/entity/user"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// ------------------------- Mocks -------------------------

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreateUser(ctx context.Context, params repository.CreateUserRepositoryRequestDTO) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

func (m *MockRepository) GetUsersList(ctx context.Context, params repository.ListRepositoryRequestDTO[repository.BaseUser]) (repository.ListRepositoryResponseDTO[repository.BaseUser], error) {
	args := m.Called(ctx, mock.Anything)
	if res, ok := args.Get(0).(repository.ListRepositoryResponseDTO[repository.BaseUser]); ok {
		return res, args.Error(1)
	}
	return repository.ListRepositoryResponseDTO[repository.BaseUser]{}, args.Error(1)
}

func (m *MockRepository) GetUserById(ctx context.Context, id string) (repository.BaseUser, error) {
	args := m.Called(ctx, id)
	if r, ok := args.Get(0).(repository.BaseUser); ok {
		return r, args.Error(1)
	}
	return repository.BaseUser{}, args.Error(1)
}

func (m *MockRepository) UpdateUser(ctx context.Context, params repository.UpdateUserRepositoryRequestDTO) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

func (m *MockRepository) DeleteUser(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Mock external user client
type MockUserClient struct {
	mock.Mock
}

func (m *MockUserClient) GetUsers(ctx context.Context, page int) (integration.UserListResponseDTO, error) {
	args := m.Called(ctx, page)
	if r, ok := args.Get(0).(integration.UserListResponseDTO); ok {
		return r, args.Error(1)
	}
	return integration.UserListResponseDTO{}, args.Error(1)
}

// ------------------------- Suite -------------------------

type UserUsecaseSuite struct {
	suite.Suite
	repo   *MockRepository
	client *MockUserClient
	uc     *userUsecase
}

func (s *UserUsecaseSuite) SetupTest() {
	s.repo = &MockRepository{}
	s.client = &MockUserClient{}

	val := NewUserUsecase(s.repo, s.client, 2)
	s.uc = &val
}

func TestUserUsecaseSuite(t *testing.T) {
	suite.Run(t, new(UserUsecaseSuite))
}

// ------------------------- Tests -------------------------

func (s *UserUsecaseSuite) Test_GetUsers_ReturnsFromRepo_WhenExists() {
	// prepare repo result (one user)
	id := user.ID("u-1")
	full := user.FullName("John Doe")
	username := user.Username("jdoe")
	email := user.Email("jdoe@example.com")
	created := time.Now()

	repoUser := repository.BaseUser{
		ID:        &id,
		FullName:  &full,
		Username:  &username,
		Email:     &email,
		CreatedAt: &created,
	}

	listResp := repository.ListRepositoryResponseDTO[repository.BaseUser]{
		List: []repository.BaseUser{repoUser},
		BasePaginationResponse: repository.BasePaginationResponse{
			Limit:   2,
			Total:   1,
			Page:    1,
			HasMore: false,
		},
	}

	// expectations
	s.repo.On("GetUsersList", mock.Anything, mock.Anything).Return(listResp, nil)

	// call
	resp, err := s.uc.GetUsers(context.Background(), 1)
	assert.NoError(s.T(), err)
	assert.Len(s.T(), resp, 1)

	// check pointer fields and values
	if assert.NotNil(s.T(), resp[0].FullName) {
		assert.Equal(s.T(), full, *resp[0].FullName)
	}
	if assert.NotNil(s.T(), resp[0].ID) {
		assert.Equal(s.T(), id, *resp[0].ID)
	}

	// ensure external client wasn't called and CreateUser didn't run
	s.client.AssertNotCalled(s.T(), "GetUsers", mock.Anything, mock.Anything)
	s.repo.AssertNotCalled(s.T(), "CreateUser", mock.Anything, mock.Anything)
	s.repo.AssertExpectations(s.T())
}

func (s *UserUsecaseSuite) Test_GetUsers_CallsClientAndPersists_WhenRepoEmpty() {
	// repo returns empty list
	emptyResp := repository.ListRepositoryResponseDTO[repository.BaseUser]{
		List: []repository.BaseUser{},
		BasePaginationResponse: repository.BasePaginationResponse{
			Limit: 2,
			Total: 0,
			Page:  1,
		},
	}
	s.repo.On("GetUsersList", mock.Anything, mock.Anything).Return(emptyResp, nil)

	// client returns 2 users
	clientUsers := []integration.UserDTO{
		{
			ID:    user.ID("10"),
			Name:  user.FullName("Alice"),
			Email: user.Email("a@x.com"),
			Extra: map[string]string{"company": "C"},
		},
		{
			ID:    user.ID("20"),
			Name:  user.FullName("Bob"),
			Email: user.Email("b@x.com"),
			Extra: map[string]string{"city": "CityX"},
		},
	}
	clientResp := integration.UserListResponseDTO{
		Provider: "jsonplaceholder",
		Users:    clientUsers,
	}
	s.client.On("GetUsers", mock.Anything, 1).Return(clientResp, nil)

	// expect CreateUser called twice
	s.repo.On("CreateUser", mock.Anything, mock.Anything).Return(nil).Times(2)

	// call
	resp, err := s.uc.GetUsers(context.Background(), 1)
	assert.NoError(s.T(), err)
	assert.Len(s.T(), resp, 2)

	// verify CreateUser called twice
	s.repo.AssertNumberOfCalls(s.T(), "CreateUser", 2)
	s.repo.AssertExpectations(s.T())
	s.client.AssertExpectations(s.T())
}

func (s *UserUsecaseSuite) Test_GetUsers_ReturnsError_WhenRepoFails() {
	s.repo.On("GetUsersList", mock.Anything, mock.Anything).Return(repository.ListRepositoryResponseDTO[repository.BaseUser]{}, errors.New("db failure"))

	resp, err := s.uc.GetUsers(context.Background(), 1)
	assert.Error(s.T(), err)
	assert.Empty(s.T(), resp)
	s.repo.AssertExpectations(s.T())
}

func (s *UserUsecaseSuite) Test_GetUsers_ReturnsError_WhenClientFails() {
	// repo empty, client fails
	emptyResp := repository.ListRepositoryResponseDTO[repository.BaseUser]{List: []repository.BaseUser{}}
	s.repo.On("GetUsersList", mock.Anything, mock.Anything).Return(emptyResp, nil)

	s.client.On("GetUsers", mock.Anything, 1).Return(integration.UserListResponseDTO{}, errors.New("client error"))

	resp, err := s.uc.GetUsers(context.Background(), 1)
	assert.Error(s.T(), err)
	assert.Empty(s.T(), resp)

	s.client.AssertExpectations(s.T())
	s.repo.AssertExpectations(s.T())
}

func (s *UserUsecaseSuite) Test_GetUsers_ReturnsError_WhenCreateUserFails() {
	// repo empty
	emptyResp := repository.ListRepositoryResponseDTO[repository.BaseUser]{List: []repository.BaseUser{}}
	s.repo.On("GetUsersList", mock.Anything, mock.Anything).Return(emptyResp, nil)

	// client returns one user
	clientUsers := []integration.UserDTO{
		{
			ID:   user.ID("100"),
			Name: user.FullName("FailUser"),
		},
	}
	clientResp := integration.UserListResponseDTO{
		Provider: "reqres",
		Users:    clientUsers,
	}
	s.client.On("GetUsers", mock.Anything, 1).Return(clientResp, nil)

	// CreateUser fails on first call
	s.repo.On("CreateUser", mock.Anything, mock.Anything).Return(errors.New("insert failed"))

	resp, err := s.uc.GetUsers(context.Background(), 1)
	assert.Error(s.T(), err)
	assert.Empty(s.T(), resp)

	s.repo.AssertExpectations(s.T())
	s.client.AssertExpectations(s.T())
}
