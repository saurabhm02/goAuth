package impl

import (
	"context"
	"errors"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"goAuth/internals/helpers"
	"goAuth/internals/logger"
	"goAuth/internals/model"
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
	repo       types.UserRepository
	otpRepo    types.OTPRepository
	otpEnabled bool
}

func NewAuthService(repo types.UserRepository, otpRepo types.OTPRepository, otpEnabled bool) *AuthServiceImpl {
	return &AuthServiceImpl{repo: repo, otpRepo: otpRepo, otpEnabled: otpEnabled}
}

func (s *AuthServiceImpl) Signup(ctx context.Context, req *model.SignupRequest) (*model.User, error) {
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
	if err != types.ErrNotFound {
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
	if s.otpEnabled && s.otpRepo != nil {
		log.Printf("[signup-service] OTP enabled, starting transaction for email %q", email)
		transactor, ok := s.repo.(types.Transactor)
		if !ok {
			return nil, errors.New("repository does not support transactions")
		}
		tx, err := transactor.BeginTx()
		if err != nil {
			log.Printf("[signup-service] failed to begin transaction: %v", err)
			return nil, err
		}
		committed := false
		defer func() {
			if !committed {
				log.Printf("[signup-service] rolling back transaction for email %q", email)
				_ = tx.Rollback()
			}
		}()

		if err := s.repo.CreateTx(tx, user); err != nil {
			if err == types.ErrConflict {
				log.Printf("[signup-service] conflict inserting user %q", email)
				return nil, types.ErrConflict
			}
			log.Printf("[signup-service] error inserting user: %v", err)
			return nil, err
		}
		log.Printf("[signup-service] user %s inserted, generating OTP", user.ID)

		code, otpHash, err := helpers.GenerateOTP()
		if err != nil {
			log.Printf("[signup-service] error generating OTP: %v", err)
			return nil, err
		}
		expiry := helpers.OTPExpiryTime()
		if err := s.otpRepo.CreateTx(tx, user.ID, otpHash, expiry); err != nil {
			log.Printf("[signup-service] error inserting OTP: %v", err)
			return nil, err
		}
		log.Printf("[signup-service] OTP inserted, sending email to %q", email)

		if err := helpers.SendOTPEmail(user.Email, code); err != nil {
			log.Printf("[signup-service] error sending OTP email to %q: %v", email, err)
			return nil, err
		}
		log.Printf("[signup-service] email sent successfully to %q. Committing transaction...", email)

		if err := tx.Commit(); err != nil {
			log.Printf("[signup-service] failed to commit transaction: %v", err)
			return nil, err
		}
		committed = true
		log.Printf("[signup-service] transaction committed successfully for %q", email)
		return user, nil
	}

	log.Printf("[signup-service] OTP disabled, performing standard user insert for %q", email)
	if err := s.repo.Create(user); err != nil {
		if err == types.ErrConflict {
			return nil, types.ErrConflict
		}
		return nil, err
	}
	return user, nil
}

func (s *AuthServiceImpl) Login(ctx context.Context, emailOrPhone, password string, useOTP bool) (*model.User, *model.TokenResponse, error) {
	emailOrPhone = strings.TrimSpace(emailOrPhone)
	identifier := emailOrPhone
	if identifier == "" {
		identifier = "-"
	}
	logger.Logf(ctx, "login", identifier, "processing login request")

	if emailOrPhone == "" {
		return nil, nil, types.ErrUnauthorized
	}
	if strings.Contains(emailOrPhone, "@") {
		emailOrPhone = strings.ToLower(emailOrPhone)
	}
	user, err := s.repo.GetByEmailOrPhone(emailOrPhone)
	if err != nil {
		if err == types.ErrNotFound {
			logger.Logf(ctx, "login", identifier, "user not found")
			return nil, nil, types.ErrUnauthorized
		}
		logger.Logf(ctx, "login", identifier, "db error fetching user: %v", err)
		return nil, nil, err
	}
	if useOTP && s.otpEnabled && s.otpRepo != nil {
		logger.Logf(ctx, "login", identifier, "OTP login requested, generating OTP")
		code, hash, err := helpers.GenerateOTP()
		if err != nil {
			logger.Logf(ctx, "login", identifier, "failed to generate OTP: %v", err)
			return nil, nil, err
		}
		expiry := helpers.OTPExpiryTime()
		if err := s.otpRepo.Create(user.ID, hash, expiry); err != nil {
			logger.Logf(ctx, "login", identifier, "failed to store OTP: %v", err)
			return nil, nil, err
		}
		if err := helpers.SendOTPEmail(user.Email, code); err != nil {
			logger.Logf(ctx, "login", identifier, "failed to send OTP email: %v", err)
			return nil, nil, err
		}
		logger.Logf(ctx, "login", identifier, "OTP sent successfully")
		return nil, nil, types.ErrOTPSent
	}
	hasher, jwt := getEnvHasherAndJWT()
	if !hasher.Verify(password, user.PasswordHash) {
		logger.Logf(ctx, "login", identifier, "invalid password")
		return nil, nil, types.ErrInvalidPassword
	}
	token, expiresIn, err := jwt.Issue(user.ID)
	if err != nil {
		logger.Logf(ctx, "login", identifier, "failed to issue token: %v", err)
		return nil, nil, err
	}
	logger.Logf(ctx, "login", identifier, "login successful, token issued")
	return user, &model.TokenResponse{
		Token:     token,
		ExpiresIn: expiresIn,
	}, nil
}

func (s *AuthServiceImpl) VerifyOTP(ctx context.Context, emailOrPhone, otp string) (*model.User, *model.TokenResponse, error) {
	emailOrPhone = strings.TrimSpace(emailOrPhone)
	identifier := emailOrPhone
	if identifier == "" {
		identifier = "-"
	}
	logger.Logf(ctx, "verify_otp", identifier, "verifying OTP")

	if emailOrPhone == "" || otp == "" {
		return nil, nil, types.ErrInvalidOTP
	}
	if strings.Contains(emailOrPhone, "@") {
		emailOrPhone = strings.ToLower(emailOrPhone)
	}
	user, err := s.repo.GetByEmailOrPhone(emailOrPhone)
	if err != nil {
		if err == types.ErrNotFound {
			logger.Logf(ctx, "verify_otp", identifier, "user not found")
			return nil, nil, types.ErrInvalidOTP
		}
		logger.Logf(ctx, "verify_otp", identifier, "db error fetching user: %v", err)
		return nil, nil, err
	}
	if s.otpRepo == nil {
		return nil, nil, types.ErrInvalidOTP
	}
	otpHash, expiry, err := s.otpRepo.GetByUserID(user.ID)
	if err != nil {
		if errors.Is(err, types.ErrNotFound) {
			logger.Logf(ctx, "verify_otp", identifier, "no OTP found for user")
			return nil, nil, types.ErrInvalidOTP
		}
		logger.Logf(ctx, "verify_otp", identifier, "db error fetching OTP: %v", err)
		return nil, nil, err
	}
	if time.Now().After(expiry) {
		logger.Logf(ctx, "verify_otp", identifier, "OTP expired")
		return nil, nil, types.ErrInvalidOTP
	}
	if !helpers.VerifyOTP(otp, otpHash) {
		logger.Logf(ctx, "verify_otp", identifier, "invalid OTP provided")
		return nil, nil, types.ErrInvalidOTP
	}
	_, jwt := getEnvHasherAndJWT()
	token, expiresIn, err := jwt.Issue(user.ID)
	if err != nil {
		logger.Logf(ctx, "verify_otp", identifier, "failed to issue token: %v", err)
		return nil, nil, err
	}
	logger.Logf(ctx, "verify_otp", identifier, "OTP verified successfully, token issued")
	return user, &model.TokenResponse{Token: token, ExpiresIn: expiresIn}, nil
}

func (s *AuthServiceImpl) GetByID(ctx context.Context, userID string) (*model.User, error) {
	logger.Logf(ctx, "get_user", userID, "fetching user profile")
	user, err := s.repo.GetByID(userID)
	if err != nil {
		if err == types.ErrNotFound {
			logger.Logf(ctx, "get_user", userID, "user not found")
			return nil, types.ErrNotFound
		}
		logger.Logf(ctx, "get_user", userID, "db error fetching user: %v", err)
		return nil, err
	}
	logger.Logf(ctx, "get_user", userID, "user successfully retrieved")
	return user, nil
}

var _ service.AuthService = (*AuthServiceImpl)(nil)
