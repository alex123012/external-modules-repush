package templates

import (
	"bytes"
	"strings"
	"text/template"

	sprig "github.com/go-task/slim-sprig"
)

type Template string

// Render returns formatted VRL rule with provided args.
func (r Template) Render(args map[string]interface{}) (string, error) {
	var res bytes.Buffer

	tpl, err := template.New("template").
		Funcs(sprig.TxtFuncMap()).
		Parse(string(r))
	if err != nil {
		return "", err
	}

	err = tpl.Execute(&res, args)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(res.String()), nil
}
