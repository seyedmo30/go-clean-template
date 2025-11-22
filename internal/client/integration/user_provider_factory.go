package integration

import (
	"__MODULE__/internal/config"
	"__MODULE__/internal/interfaces"
	"__MODULE__/pkg"
)

type UserServiceFactory func(config.App, string) (interfaces.UserService, error)

var userRegistry = make(map[string]UserServiceFactory)

func RegisterUserServiceFactory(providerName string, factory UserServiceFactory) {
	userRegistry[providerName] = factory
}

// Provider descriptor (same shape as your other provider types)
type Provider struct {
	Name string
}

type userProviderService struct {
	config         config.App
	UserServiceMap map[string]interfaces.UserService
	Providers      []Provider
}

var _ interfaces.ProviderService = (*userProviderService)(nil) // optional assert if your ProviderService interface aligns

// NewUserProviderService returns a manager for user providers.
func NewUserProviderService(cfg config.App) *userProviderService {
	return &userProviderService{
		config:         cfg,
		UserServiceMap: make(map[string]interfaces.UserService),
		Providers:      []Provider{},
	}
}

// RegisterNewProvider creates & registers a new user provider instance
// id - unique id in your system, providerName - "reqres" or "jsonplaceholder", providerConfig optional.
func (u *userProviderService) RegisterNewProvider(id string, providerName string, providerConfig string) error {
	factory, ok := userRegistry[providerName]
	if !ok {
		// return pkg.ErrProviderRegistration.AddDescription(fmt.Sprintf("unrecognized user provider: %s", providerName))
		return pkg.NewAppError(pkg.ErrBadRequest)
	}
	svc, err := factory(u.config, providerConfig)
	if err != nil {
		return err
	}
	u.UserServiceMap[id] = svc
	u.Providers = append(u.Providers, Provider{Name: providerName})
	return nil
}

func (u *userProviderService) GetUserService(id string) (interfaces.UserService, error) {
	svc, ok := u.UserServiceMap[id]
	if !ok {
		// return nil, pkg.ErrBankServiceNotActive.AddStack() // reuse existing error or define a new one
		return nil, pkg.NewAppError(pkg.ErrBadRequest)
	}
	return svc, nil
}

func (u *userProviderService) StopUserService(id string) {
	delete(u.UserServiceMap, id)
}
