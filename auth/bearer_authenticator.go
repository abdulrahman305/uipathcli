package auth

import (
	"fmt"
	"net/url"

	"github.com/UiPath/uipathcli/cache"
)

type BearerAuthenticator struct {
	Cache cache.Cache
}

func (a BearerAuthenticator) Auth(ctx AuthenticatorContext) AuthenticatorResult {
	if !a.enabled(ctx) {
		return *AuthenticatorSuccess(ctx.Request.Header, ctx.Config)
	}
	config, err := a.getConfig(ctx)
	if err != nil {
		return *AuthenticatorError(fmt.Errorf("Invalid bearer authenticator configuration: %v", err))
	}
	identityBaseUri := config.IdentityUri
	if identityBaseUri == nil {
		requestUrl, err := url.Parse(ctx.Request.URL)
		if err != nil {
			return *AuthenticatorError(fmt.Errorf("Invalid request url '%s': %v", ctx.Request.URL, err))
		}
		identityBaseUri, err = url.Parse(fmt.Sprintf("%s://%s/identity_", requestUrl.Scheme, requestUrl.Host))
		if err != nil {
			return *AuthenticatorError(fmt.Errorf("Invalid identity url '%s': %v", ctx.Request.URL, err))
		}
	}

	identityClient := identityClient(a)
	tokenRequest := newClientCredentialTokenRequest(
		*identityBaseUri,
		config.ClientId,
		config.ClientSecret,
		ctx.Insecure)
	tokenResponse, err := identityClient.GetToken(*tokenRequest)
	if err != nil {
		return *AuthenticatorError(fmt.Errorf("Error retrieving bearer token: %v", err))
	}
	ctx.Request.Header["Authorization"] = "Bearer " + tokenResponse.AccessToken
	return *AuthenticatorSuccess(ctx.Request.Header, ctx.Config)
}

func (a BearerAuthenticator) enabled(ctx AuthenticatorContext) bool {
	return ctx.Config["clientId"] != nil && ctx.Config["clientSecret"] != nil
}

func (a BearerAuthenticator) getConfig(ctx AuthenticatorContext) (*BearerAuthenticatorConfig, error) {
	clientId, err := a.parseRequiredString(ctx.Config, "clientId")
	if err != nil {
		return nil, err
	}
	clientSecret, err := a.parseRequiredString(ctx.Config, "clientSecret")
	if err != nil {
		return nil, err
	}
	var uri *url.URL
	uriString, err := a.parseRequiredString(ctx.Config, "uri")
	if err == nil {
		uri, err = url.Parse(uriString)
		if err != nil {
			return nil, fmt.Errorf("Error parsing identity uri: %v", err)
		}
	}
	return NewBearerAuthenticatorConfig(clientId, clientSecret, uri), nil
}

func (a BearerAuthenticator) parseRequiredString(config map[string]interface{}, name string) (string, error) {
	value := config[name]
	result, valid := value.(string)
	if !valid || result == "" {
		return "", fmt.Errorf("Invalid value for %s: '%v'", name, value)
	}
	return result, nil
}
