package shared

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// unused for now,
type AttributeScope int

const (
	ScopeNone = iota
	// ScopeRequestBody
	// ScopeRequestResponse
)

type AttributeInfo struct {
	Scope     AttributeScope
	queryPath []string
	modelPath string
}

func (ai AttributeInfo) String() string {
	return fmt.Sprintf("[%+v] <%s> <%s>", ai.Scope, ai.QueryPath(), ai.ModelPath())
}

func (ai AttributeInfo) Empty() bool {
	return len(ai.queryPath) == 0
}

func (ai AttributeInfo) QueryPath() string {
	return strings.Join(ai.queryPath, ".")
}

func (ai AttributeInfo) WithModelPath(path string) AttributeInfo {
	ai.modelPath = path
	return ai
}

func (ai AttributeInfo) ModelPath() string {
	return ai.modelPath
}

func (ai AttributeInfo) AppendPath(segment string) AttributeInfo {
	return AttributeInfo{
		Scope:     ai.Scope,
		queryPath: append(ai.queryPath, segment),
		modelPath: ai.modelPath,
	}
}

type EnumEntry struct {
	Title interface{}
	Value interface{}
}
type Enum = []EnumEntry

// This object can implement FilterRoute/FilterField/EnumResolver/SchemaResolver
type ChipiCallbackInterface interface {
}

type FilterRouteInterface interface {
	FilterRoute(ctx context.Context, method string, pattern string) (bool, error)
}
type FilterFieldInterface interface {
	FilterField(ctx context.Context, fieldInfo AttributeInfo) (bool, error)
}
type EnumResolverInterface interface {
	EnumResolver(t reflect.Type) (bool, Enum)
}

type GenerateSchemaCallbackType func(t reflect.Type, fieldInfo AttributeInfo) (*openapi3.SchemaRef, error)

type SchemaResolverInterface interface {
	// CastName is the string after the `as:` in the field tag
	// e.g. `chipi:"as:uuid"` => CastName = "uuid"
	// The openapi3.Schema is the schema that will be used to define the CastName field
	// The boolean decide if we create a $ref in the openapi file
	SchemaResolver(fieldInfo AttributeInfo, castName string, fieldTyp reflect.Type, newSchemaCallbackType GenerateSchemaCallbackType) (*openapi3.Schema, bool)
}

type ExtraComponentsAndPathsInterface interface {
	ExtraComponentsAndPaths() (openapi3.Schemas, openapi3.Paths)
}

type ChipiCallbacks struct {
	FilterRouteInterface
	FilterFieldInterface
	EnumResolverInterface
	SchemaResolverInterface
	i ChipiCallbackInterface
}

var _ FilterRouteInterface = (*ChipiCallbacks)(nil)
var _ FilterFieldInterface = (*ChipiCallbacks)(nil)
var _ EnumResolverInterface = (*ChipiCallbacks)(nil)
var _ SchemaResolverInterface = (*ChipiCallbacks)(nil)

func NewChipiCallbacks(i ChipiCallbackInterface) ChipiCallbacks {
	return ChipiCallbacks{i: i}
}

func (c *ChipiCallbacks) FilterRoute(ctx context.Context, method string, pattern string) (bool, error) {
	if filterInterface, hasFilter := c.i.(FilterRouteInterface); c.i != nil && hasFilter {
		return filterInterface.FilterRoute(ctx, method, pattern)
	} else {
		return false, nil
	}
}

func (c *ChipiCallbacks) FilterField(ctx context.Context, fieldInfo AttributeInfo) (bool, error) {
	if filterInterface, hasFilter := c.i.(FilterFieldInterface); c.i != nil && hasFilter {
		return filterInterface.FilterField(ctx, fieldInfo)
	} else {
		return false, nil
	}
}

func (c *ChipiCallbacks) EnumResolver(t reflect.Type) (bool, Enum) {
	if enumInterface, hasEnum := c.i.(EnumResolverInterface); c.i != nil && hasEnum {
		return enumInterface.EnumResolver(t)
	} else {
		return false, nil
	}
}

func (c *ChipiCallbacks) SchemaResolver(fieldInfo AttributeInfo, castName string, fieldTyp reflect.Type, newSchemaCallbackType GenerateSchemaCallbackType) (*openapi3.Schema, bool) {
	if schemaInterface, hasSchema := c.i.(SchemaResolverInterface); c.i != nil && hasSchema {
		return schemaInterface.SchemaResolver(fieldInfo, castName, fieldTyp, newSchemaCallbackType)
	} else {
		return openapi3.NewObjectSchema(), false
	}
}

func (c *ChipiCallbacks) ExtraComponentsAndPaths() (openapi3.Schemas, openapi3.Paths) {
	if schemaInterface, hasExtraComponents := c.i.(ExtraComponentsAndPathsInterface); c.i != nil && hasExtraComponents {
		return schemaInterface.ExtraComponentsAndPaths()
	} else {
		return nil, openapi3.Paths{}
	}
}
