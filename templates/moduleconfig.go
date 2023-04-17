package templates

var moduleConfig Template = `
---
apiVersion: deckhouse.io/v1alpha1
kind: ModuleConfig
metadata:
  name: {{.name}}
spec:
  enabled: true
  settings: {}
  version: 1
`

func RenderModuleConfig(name string) (string, error) {
	rend, err := moduleConfig.Render(map[string]interface{}{
		"name": name,
	})
	if err != nil {
		return "", err
	}
	return rend, nil
}
