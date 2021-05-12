package main

import (
	"flag"
	"fmt"
	"go/parser"
	"go/token"
	"io/ioutil"
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

	data, err := ioutil.ReadFile("./example/pet.go")
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
						fmt.Printf(":type: %s %v\n", tt.Name, st.Incomplete)

						for _, f := range st.Fields.List {
							// we found a sub structure, iterate on its fields
							if sectionStruct, ok := f.Type.(*dst.StructType); ok {
								for _, sectionFields := range sectionStruct.Fields.List {
									fmt.Printf("%s[%s] %s\n", tt.Name, f.Names[0], sectionFields.Names[0])
								}
							}
						}

						// dst.Inspect(tt, func(n dst.Node) bool {
						// 	// if ff, ok := n.(*dst.Field); ok {
						// 	// 	before := ff.Decorations().Start
						// 	// 	if len(before) > 0 {
						// 	// 		fmt.Printf("%s[%s]:\n", tt.Name, ff.Names[0])

						// 	// 		for _, line := range before {
						// 	// 			fmt.Printf("  '%s'\n", line)
						// 	// 		}

						// 	// 	}
						// 	// }

						// 	if str, ok := tt.Type.(*dst.StructType); ok {
						// 		for _, f := range str.Fields.List {

						// 			if str2, ok := f.Type.(*dst.StructType); ok {
						// 				for _, f2 := range str2.Fields.List {
						// 					fmt.Printf("%s[%s] %s\n", tt.Name, f.Names[0], f2.Names[0])
						// 				}
						// 			}
						// 		}

						// 	}

						// 	return true
						// })

					}

				}
			}
		}

	}

}
