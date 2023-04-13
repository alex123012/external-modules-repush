package cr

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net"
	"net/http"
	"runtime"
	"time"

	"github.com/alex123012/external-modules-transfer/cr/creds"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

type client struct {
	registryURL string
	options     *registryOptions
}

func newClient(repo string, opts ...option) (*client, error) {
	regOpts := &registryOptions{}

	for _, opt := range opts {
		opt(regOpts)
	}

	if !regOpts.withoutAuth {
		regOpts.authConfig = getKeyChain(repo)
	}

	r := &client{
		registryURL: repo,
		options:     regOpts,
	}

	return r, nil
}

func getKeyChain(repo string) authn.Keychain {
	switch runtime.GOOS {
	case "windows":
		return authn.NewKeychainFromHelper(creds.GetDarwinKeyChain())
	case "darwin":
		return authn.NewKeychainFromHelper(creds.GetDarwinKeyChain())
	}
	return authn.DefaultKeychain
}

func (r *client) Image(imageTag string) (v1.Image, error) {
	ref, err := r.parseImageReference(imageTag)
	if err != nil {
		return nil, err
	}

	log.Println("pulling image:", ref)
	return remote.Image(ref, r.getImageOptions()...)
}

func (r *client) PushImage(imageTag string, image v1.Image) error {
	ref, err := r.parseImageReference(imageTag)
	if err != nil {
		return err
	}
	log.Println("uploading image:", ref)
	return remote.Write(ref, image, r.getImageOptions()...)
}

func (r *client) getImageOptions() []remote.Option {
	imageOptions := make([]remote.Option, 0)
	if !r.options.withoutAuth {
		imageOptions = append(imageOptions, remote.WithAuthFromKeychain(r.options.authConfig))
	}
	if r.options.ca != "" {
		imageOptions = append(imageOptions, remote.WithTransport(getHTTPTransport(r.options.ca)))
	}
	return imageOptions
}

func (r *client) parseImageReference(tag string) (name.Reference, error) {
	var nameOpts []name.Option
	if r.options.useHTTP {
		nameOpts = append(nameOpts, name.Insecure)
	}
	ref, err := name.ParseReference(r.registryURL+":"+tag, nameOpts...)
	if err != nil {
		return nil, fmt.Errorf("parsing image ref %q with tag %s: %w", r.registryURL, tag, err)
	}
	return ref, nil
}

func getHTTPTransport(ca string) (transport http.RoundTripper) {
	if ca == "" {
		return http.DefaultTransport
	}
	caPool, err := x509.SystemCertPool()
	if err != nil {
		panic(fmt.Errorf("cannot get system cert pool: %v", err))
	}

	caPool.AppendCertsFromPEM([]byte(ca))

	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig:       &tls.Config{RootCAs: caPool},
		TLSNextProto:          make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
	}
}
