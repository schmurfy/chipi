package gen

import (
	"bytes"
	"fmt"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"text/template"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
)

func FilterIncludeAll(fi os.FileInfo) bool {
	return true
}

func InspectDir(path string, noCreate bool) error {
	return InspectDirWithFilter(path, noCreate, FilterIncludeAll)
}

func InspectDirWithFilter(path string, noCreate bool, filterFunc func(fi os.FileInfo) bool) error {
	fset := token.NewFileSet()
	packages, err := decorator.ParseDir(fset, path, filterFunc, parser.ParseComments)
	if err != nil {
		return err
	}

	for _, pkg := range packages {
		for _, file := range pkg.Files {
			var w io.Writer
			generatedPath := filepath.Join(path, file.Name.String()) + ".generated.go"

			if noCreate {
				w = bytes.NewBufferString("")
			} else {
				w = sys.Open()
			}

			err := GenerateAnnotations(w, file, pkg)
			if err != nil {
				return err
			}

			// write file (or log)
			fmt.Printf("path: %s\n", path)
			// path := file.Name
		}
	}

	return nil
}

type inspectFunc func(parentStructName string, sectionName string, fieldName string, data map[string]string) error

var tmpl = template.Must(template.New("template").Parse(`
	{{ $fields := .Fields }}
	{{ $sep := .StrSep }}
	{{ with $first := index .Fields 0 }}

	{{ with $first }}
	func (*{{.Parent}}) Chipi_{{.Section}}_Annotations(attr string) *openapi3.Parameter {
	{{end}}
		switch attr {
		{{ range $fields }}
		case "{{ .Field }}":
			return &openapi3.Parameter{
				{{ if .HasDescription }}
					Description: {{ $sep }}{{.Description}}{{ $sep }},
				{{ end }}

				{{ if .HasExample }}
					Example: {{ $sep }}{{ .Example }}{{ $sep }},
				{{end}}
			}
		{{ end }}
		}
	}

	{{ end }}
`))

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

func GenerateAnnotations(w io.Writer, f *dst.File, pkg *dst.Package) error {
	header := fmt.Sprintf(`
	package %s

	import "github.com/getkin/kin-openapi/openapi3"
	
	`, pkg.Name)

	w.Write([]byte(header))

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
				cf.Description = v
			case "example":
				cf.Example = v
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

	for _, cfs := range group {
		err := tmpl.Execute(w, map[string]interface{}{
			"Fields": cfs,
			"StrSep": "`",
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func inspectStructures(f *dst.File, cb inspectFunc) error {
	for _, node := range f.Decls {
		if decl, ok := node.(*dst.GenDecl); ok {
			for _, spec := range decl.Specs {
				if tt, ok := spec.(*dst.TypeSpec); ok {

					// we found a struct
					if st, ok := tt.Type.(*dst.StructType); ok {

						err := inspectRequestStructure(tt, st, cb)
						if err != nil {
							return err
						}
					}
				}
			}
		}

	}

	return nil
}

// inspect each sub structures and invoke the callback whenever a
// comment is found
func inspectRequestStructure(parentTypeSpec *dst.TypeSpec, parentStruct *dst.StructType, cb inspectFunc) error {
	// we are inspecting the top level request structure
	// (ex: GetPetRequest)
	// look for sub structures fields (ex: Header, Body, Query)
	for _, sectionField := range parentStruct.Fields.List {
		if sectionStruct, ok := sectionField.Type.(*dst.StructType); ok {

			// the section struct might have comments too
			startDecoration := sectionField.Decorations().Start
			if len(startDecoration) > 0 {
				commentData, err := parseComment(startDecoration)
				if err != nil {
					return err
				}

				err = cb(
					parentTypeSpec.Name.String(),
					sectionField.Names[0].String(),
					"",
					commentData,
				)

				if err != nil {
					return err
				}

			}

			// we got one we need to inspect each field (ex: Id, Count)
			for _, sectionStructField := range sectionStruct.Fields.List {
				startDecoration := sectionStructField.Decorations().Start
				if len(startDecoration) > 0 {
					commentData, err := parseComment(startDecoration)
					if err != nil {
						return err
					}

					err = cb(
						parentTypeSpec.Name.String(),
						sectionField.Names[0].String(),
						sectionStructField.Names[0].String(),
						commentData,
					)

					// fmt.Printf("%s :: %s :: %s :: %v\n",
					// 	parentTypeSpec.Name,
					// 	sectionField.Names[0],
					// 	sectionStructField.Names[0],
					// 	commentData,
					// )

					if err != nil {
						return err
					}

				}
			}
		}
	}

	return nil
}
