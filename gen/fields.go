package gen

import (
	"fmt"
	"io"
	"strings"
	"text/template"

	"github.com/dave/dst"
)

type commentedField struct {
	Parent  string
	Section string
	Field   string

	Description string
	Example     string
}

func (cf commentedField) HasDescription() bool {
	return cf.Description != ""
}

func (cf commentedField) HasExample() bool {
	return cf.Example != ""
}

var paramTemplate = template.Must(template.New("param_template").Parse(`
	{{ $fields := .Fields }}
	{{ $sep := .StrSep }}
	{{ with $first := index .Fields 0 }}

	{{ with $first }}
	func (*{{.Parent}}) CHIPI_{{.Section}}_Annotations(attr string) *openapi3.Parameter {
	{{end}}
		switch attr {
		{{ range $fields }}
		case "{{ .Field }}":
			return &openapi3.Parameter{
				{{ if .HasDescription }}
					Description: strings.ReplaceAll({{ $sep }}{{.Description}}{{ $sep }}, gen.RepBackticks, gen.Backticks),
				{{ end }}

				{{ if .HasExample }}
					Example: {{ $sep }}{{ .Example }}{{ $sep }},
				{{end}}
			}
		{{ end }}
		}

		return nil
	}

	{{ end }}
`))

func GenerateFieldAnnotations(w io.Writer, f *dst.File, pkgName string) error {
	group := map[string][]commentedField{}

	err := inspectStructures(f, func(parentStructName string, sectionName string, fieldName string, data map[string]string) error {
		if sectionName == "Operation" {
			return nil
		}

		key := fmt.Sprintf("%s::%s", parentStructName, sectionName)
		if _, exists := group[key]; !exists {
			group[key] = []commentedField{}
		}

		cf := commentedField{
			Parent:  parentStructName,
			Section: sectionName,
			Field:   fieldName,
		}

		for k, v := range data {
			switch k {
			case "description":
				cf.Description = strings.ReplaceAll(v, "`", repBackticks)
			case "example":
				cf.Example = strings.ReplaceAll(v, "`", repBackticks)
			case "":
				// don't freak out if this is just a comment
				// with no properties

			default:
				return fmt.Errorf("unknown property %q", k)
			}
		}

		group[key] = append(group[key], cf)
		return nil
	})

	if err != nil {
		return err
	}

	if len(group) > 0 {
		for _, cfs := range group {
			err := paramTemplate.Execute(w, map[string]interface{}{
				"Fields": cfs,
				"StrSep": "`",
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}
