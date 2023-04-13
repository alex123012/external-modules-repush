package cr

import "github.com/google/go-containerregistry/pkg/authn"

type registryOptions struct {
	ca          string
	useHTTP     bool
	withoutAuth bool
	authConfig  authn.Keychain
}

type option func(options *registryOptions)

// WithCA use custom CA certificate
func WithCA(ca string) option {
	return func(options *registryOptions) {
		options.ca = ca
	}
}

// WithInsecureSchema use http schema instead of https
func WithInsecureSchema(insecure bool) option {
	return func(options *registryOptions) {
		options.useHTTP = insecure
	}
}

// WithDisabledAuth don't use authConfig
func WithDisabledAuth() option {
	return func(options *registryOptions) {
		options.withoutAuth = true
	}
}
