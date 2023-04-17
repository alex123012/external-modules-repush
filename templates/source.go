package templates

import (
	"encoding/json"

	"github.com/alex123012/external-modules-transfer/cr"
	"github.com/google/go-containerregistry/pkg/authn"
)

var externalModuleSource Template = `
---
apiVersion: deckhouse.io/v1alpha1
kind: ExternalModuleSource
metadata:
  name: {{.name}}
spec:
  releaseChannel: {{.releaseChannel}}
  registry:
    dockerCfg: {{.dockerCfg|b64enc}}
    repo: {{.repo}}
`

func RenderExternalModuleSource(name, repo, releaseChannel string, opts ...cr.Option) (string, error) {
	conf, host, err := cr.GetAuthConfig(repo, opts...)
	if err != nil {
		return "", err
	}
	confBytes, err := json.Marshal(map[string]map[string]*authn.AuthConfig{
		"auths": {
			host: conf,
		},
	})
	if err != nil {
		return "", err
	}
	rend, err := externalModuleSource.Render(map[string]interface{}{
		"name":           name,
		"releaseChannel": releaseChannel,
		"repo":           repo,
		"dockerCfg":      string(confBytes),
	})
	if err != nil {
		return "", err
	}
	return rend, nil
}
