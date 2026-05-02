package auth

import (
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/example/coworking/internal/repository"
)

// Service encapsulates JWT issuance for students (Telegram Mini App) and admins.
type Service struct {
	repo        *repository.Repo
	jwtSecret   []byte
	jwtTTL      time.Duration
	botTok      string
	initDataTTL time.Duration
}

func NewService(repo *repository.Repo, jwtSecret []byte, jwtTTL time.Duration, botToken string) *Service {
	return &Service{
		repo:        repo,
		jwtSecret:   jwtSecret,
		jwtTTL:      jwtTTL,
		botTok:      strings.TrimSpace(botToken),
		initDataTTL: 24 * time.Hour,
	}
}

func (s *Service) AdminLogin(email, password string) (string, error) {
	if strings.TrimSpace(email) == "" || password == "" {
		return "", ErrUnauthorized
	}
	if s.repo == nil {
		return "", ErrUnauthorized
	}
	u, err := s.repo.GetAdminByEmail(strings.TrimSpace(email))
	if err != nil {
		return "", ErrUnauthorized
	}
	if u.PasswordHash == "" {
		return "", ErrUnauthorized
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return "", ErrUnauthorized
	}
	return s.signToken(u.ID, u.Role)
}

func (s *Service) UpsertStudentFromTelegramInitData(initDataRaw string) (string, error) {
	if strings.TrimSpace(s.botTok) == "" {
		return "", ErrTelegramNotConfigured
	}
	if s.repo == nil {
		return "", ErrUnauthorized
	}

	tgUID, err := ValidateTelegramInitData(strings.TrimSpace(initDataRaw), s.botTok, s.initDataTTL)
	if err != nil {
		return "", err
	}

	u, err := s.repo.UpsertStudentTelegramUser(tgUID)
	if err != nil {
		return "", err
	}
	return s.signToken(u.ID, u.Role)
}
