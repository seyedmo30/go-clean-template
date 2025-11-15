package repository

import (
	"__MODULE__/internal/config"
	"__MODULE__/internal/dto/repository"
	"__MODULE__/internal/entity/user"
	"context"
	"fmt"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// RepositorySuite is the testify suite for repository tests.
type RepositorySuite struct {
	suite.Suite
	db  *gorm.DB
	ctx context.Context
	r   *serviceRepository
}

// SetupSuite runs once before the suite.
func (s *RepositorySuite) SetupSuite() {
	s.T().Helper()

	// For local testing, connect to a PostgreSQL test database.
	// You can use a dedicated "test" DB to safely run migrations and truncations.
	// Example DSN for local setup:
	//   postgres://user:password@localhost:5432/testdb?sslmode=disable
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		"localhost", "user", "password", "template_clean", "5432",
	)

	gdb, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(s.T(), err, "open PostgreSQL connection")

	// assign to package-level db used by repository methods
	db = gdb
	s.db = gdb

	// Auto-migrate test models (ensures required tables exist)
	require.NoError(s.T(), db.AutoMigrate(&repository.BaseUser{}), "auto migrate BaseUser")

	s.ctx = context.Background()
	s.r = &serviceRepository{}
}

// TearDownSuite runs after all tests in the suite.
func (s *RepositorySuite) TearDownSuite() {
	s.T().Helper()
	if s.db != nil {
		sqlDB, err := s.db.DB()
		if err == nil && sqlDB != nil {
			_ = sqlDB.Close()
		}
	}
}

// BeforeTest ensures a clean table for each test
func (s *RepositorySuite) BeforeTest(_, _ string) {
	// Truncate users table for clean state before each test
	err := s.db.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE").Error
	require.NoError(s.T(), err, "truncate users table")
}

// TestConstructor_basic shows a minimal constructor-like behavior test.
// NOTE: We avoid calling NewServiceRepository (because it calls SetupDB).
func (s *RepositorySuite) TestConstructor_basic() {
	r := &serviceRepository{
		config: config.App{
			DatabaseConfig: config.DatabaseConfig{
				Database: "template_clean",
				Username: "template_clean",
				Password: "template_clean",
				Host:     "localhost",
				Port:     "5432",
			},
		},
	}
	require.NotNil(s.T(), r)
}

// TestCRUD_flow performs full Create -> List -> Get -> Update -> Delete flow.
func (s *RepositorySuite) TestCRUD_flow() {
	ctx := s.ctx
	r := s.r

	// prepare entity values (user.* are type aliases in your codebase)
	id := user.ID("u1")
	full := user.FullName("John Doe")
	username := user.Username("jdoe")
	email := user.Email("jdoe@example.com")
	avatar := user.Avatar("/avatar.png")
	phone := user.Phone("+1000000000")
	website := user.Website("https://example.com")
	company := user.Company("Example Inc")
	city := user.City("Kyiv")
	isActive := true
	now := time.Now().UTC()

	// 1) Create user
	createParams := repository.CreateUserRepositoryRequestDTO{
		BaseUser: repository.BaseUser{
			ID:        &id,
			FullName:  &full,
			Username:  &username,
			Email:     &email,
			Avatar:    &avatar,
			Phone:     &phone,
			Website:   &website,
			Company:   &company,
			City:      &city,
			IsActive:  &isActive,
			CreatedAt: &now,
			UpdatedAt: &now,
		},
	}

	require.NoError(s.T(), r.CreateUser(ctx, createParams), "CreateUser failed")

	// 2) List users
	listParams := repository.ListRepositoryRequestDTO[repository.BaseUser]{BasePaginationRequest: repository.BasePaginationRequest{Limit: 10, Page: 1}}
	res, err := r.GetUsersList(ctx, listParams)
	require.NoError(s.T(), err)
	require.Len(s.T(), res.List, 1)

	// 3) Get user by id
	got, err := r.GetUserById(ctx, string(*createParams.ID))
	require.NoError(s.T(), err)
	require.NotNil(s.T(), got.Username)
	require.Equal(s.T(), string(*createParams.Username), string(*got.Username))

	// 4) Update user (change full name)
	newFull := user.FullName("Jane Doe")
	updateParams := repository.UpdateUserRepositoryRequestDTO{
		BaseUser: repository.BaseUser{
			ID:       createParams.ID,
			FullName: &newFull,
		},
	}
	require.NoError(s.T(), r.UpdateUser(ctx, updateParams))

	// verify update
	got2, err := r.GetUserById(ctx, string(*createParams.ID))
	require.NoError(s.T(), err)
	require.NotNil(s.T(), got2.FullName)
	require.Equal(s.T(), string(newFull), string(*got2.FullName))

	// 5) Delete user
	require.NoError(s.T(), r.DeleteUser(ctx, string(*createParams.ID)))

	// 6) Get after delete should fail
	_, err = r.GetUserById(ctx, string(*createParams.ID))
	require.Error(s.T(), err)
}

// Run the suite
func TestRepositorySuite(t *testing.T) {
	suite.Run(t, new(RepositorySuite))
}
