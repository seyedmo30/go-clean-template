package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"__MODULE__/internal/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// Run real external integration tests only when SKIP_REAL_EXTERNAL_TESTS != "1".
// This is useful to avoid flaky CI when external services are unreachable.

type UserProviderRealSuite struct {
	suite.Suite
	cfg config.App
	svc *userProviderService
}

func (s *UserProviderRealSuite) SetupSuite() {
	s.cfg = config.App{}
	s.svc = NewUserProviderService(s.cfg)
}

func (s *UserProviderRealSuite) shouldSkip() bool {
	return os.Getenv("SKIP_REAL_EXTERNAL_TESTS") == "1"
}

func (s *UserProviderRealSuite) TestRegisterUnknownProvider() {
	err := s.svc.RegisterNewProvider("p1", "unknown-provider", "")
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "unrecognized user provider")
}

func (s *UserProviderRealSuite) TestRegisterReqresProviderExists() {
	err := s.svc.RegisterNewProvider("p1", ReqresProvider, "")
	assert.NoError(s.T(), err)

	p, err := s.svc.GetUserService("p1")
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), p)
}

func (s *UserProviderRealSuite) TestStopUserService() {
	_ = s.svc.RegisterNewProvider("stop-test", ReqresProvider, "")

	s.svc.StopUserService("stop-test")

	_, err := s.svc.GetUserService("stop-test")
	assert.Error(s.T(), err)
}

// ----- Real external calls -----

func (s *UserProviderRealSuite) TestReqresProvider_GetUsers_Real() {
	if s.shouldSkip() {
		s.T().Skip("skipping real external tests (SKIP_REAL_EXTERNAL_TESTS=1)")
	}

	// register provider (uses default baseURL in your provider)
	err := s.svc.RegisterNewProvider("req-real", ReqresProvider, "")
	assert.NoError(s.T(), err)

	p, err := s.svc.GetUserService("req-real")
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), p)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	resp, err := p.GetUsers(ctx, 1)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), ReqresProvider, resp.Provider)
	assert.Greater(s.T(), len(resp.Users), 0, "expected at least one user from reqres")
}

func (s *UserProviderRealSuite) TestJsonPlaceholder_GetUsers_Real() {
	if s.shouldSkip() {
		s.T().Skip("skipping real external tests (SKIP_REAL_EXTERNAL_TESTS=1)")
	}

	err := s.svc.RegisterNewProvider("jp-real", JsonPlaceholderProvider, "")
	assert.NoError(s.T(), err)

	p, err := s.svc.GetUserService("jp-real")
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), p)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	resp, err := p.GetUsers(ctx, 0)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), JsonPlaceholderProvider, resp.Provider)
	assert.Greater(s.T(), len(resp.Users), 0, "expected at least one user from jsonplaceholder")
}

func TestUserProviderRealSuite(t *testing.T) {
	suite.Run(t, new(UserProviderRealSuite))
}
