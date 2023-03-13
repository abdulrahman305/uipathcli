package executor

import (
	"errors"
	"net/url"

	"github.com/UiPath/uipathcli/auth"
	"github.com/UiPath/uipathcli/config"
	"github.com/UiPath/uipathcli/log"
	"github.com/UiPath/uipathcli/output"
	"github.com/UiPath/uipathcli/plugin"
)

// The PluginExecutor implements the Executor interface and invokes the
// registered plugin for the executed command.
// The plugin takes care of sending the HTTP request or performing other
// operations.
type PluginExecutor struct {
	authenticators []auth.Authenticator
}

func (e PluginExecutor) executeAuthenticators(baseUri url.URL, authConfig config.AuthConfig, debug bool, insecure bool) (*auth.AuthenticatorResult, error) {
	authRequest := *auth.NewAuthenticatorRequest(baseUri.String(), map[string]string{})
	ctx := *auth.NewAuthenticatorContext(authConfig.Type, authConfig.Config, debug, insecure, authRequest)
	for _, authProvider := range e.authenticators {
		result := authProvider.Auth(ctx)
		if result.Error != "" {
			return nil, errors.New(result.Error)
		}
		ctx.Config = result.Config
		for k, v := range result.RequestHeader {
			ctx.Request.Header[k] = v
		}
	}
	return auth.AuthenticatorSuccess(ctx.Request.Header, ctx.Config), nil
}

func (e PluginExecutor) convertToPluginParameters(parameters []ExecutionParameter) []plugin.ExecutionParameter {
	result := []plugin.ExecutionParameter{}
	for _, parameter := range parameters {
		name := parameter.Name
		value := parameter.Value
		result = append(result, *plugin.NewExecutionParameter(name, value))
	}
	return result
}

func (e PluginExecutor) pluginParameters(context ExecutionContext) []plugin.ExecutionParameter {
	params := context.Parameters.Path
	params = append(params, context.Parameters.Query...)
	params = append(params, context.Parameters.Header...)
	params = append(params, context.Parameters.Body...)
	params = append(params, context.Parameters.Form...)
	return e.convertToPluginParameters(params)
}

func (e PluginExecutor) pluginAuth(auth *auth.AuthenticatorResult) plugin.AuthResult {
	return plugin.AuthResult{
		Header: auth.RequestHeader,
	}
}

func (e PluginExecutor) Call(context ExecutionContext, writer output.OutputWriter, logger log.Logger) error {
	auth, err := e.executeAuthenticators(context.BaseUri, context.AuthConfig, context.Debug, context.Insecure)
	if err != nil {
		return err
	}

	pluginAuth := e.pluginAuth(auth)
	pluginParams := e.pluginParameters(context)
	pluginContext := plugin.NewExecutionContext(
		context.Organization,
		context.Tenant,
		context.BaseUri,
		pluginAuth,
		context.Input,
		pluginParams,
		context.Insecure,
		context.Debug)
	return context.Plugin.Execute(*pluginContext, writer, logger)
}

func NewPluginExecutor(authenticators []auth.Authenticator) *PluginExecutor {
	return &PluginExecutor{authenticators}
}
