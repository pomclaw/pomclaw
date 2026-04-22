package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pomclaw/pomclaw/pkg/channels/gateway/store"
	"golang.org/x/crypto/bcrypt"
)

type registerReq struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type authResponse struct {
	AccessToken string `json:"access_token"`
	UserID      string `json:"user_id"`
	Username    string `json:"username"`
}

// Register creates a new user and returns a JWT.
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid JSON")
		return
	}
	if req.Username == "" || req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "username, email, and password are required")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "server_error", "failed to hash password")
		return
	}

	user, err := store.CreateUser(h.DB, req.Username, req.Email, string(hash))
	if err != nil {
		writeError(w, http.StatusConflict, "user_exists", "username or email already exists")
		return
	}

	token, err := h.generateToken(user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "server_error", "failed to generate token")
		return
	}

	writeJSON(w, http.StatusCreated, authResponse{
		AccessToken: token,
		UserID:      user.ID,
		Username:    user.Username,
	})
}

// Login authenticates a user and returns a JWT.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid JSON")
		return
	}
	if req.Username == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "username and password are required")
		return
	}

	user, err := store.GetUserByUsername(h.DB, req.Username)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid_credentials", "invalid username or password")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		writeError(w, http.StatusUnauthorized, "invalid_credentials", "invalid username or password")
		return
	}

	token, err := h.generateToken(user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "server_error", "failed to generate token")
		return
	}

	writeJSON(w, http.StatusOK, authResponse{
		AccessToken: token,
		UserID:      user.ID,
		Username:    user.Username,
	})
}

// jwtClaims is the payload for gateway JWTs.
type jwtClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// GetMe returns the authenticated user's info.
func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID := userIDFrom(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized", "missing user context")
		return
	}

	user, err := store.GetUserByID(h.DB, userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "user not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"user": map[string]string{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
		},
	})
}

// Refresh generates a new access token for the authenticated user.
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	userID := userIDFrom(r.Context())
	if userID == "" {
		// No authentication needed for refresh in this simple implementation
		// In a real app, you'd validate a refresh token
		writeError(w, http.StatusUnauthorized, "unauthorized", "missing user context")
		return
	}

	token, err := h.generateToken(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "server_error", "failed to generate token")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"access_token": token,
	})
}

// Logout is a no-op since we use JWT tokens (stateless).
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"message": "logged out",
	})
}

func (h *Handler) generateToken(userID string) (string, error) {
	claims := jwtClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.Secret))
}
