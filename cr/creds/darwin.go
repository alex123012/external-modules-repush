//go:build darwin

package creds

import (
	"github.com/docker/docker-credential-helpers/osxkeychain"
	"github.com/google/go-containerregistry/pkg/authn"
)

func GetDarwinKeyChain() authn.Helper {
	return osxkeychain.Osxkeychain{}
}
