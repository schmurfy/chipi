package shared

import (
	"context"
	"fmt"
	"strings"
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

type FilterInterface interface {
	FilterRoute(ctx context.Context, method string, pattern string) (bool, error)
	FilterField(ctx context.Context, fieldInfo AttributeInfo) (bool, error)
}

// type RoutePatcherInterface interface {
// 	PatchRoute()
// }

// type FieldFilterInterface interface {
// 	FilterRoute()
// }
