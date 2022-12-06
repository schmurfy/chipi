package main

import (
	"flag"
	"fmt"
	"go/parser"
	"go/token"
	"os"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/schmurfy/chipi/gen"
)

func main() {
	folder := ""
	noCreate := false

	flag.StringVar(&folder, "dir", "", "which folder to generate data for")
	flag.BoolVar(&noCreate, "dry", false, "only shows which files would be created")
	flag.Parse()

	if folder == "" {
		flag.Usage()
		os.Exit(1)
	}

	err := gen.InspectDir(folder, noCreate)
	if err != nil {
		panic(err)
	}

	os.Exit(0)

	data, err := os.ReadFile("./example/pet.go")
	if err != nil {
		panic(err)
	}

	fset := token.NewFileSet()
	f, err := decorator.ParseFile(fset, "", data, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	// record := false

	for _, node := range f.Decls {
		if decl, ok := node.(*dst.GenDecl); ok {
			for _, spec := range decl.Specs {
				if tt, ok := spec.(*dst.TypeSpec); ok {
					if st, ok := tt.Type.(*dst.StructType); ok {
						// fmt.Printf(":type: %s %v\n", tt.Name, st.Incomplete)

						for _, f := range st.Fields.List {
							// we found a sub structure, iterate on its fields
							if sectionStruct, ok := f.Type.(*dst.StructType); ok {
								for _, sectionFields := range sectionStruct.Fields.List {
									fmt.Printf("%s[%s] %s\n", tt.Name, f.Names[0], sectionFields.Names[0])
								}
							}
						}

					}

				}
			}
		}

	}

}
