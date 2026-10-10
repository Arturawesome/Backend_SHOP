package web

import (
	pbAuth "BE2_API_Shop/gen/auth"
	"encoding/json"
	"net/http"
	"strings"
)

type registerRequest struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Surname  string `json:"surname"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
}
type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	AccessToken string `json:"access_token"`
}

type changePasswordRequest struct {
	NewPassword string `json:"new_password"`
}

type changePasswordResponse struct {
	IsChanged bool `json:"is_changed"`
}

type resetPasswordRequest struct {
	Email string `json:"email"`
}

type resetPasswordResponse struct {
	IsRestored bool `json:"is_restored"`
}

// ResetPassword godoc
// @Summary Reset user password
// @Description Generates a new password for the user. The authorization service prints the generated password to its console as an email stub.
// @Tags auth
// @Accept json
// @Produce json
// @Param input body resetPasswordRequest true "User email"
// @Success 200 {object} resetPasswordResponse
// @Failure 400 {string} string "Bad Request"
// @Failure 404 {string} string "Not Found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /reset [post]
func (app *Application) ResetPassword(w http.ResponseWriter,
	r *http.Request) {

	var req resetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest),
			http.StatusBadRequest)
		return
	}

	grpcReq := pbAuth.RestorePasswordRequest{
		Email: req.Email}

	resp, err := app.grpcAuthClient.RestorePassword(
		r.Context(),
		&grpcReq)

	if err != nil {
		statusCode := grpcErrorToHTTP(err)

		http.Error(w, http.StatusText(statusCode), statusCode)
		return
	}

	w.Header().Set("Content-Type", "Application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(resetPasswordResponse{
		IsRestored: resp.IsRestored}); err != nil {

		app.logger.Error("failed to encode reset password response",
			"error", err)
	}
}

// ChangePassword godoc
// @Summary Change user password
// @Description Changes the password of the authenticated user
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param input body changePasswordRequest true "New password"
// @Success 200 {object} changePasswordResponse
// @Failure 400 {string} string "Bad Request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Not Found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /auth/password [patch]
func (app *Application) ChangePassword(w http.ResponseWriter,
	r *http.Request) {
	var req changePasswordRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	authHeader := r.Header.Get("Authorization")
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	grpcReq := pbAuth.ChangePasswordRequest{
		AccessToken: tokenString,
		NewPassword: req.NewPassword,
	}

	resp, err := app.grpcAuthClient.ChangePassword(
		r.Context(), &grpcReq)

	if err != nil {
		statusCode := grpcErrorToHTTP(err)
		http.Error(w, http.StatusText(statusCode), statusCode)
		return
	}

	w.Header().Set("Content-Type", "Application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(changePasswordResponse{
		IsChanged: resp.IsChanged}); err != nil {
		app.logger.Error("failed to encode change password response",
			"error", err)
	}
}

// Register godoc
// @Summary Register user
// @Description Creates a new user and returns a signed JWT access token
// @Tags auth
// @Accept json
// @Produce json
// @Param input body registerRequest true "Registration data"
// @Success 201 {object} authResponse
// @Failure 400 {string} string "Bad Request"
// @Failure 409 {string} string "Conflict"
// @Failure 500 {string} string "Internal Server Error"
// @Router /register [post]
func (app *Application) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	grpcReq := pbAuth.RegistrationRequest{
		Email:    req.Email,
		Name:     req.Name,
		Surname:  req.Surname,
		Phone:    req.Phone,
		Password: req.Password,
	}
	resp, err := app.grpcAuthClient.Registration(r.Context(), &grpcReq)
	if err != nil {
		http.Error(w, http.StatusText(grpcErrorToHTTP(err)),
			grpcErrorToHTTP(err))
		return
	}
	w.Header().Set("Content-Type", "Application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(authResponse{
		AccessToken: resp.AccessToken,
	})

}

// Login godoc
// @Summary Authenticate user
// @Description Authenticates user by email and password and returns a signed JWT access token
// @Tags auth
// @Accept json
// @Produce json
// @Param input body loginRequest true "Login data"
// @Success 200 {object} authResponse
// @Failure 400 {string} string "Bad Request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal Server Error"
// @Router /auth [post]
func (app *Application) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)

		return
	}
	grpcReq := pbAuth.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	}
	resp, err := app.grpcAuthClient.Login(r.Context(), &grpcReq)
	if err != nil {
		statusCode := grpcErrorToHTTP(err)

		http.Error(w, http.StatusText(statusCode), statusCode)
		return
	}
	w.Header().Set("Content-Type", "Application/json")
	json.NewEncoder(w).Encode(authResponse{
		AccessToken: resp.AccessToken,
	})

}
