package gen

import (
	"bytes"
	"fmt"
	"go/parser"
	"go/token"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
)

const repBackticks = "”"

var fileHeader = `
package %s

import (
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)
`

var tmpl = template.Must(template.New("template").Parse(`
	{{ $fields := .Fields }}
	{{ $sep := .StrSep }}
	{{ with $first := index .Fields 0 }}

	{{ with $first }}
	func (*{{.Parent}}) CHIPI_{{.Section}}_Annotations(attr string) *openapi3.Parameter {
	{{end}}
		repBackticks := "”"
		backticks := "` + "`" + `"

		switch attr {
		{{ range $fields }}
		case "{{ .Field }}":
			return &openapi3.Parameter{
				{{ if .HasDescription }}
					Description: strings.ReplaceAll({{ $sep }}{{.Description}}{{ $sep }}, repBackticks, backticks),
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

func FilterIncludeAll(fi os.FileInfo) bool {
	return true
}

func InspectDir(path string, noWrite bool) error {
	return filepath.WalkDir(path, func(path string, entry os.DirEntry, err error) error {
		if filepath.Ext(path) == ".go" {
			pathWithoutExt := strings.TrimSuffix(path, filepath.Ext(path))

			err := generateDataForFile(pathWithoutExt, noWrite)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func generateDataForFile(path string, noWrite bool) error {
	fset := token.NewFileSet()

	data, err := ioutil.ReadFile(path + ".go")
	if err != nil {
		panic(err)
	}

	file, err := decorator.ParseFile(fset, path, data, parser.ParseComments)
	if err != nil {
		return err
	}

	generatedPath := path + ".generated.go"

	buffer := bytes.NewBufferString("")

	err = GenerateAnnotations(buffer, file, file.Name.String())
	if err != nil {
		return err
	}

	// write file (or log)
	if len(buffer.String()) > 0 {
		if noWrite {
			fmt.Printf("Would have written to %s\n", generatedPath)
		} else {
			w, err := os.OpenFile(generatedPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
			if err != nil {
				return err
			}

			_, err = w.WriteString(buffer.String())
			if err != nil {
				return err
			}

			fmt.Printf("Saved %s\n", generatedPath)
		}
	}

	return nil
}

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

func GenerateAnnotations(w io.Writer, f *dst.File, pkgName string) error {
	group := map[string][]commentedField{}

	err := inspectStructures(f, func(parentStructName string, sectionName string, fieldName string, data map[string]string) error {
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
			default:
				return fmt.Errorf("unknown property %s", k)
			}
		}

		group[key] = append(group[key], cf)
		return nil
	})

	if err != nil {
		return nil
	}

	if len(group) > 0 {
		header := fmt.Sprintf(fileHeader, pkgName)

		w.Write([]byte(header))

		for _, cfs := range group {
			err := tmpl.Execute(w, map[string]interface{}{
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
