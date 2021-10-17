package auth

type AuthenticationRequest struct {
	User     string
	Password string
	Domain   string
}

func IsAuthenticated(request AuthenticationRequest) (bool, error) {
	return true, nil
}
