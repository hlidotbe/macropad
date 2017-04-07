package auxilium

import "net/http"

// LoginService provides a way to fetch an authentification token from Auxilium
type LoginService struct {
	client *Client
}

// Authenticate sends the userName and password and return a token
func (l *LoginService) Authenticate(userName string, password string) (string, *http.Response, error) {
	opt := loginOptions{
		User: &loginUser{
			Email:    userName,
			Password: password,
		},
		Identifier: l.client.ApplicationIdentifier,
	}
	req, err := l.client.NewRequest("POST", "user/request_token", opt)
	if err != nil {
		return "", nil, err
	}
	lToken := loginToken{}
	resp, err := l.client.Do(req, &lToken)
	if err != nil {
		return "", resp, err
	}
	if len(l.client.token) <= 0 {
		l.client.token = lToken.Token
	}
	return lToken.Token, resp, nil
}

type loginOptions struct {
	User       *loginUser `json:"user"`
	Identifier string     `json:"identifier"`
}

type loginUser struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginToken struct {
	Token string `json:"token"`
}
