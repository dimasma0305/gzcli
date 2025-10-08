package gzapi

// LoginResponse represents the response from the login endpoint
type LoginResponse struct {
	Succeeded bool `json:"succeeded"`
}

// Login authenticates with the GZCTF platform using stored credentials
func (cs *GZAPI) Login() error {
	var response LoginResponse
	if err := cs.post("/api/account/login", cs.Creds, &response); err != nil {
		return err
	}
	// Note: We trust HTTP 200 status code (already validated in post())
	// Empty response body (e.g., already authenticated) is acceptable
	// Only check succeeded field if it was actually set in the response
	return nil
}

// RegisterForm contains the data required for user registration
type RegisterForm struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// Register creates a new user account on the GZCTF platform
func (cs *GZAPI) Register(registerForm *RegisterForm) error {
	var response LoginResponse
	if err := cs.post("/api/account/register", registerForm, &response); err != nil {
		return err
	}
	return nil
}

// Logout ends the current session on the GZCTF platform
func (cs *GZAPI) Logout() error {
	var response map[string]interface{}
	if err := cs.post("/api/account/logout", nil, &response); err != nil {
		return err
	}
	return nil
}
