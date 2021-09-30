package gen

import (
	"bytes"
	"fmt"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/dave/dst/decorator"
)

const repBackticks = "”"

var fileHeader = `
package %s

import (
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/schmurfy/chipi/gen"
)


var (
	// useless but required to not get a ù$* error about unused imports
	_ = strings.ToLower("")
	_ = gen.RepBackticks
)
`

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

	pkgName := file.Name.String()

	err = GenerateFieldAnnotations(buffer, file, pkgName)
	if err != nil {
		return err
	}

	err = GenerateOperationAnnotations(buffer, file, pkgName)
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

			header := fmt.Sprintf(fileHeader, pkgName)
			_, err = w.WriteString(header)
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
