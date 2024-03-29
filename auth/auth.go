package auth

import (
	"errors"

	"github.com/MarekWojt/gertdns/util"
	"github.com/raja/argon2pw"
)

type PasswordAuthenticationRequest struct {
	User     string
	Password string
	Domain   string
}

type userRaw struct {
	Password string
	Hashed   bool
	Domains  []string
}

type user struct {
	password string
	domains  map[string]string
}

var parsedUsers map[string]user = make(map[string]user)

func (user user) Authenticate(password string) (bool, error) {
	return argon2pw.CompareHashWithPassword(user.password, password)
}

func (selfUser *userRaw) Tidy() (user, error) {
	if !selfUser.Hashed {
		pw, err := argon2pw.GenerateSaltedHash(selfUser.Password)
		if err != nil {
			return user{}, err
		}
		selfUser.Password = pw
		selfUser.Hashed = true
	}

	// Create a map => faster access times
	parsedDomains := make(map[string]string)
	for _, domain := range selfUser.Domains {
		parsedDomain := util.ParseDomain(domain)
		parsedDomains[parsedDomain] = parsedDomain
	}

	parsedUser := user{
		password: selfUser.Password,
		domains:  parsedDomains,
	}

	return parsedUser, nil
}

func IsPasswordAuthenticated(request PasswordAuthenticationRequest) (bool, error) {
	currentUser, found := parsedUsers[request.User]
	if !found {
		return false, errors.New("user does not exist")
	}

	if _, ok := currentUser.domains[request.Domain]; !ok {
		return false, errors.New("user does not have access to this domain")
	}

	return currentUser.Authenticate(request.Password)
}

func Init(authFilePath string) error {
	users, err := loadAuthFile(authFilePath)
	if err != nil {
		return err
	}

	for name, user := range users {
		parsedUser, err := user.Tidy()
		if err != nil {
			return err
		}

		parsedUsers[name] = parsedUser
	}

	writeAuthFile(authFilePath, users)

	return nil
}
