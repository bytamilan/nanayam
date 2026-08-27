package main

import (
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func enabledStore(t *testing.T) *AuthStore {
	t.Helper()
	t.Setenv("AUTH_SIGNUP_ENABLED", "true")
	return NewAuthStore()
}

func TestNewAuthStoreDefaults(t *testing.T) {
	t.Setenv("AUTH_SIGNUP_ENABLED", "")
	t.Setenv("AUTH_JWT_SECRET", "")
	t.Setenv("AUTH_SESSION_HOURS", "")

	store := NewAuthStore()

	// Signup must default to closed: an open one would let anyone mint a
	// console account against a running ledger.
	if store.IsSignupEnabled() {
		t.Error("signup is enabled by default, want disabled")
	}
	if store.config.SessionHours != 24 {
		t.Errorf("SessionHours = %d, want 24", store.config.SessionHours)
	}
	if len(store.config.JWTSecret) == 0 {
		t.Error("JWTSecret is empty; tokens could be signed with no key")
	}
}

func TestSignupEnabledAcceptsTrueAndOne(t *testing.T) {
	for _, value := range []string{"true", "TRUE", "True", "1"} {
		t.Run(value, func(t *testing.T) {
			t.Setenv("AUTH_SIGNUP_ENABLED", value)
			if !NewAuthStore().IsSignupEnabled() {
				t.Errorf("AUTH_SIGNUP_ENABLED=%q did not enable signup", value)
			}
		})
	}

	for _, value := range []string{"false", "0", "yes", "", "no"} {
		t.Run("disabled/"+value, func(t *testing.T) {
			t.Setenv("AUTH_SIGNUP_ENABLED", value)
			if NewAuthStore().IsSignupEnabled() {
				t.Errorf("AUTH_SIGNUP_ENABLED=%q unexpectedly enabled signup", value)
			}
		})
	}
}

func TestSessionHoursOverride(t *testing.T) {
	t.Setenv("AUTH_SESSION_HOURS", "8")
	if got := NewAuthStore().config.SessionHours; got != 8 {
		t.Errorf("SessionHours = %d, want 8", got)
	}

	// A nonsensical value must not shorten sessions to zero or go negative.
	for _, value := range []string{"0", "-5", "abc", ""} {
		t.Setenv("AUTH_SESSION_HOURS", value)
		if got := NewAuthStore().config.SessionHours; got != 24 {
			t.Errorf("AUTH_SESSION_HOURS=%q gave SessionHours = %d, want the 24h default", value, got)
		}
	}
}

func TestRegisterRequiresSignupEnabled(t *testing.T) {
	t.Setenv("AUTH_SIGNUP_ENABLED", "false")
	store := NewAuthStore()

	if _, err := store.Register("alice", "pw", "Org1MSP", "user"); err == nil {
		t.Fatal("Register() = nil with signup disabled, want an error")
	}
}

func TestRegisterRejectsEmptyCredentials(t *testing.T) {
	store := enabledStore(t)

	if _, err := store.Register("", "pw", "Org1MSP", "user"); err == nil {
		t.Error("Register() accepted an empty username")
	}
	if _, err := store.Register("alice", "", "Org1MSP", "user"); err == nil {
		t.Error("Register() accepted an empty password")
	}
}

func TestRegisterRejectsDuplicateUsernames(t *testing.T) {
	store := enabledStore(t)

	if _, err := store.Register("alice", "pw", "Org1MSP", "user"); err != nil {
		t.Fatalf("first Register() = %v", err)
	}
	if _, err := store.Register("alice", "other-pw", "Org1MSP", "user"); err == nil {
		t.Fatal("Register() allowed a duplicate username")
	}

	// Usernames are matched case-insensitively, so Alice cannot be taken twice.
	if _, err := store.Register("ALICE", "other-pw", "Org1MSP", "user"); err == nil {
		t.Fatal("Register() allowed a duplicate username differing only in case")
	}
}

func TestRegisterHashesThePassword(t *testing.T) {
	store := enabledStore(t)

	user, err := store.Register("alice", "super-secret", "Org1MSP", "user")
	if err != nil {
		t.Fatalf("Register() = %v", err)
	}

	if user.PasswordHash == "super-secret" {
		t.Fatal("the password was stored in plain text")
	}
	if !strings.HasPrefix(user.PasswordHash, "$2") {
		t.Errorf("PasswordHash = %q, want a bcrypt hash", user.PasswordHash)
	}
}

func TestRegisterDefaultsRoleToUser(t *testing.T) {
	store := enabledStore(t)

	user, err := store.Register("alice", "pw", "Org1MSP", "")
	if err != nil {
		t.Fatalf("Register() = %v", err)
	}
	if user.Role != "user" {
		t.Errorf("Role = %q, want user", user.Role)
	}
}

func TestLoginIsCaseInsensitiveOnUsername(t *testing.T) {
	store := enabledStore(t)
	if _, err := store.Register("Alice", "pw", "Org1MSP", "user"); err != nil {
		t.Fatalf("Register() = %v", err)
	}

	for _, username := range []string{"Alice", "alice", "ALICE"} {
		if _, err := store.Login(username, "pw"); err != nil {
			t.Errorf("Login(%q) = %v", username, err)
		}
	}
}

// The error must not reveal whether the username exists, or it becomes a user
// enumeration oracle.
func TestLoginFailuresAreIndistinguishable(t *testing.T) {
	store := enabledStore(t)
	if _, err := store.Register("alice", "pw", "Org1MSP", "user"); err != nil {
		t.Fatalf("Register() = %v", err)
	}

	_, wrongPassword := store.Login("alice", "wrong")
	_, noSuchUser := store.Login("nobody", "wrong")

	if wrongPassword == nil || noSuchUser == nil {
		t.Fatal("both logins should fail")
	}
	if wrongPassword.Error() != noSuchUser.Error() {
		t.Errorf("login errors differ and leak account existence: %q vs %q", wrongPassword, noSuchUser)
	}
}

func TestLoginIssuesATokenCarryingTheUserClaims(t *testing.T) {
	store := enabledStore(t)
	if _, err := store.Register("alice", "pw", "Org1MSP", "auditor"); err != nil {
		t.Fatalf("Register() = %v", err)
	}

	token, err := store.Login("alice", "pw")
	if err != nil {
		t.Fatalf("Login() = %v", err)
	}

	claims, err := store.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken() = %v", err)
	}
	if claims["usr"] != "alice" {
		t.Errorf("usr claim = %v, want alice", claims["usr"])
	}
	if claims["org"] != "Org1MSP" {
		t.Errorf("org claim = %v, want Org1MSP", claims["org"])
	}
	if claims["role"] != "auditor" {
		t.Errorf("role claim = %v, want auditor", claims["role"])
	}
	if claims["exp"] == nil {
		t.Error("token has no expiry claim")
	}
}

