package gzapi

// Login authenticates with the GZCTF platform using stored credentials
func (cs *GZAPI) Login() error {
	if err := cs.post("/api/account/login", cs.Creds, nil); err != nil {
		return err
	}
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
	if err := cs.post("/api/account/register", registerForm, nil); err != nil {
		return err
	}
	return nil
}

// Logout ends the current session on the GZCTF platform
func (cs *GZAPI) Logout() error {
	if err := cs.post("/api/account/logout", nil, nil); err != nil {
		return err
	}
	return nil
}
