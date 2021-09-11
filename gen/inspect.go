package gen

import (
	"github.com/dave/dst"
)

type inspectFunc func(parentStructName string, sectionName string, fieldName string, data map[string]string) error

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
