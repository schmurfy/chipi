package gen

import (
	"fmt"
	"io"
	"strings"
	"text/template"

	"github.com/dave/dst"
)

type commentedOperation struct {
	Parent string

	Tags        string
	Summary     string
	Description string
	Deprecated  bool
}

func (cf commentedOperation) HasTags() bool {
	return cf.Tags != ""
}

func (cf commentedOperation) FormattedTags() string {
	parts := []string{}

	for _, s := range strings.Split(cf.Tags, ",") {
		parts = append(parts, fmt.Sprintf(`"%s"`, s))
	}

	return strings.Join(parts, ",")
}

func (cf commentedOperation) HasSummary() bool {
	return cf.Summary != ""
}

func (cf commentedOperation) HasDescription() bool {
	return cf.Description != ""
}

var operationTemplate = template.Must(template.New("operation_template").Parse(`
	{{ $fields := .Fields }}
	{{ $sep := .StrSep }}
	{{ with $first := index .Fields 0 }}

	{{ with $first }}
	func (*{{.Parent}}) CHIPI_Operation_Annotations() *openapi3.Operation {
	{{end}}

		return &openapi3.Operation{
			{{ if .HasTags }}
				Tags: []string{ {{.FormattedTags}} },
			{{ end }}

			{{ if .HasSummary }}
				Summary: {{ $sep }}{{ .Summary }}{{ $sep }},
			{{end}}

			{{ if .HasDescription }}
				Description: {{ $sep }}{{ .Description }}{{ $sep }},
			{{end}}


			Deprecated: {{.Deprecated}},
		}
	}

	{{ end }}
`))

func GenerateOperationAnnotations(w io.Writer, f *dst.File, pkgName string) error {
	operations := map[string][]commentedOperation{}

	err := inspectStructures(f, func(parentStructName string, sectionName string, fieldName string, data map[string]string) error {
		if sectionName != "Operation" {
			return nil
		}

		key := parentStructName
		if _, exists := operations[key]; !exists {
			operations[key] = []commentedOperation{}
		}

		cf := commentedOperation{
			Parent: parentStructName,
		}

		for k, v := range data {
			switch k {
			case "tag", "tags":
				cf.Tags = strings.ReplaceAll(v, "`", repBackticks)
			case "summary":
				cf.Summary = strings.ReplaceAll(v, "`", repBackticks)
			case "description":
				cf.Description = strings.ReplaceAll(v, "`", repBackticks)
			case "deprecated":
				cf.Deprecated = true
			case "":
				// don't freak out if this is just a comment
				// with no properties
			default:
				return fmt.Errorf("unknown property %s", k)
			}
		}

		operations[key] = append(operations[key], cf)
		return nil
	})

	if err != nil {
		return err
	}

	if len(operations) > 0 {
		for _, cfs := range operations {
			err := operationTemplate.Execute(w, map[string]interface{}{
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
