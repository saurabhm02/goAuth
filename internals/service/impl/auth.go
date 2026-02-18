package impl

import (
	"os"
	"regexp"
	"strings"
	"sync"

	"goAuth/internals/helpers"
	"goAuth/internals/model"
	"goAuth/internals/repository"
	"goAuth/internals/service"
	"goAuth/internals/types"
)

const minPasswordLen = 8

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

var (
	envHasher types.Hasher
	envJWT    helpers.TokenProvider
	envOnce   sync.Once
)

func getEnvHasherAndJWT() (types.Hasher, helpers.TokenProvider) {
	envOnce.Do(func() {
		envHasher = helpers.NewBcryptHasher()
		secret := os.Getenv(types.EnvJWTSecret)
		if secret == "" {
			secret = "change-me-in-production"
		}
		envJWT = helpers.NewJWTProvider(secret, types.TokenExpiry)
	})
	return envHasher, envJWT
}

type AuthServiceImpl struct {
	repo repository.UserRepository
}

func NewAuthService(repo repository.UserRepository) *AuthServiceImpl {
	return &AuthServiceImpl{repo: repo}
}

func (s *AuthServiceImpl) Signup(req *model.SignupRequest) (*model.User, error) {
	email := strings.TrimSpace(strings.ToLower(req.Email))
	phone := strings.TrimSpace(req.Phone)
	if email == "" {
		return nil, types.ErrValidation
	}
	if !emailRegex.MatchString(email) {
		return nil, types.ErrValidation
	}
	if len(req.Password) < minPasswordLen {
		return nil, types.ErrValidation
	}
	_, err := s.repo.GetByEmail(email)
	if err == nil {
		return nil, types.ErrConflict
	}
	if err != repository.ErrNotFound {
		return nil, err
	}
	hasher, _ := getEnvHasherAndJWT()
	hash, err := hasher.Hash(req.Password)
	if err != nil {
		return nil, err
	}
	user := &model.User{
		Email:        email,
		Phone:        phone,
		PasswordHash: hash,
	}
	if err := s.repo.Create(user); err != nil {
		if err == repository.ErrConflict {
			return nil, types.ErrConflict
		}
		return nil, err
	}
	return user, nil
}

func (s *AuthServiceImpl) Login(emailOrPhone, password string) (*model.User, *model.TokenResponse, error) {
	emailOrPhone = strings.TrimSpace(emailOrPhone)
	if emailOrPhone == "" {
		return nil, nil, types.ErrUnauthorized
	}
	if strings.Contains(emailOrPhone, "@") {
		emailOrPhone = strings.ToLower(emailOrPhone)
	}
	user, err := s.repo.GetByEmailOrPhone(emailOrPhone)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, nil, types.ErrUnauthorized
		}
		return nil, nil, err
	}
	hasher, jwt := getEnvHasherAndJWT()
	if !hasher.Verify(password, user.PasswordHash) {
		return nil, nil, types.ErrInvalidPassword
	}
	token, expiresIn, err := jwt.Issue(user.ID)
	if err != nil {
		return nil, nil, err
	}
	return user, &model.TokenResponse{
		Token:     token,
		ExpiresIn: expiresIn,
	}, nil
}

func (s *AuthServiceImpl) GetByID(userID string) (*model.User, error) {
	user, err := s.repo.GetByID(userID)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, types.ErrNotFound
		}
		return nil, err
	}
	return user, nil
}

var _ service.AuthService = (*AuthServiceImpl)(nil)
