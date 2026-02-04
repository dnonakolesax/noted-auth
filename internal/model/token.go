package model

type TokenDTO struct { //nolint:recvcheck // autogen issues
	AccessToken     string `json:"access_token"`
	ExpiresIn       int    `json:"expires_in"`
	RefreshExp      int    `json:"refresh_expires_in"`
	RefreshToken    string `json:"refresh_token"`
	IDToken         string `json:"id_token"`
	TokenType       string `json:"token_type"`
	NotBeforePolicy int    `json:"not-before-policy"`
	SessionState    string `json:"session_state"`
	State           string `json:"state"`
	Scope           string `json:"scope"`
	ReturnURL       string `json:"return_url"`
}

type TokenGRPCDTO struct { //nolint:recvcheck // autogen issues
	AccessToken  string
	RefreshToken string
	IDToken      string
	ExpiresIn    int
	RefreshExp   int
	UserID       string
}

type IntrospectDTO struct { //nolint:recvcheck // autogen issues
	Active  bool   `json:"active"`
	Subject string `json:"sub"`
}

func (td *TokenGRPCDTO) ToTokenDTO() TokenDTO {
	return TokenDTO{
		AccessToken:  td.AccessToken,
		RefreshToken: td.RefreshToken,
		IDToken:      td.IDToken,
		ExpiresIn:    td.ExpiresIn,
		RefreshExp:   td.RefreshExp,
	}
}
