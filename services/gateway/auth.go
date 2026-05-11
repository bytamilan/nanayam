package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// User represents a console user (separate from Fabric identity).
type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	Org          string    `json:"org"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"createdAt"`
}

// AuthConfig holds signup/auth settings from env.
type AuthConfig struct {
	SignupEnabled bool
	JWTSecret     []byte
	SessionHours  int
}

// AuthStore is a thread-safe in-memory user store.
type AuthStore struct {
	mu     sync.RWMutex
	users  map[string]*User // key: username (lowercased)
	config AuthConfig
}

// NewAuthStore creates an auth store from environment variables.
func NewAuthStore() *AuthStore {
	signupEnabled := false
	if v := os.Getenv("AUTH_SIGNUP_ENABLED"); strings.ToLower(v) == "true" || v == "1" {
		signupEnabled = true
	}

	jwtSecret := os.Getenv("AUTH_JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "nanayam-default-secret-change-me"
	}

	sessionHours := 24
	if v := os.Getenv("AUTH_SESSION_HOURS"); v != "" {
		if h, err := strconv.Atoi(v); err == nil && h > 0 {
			sessionHours = h
		}
	}

	return &AuthStore{
		users: make(map[string]*User),
		config: AuthConfig{
			SignupEnabled: signupEnabled,
			JWTSecret:     []byte(jwtSecret),
			SessionHours:  sessionHours,
		},
	}
}

// IsSignupEnabled returns whether registration is allowed.
func (s *AuthStore) IsSignupEnabled() bool {
	return s.config.SignupEnabled
}

// Register creates a new user if signup is enabled and username is available.
func (s *AuthStore) Register(username, password, org, role string) (*User, error) {
	if !s.config.SignupEnabled {
		return nil, fmt.Errorf("registration is disabled")
	}

	if username == "" || password == "" {
		return nil, fmt.Errorf("username and password are required")
	}

	key := strings.ToLower(username)
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[key]; exists {
		return nil, fmt.Errorf("username already exists")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	if role == "" {
		role = "user"
	}

	user := &User{
		ID:           fmt.Sprintf("usr-%d", time.Now().UnixNano()),
		Username:     username,
		PasswordHash: string(hash),
		Org:          org,
		Role:         role,
		CreatedAt:    time.Now().UTC(),
	}
	s.users[key] = user
	return user, nil
}

// Login validates credentials and returns a JWT string.
func (s *AuthStore) Login(username, password string) (string, error) {
	key := strings.ToLower(username)

	s.mu.RLock()
	user, exists := s.users[key]
	s.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("invalid username or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", fmt.Errorf("invalid username or password")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  user.ID,
		"usr":  user.Username,
		"org":  user.Org,
		"role": user.Role,
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(time.Duration(s.config.SessionHours) * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString(s.config.JWTSecret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// ValidateToken parses and validates a JWT, returning the claims.
func (s *AuthStore) ValidateToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.config.JWTSecret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

// GetUserByUsername returns a user by username (read-only copy).
func (s *AuthStore) GetUserByUsername(username string) (*User, bool) {
	key := strings.ToLower(username)
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.users[key]
	if !ok {
		return nil, false
	}
	// Return copy without password hash
	return &User{
		ID:        u.ID,
		Username:  u.Username,
		Org:       u.Org,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
	}, true
}

// SeedAdmin creates a default admin user if no users exist (useful for first-run).
func (s *AuthStore) SeedAdmin() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.users) > 0 {
		return
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
	s.users["admin"] = &User{
		ID:           "usr-admin",
		Username:     "admin",
		PasswordHash: string(hash),
		Org:          "ACBMSP",
		Role:         "admin",
		CreatedAt:    time.Now().UTC(),
	}
}
