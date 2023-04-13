//go:build windows

package creds

import (
	"github.com/docker/docker-credential-helpers/wincred"
	"github.com/google/go-containerregistry/pkg/authn"
)

func GetWindowsKeyChain() authn.Helper {
	return wincred.Wincred{}
}