func TestValidateTokenRejectsGarbage(t *testing.T) {
	store := NewAuthStore()

	for _, token := range []string{"", "not-a-jwt", "a.b.c", strings.Repeat("x", 100)} {
		if _, err := store.ValidateToken(token); err == nil {
			t.Errorf("ValidateToken(%q) = nil, want an error", token)
		}
	}
}

func TestValidateTokenRejectsAForeignSecret(t *testing.T) {
	t.Setenv("AUTH_JWT_SECRET", "secret-a")
	storeA := NewAuthStore()
	storeA.SeedAdmin()
	token, err := storeA.Login("admin", "admin")
	if err != nil {
		t.Fatalf("Login() = %v", err)
	}

	t.Setenv("AUTH_JWT_SECRET", "secret-b")
	storeB := NewAuthStore()

	if _, err := storeB.ValidateToken(token); err == nil {
		t.Fatal("a token signed with another secret was accepted")
	}
}

// The "none" algorithm attack: a token with no signature must be rejected.
func TestValidateTokenRejectsUnsignedTokens(t *testing.T) {
	store := NewAuthStore()

	unsigned := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{
		"usr": "admin",
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	token, err := unsigned.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("sign unsigned token: %v", err)
	}

	if _, err := store.ValidateToken(token); err == nil {
		t.Fatal("an alg=none token was accepted")
	}
}

func TestValidateTokenRejectsExpiredTokens(t *testing.T) {
	store := NewAuthStore()

	expired := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"usr": "admin",
		"exp": time.Now().Add(-time.Hour).Unix(),
	})
	token, err := expired.SignedString(store.config.JWTSecret)
	if err != nil {
		t.Fatalf("sign expired token: %v", err)
	}

	if _, err := store.ValidateToken(token); err == nil {
		t.Fatal("an expired token was accepted")
	}
}

func TestGetUserByUsernameNeverReturnsTheHash(t *testing.T) {
	store := enabledStore(t)
	if _, err := store.Register("alice", "pw", "Org1MSP", "user"); err != nil {
		t.Fatalf("Register() = %v", err)
	}

	user, ok := store.GetUserByUsername("ALICE")
	if !ok {
		t.Fatal("GetUserByUsername() did not find the user case-insensitively")
	}
	if user.PasswordHash != "" {
		t.Errorf("GetUserByUsername() returned the password hash %q", user.PasswordHash)
	}
	if user.Username != "alice" {
		t.Errorf("Username = %q, want alice", user.Username)
	}

	if _, ok := store.GetUserByUsername("nobody"); ok {
		t.Error("GetUserByUsername() reported an unknown user as found")
	}
}

func TestSeedAdminOnlySeedsAnEmptyStore(t *testing.T) {
	store := NewAuthStore()
	store.SeedAdmin()

	admin, ok := store.GetUserByUsername("admin")
	if !ok {
		t.Fatal("SeedAdmin() did not create the admin user")
	}
	if admin.Role != "admin" {
		t.Errorf("seeded role = %q, want admin", admin.Role)
	}

	// Re-seeding an existing deployment must not reset a changed password.
	store.SeedAdmin()
	if _, err := store.Login("admin", "admin"); err != nil {
		t.Fatalf("admin login after re-seed = %v", err)
	}
}

func TestSeedAdminSkipsWhenUsersExist(t *testing.T) {
	store := enabledStore(t)
	if _, err := store.Register("alice", "pw", "Org1MSP", "user"); err != nil {
		t.Fatalf("Register() = %v", err)
	}

	store.SeedAdmin()

	if _, ok := store.GetUserByUsername("admin"); ok {
		t.Error("SeedAdmin() seeded an admin into a store that already had users")
	}
}

// The store is shared by every in-flight HTTP request, so concurrent access
// must not race.
func TestAuthStoreIsSafeForConcurrentUse(t *testing.T) {
	store := enabledStore(t)
	store.SeedAdmin()

	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()

			username := string(rune('a'+n)) + "-user"
			_, _ = store.Register(username, "pw", "Org1MSP", "user")
			_, _ = store.Login(username, "pw")
			_, _ = store.GetUserByUsername(username)
			store.SeedAdmin()
		}(i)
	}
	wg.Wait()
}
