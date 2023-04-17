package cr

import "github.com/google/go-containerregistry/pkg/authn"

type registryOptions struct {
	ca           string
	useHTTP      bool
	withoutAuth  bool
	useDigest    bool
	authKeyChain authn.Keychain
}

type Option func(options *registryOptions)

// WithCA use custom CA certificate
func WithCA(ca string) Option {
	return func(options *registryOptions) {
		options.ca = ca
	}
}

// WithInsecureSchema use http schema instead of https
func WithInsecureSchema() Option {
	return func(options *registryOptions) {
		options.useHTTP = true
	}
}

// WithDisabledAuth don't use authConfig
func WithDisabledAuth() Option {
	return func(options *registryOptions) {
		options.withoutAuth = true
	}
}

// WithUseDigest use digest except of tag
func WithUseDigest() Option {
	return func(options *registryOptions) {
		options.useDigest = true
	}
}
