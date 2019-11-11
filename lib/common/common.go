package common

import (
	"fmt"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
)

type TokenAuth struct {
	token string
}

func NewTokenAuth(token string) TokenAuth {
	return TokenAuth{
		token: token,
	}
}

func (t TokenAuth) AuthenticateRequest(req runtime.ClientRequest, _ strfmt.Registry) error {
	req.SetHeaderParam("Authorization", fmt.Sprintf("Token %s", t.token))
	return nil
}
