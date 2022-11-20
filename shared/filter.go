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
	Scope AttributeScope
	path  []string
}

func (ai AttributeInfo) String() string {
	return fmt.Sprintf("[%+v] <%s>", ai.Scope, ai.Path())
}

func (ai AttributeInfo) Empty() bool {
	return len(ai.path) == 0
}

func (ai AttributeInfo) Path() string {
	return strings.Join(ai.path, ".")
}

func (ai AttributeInfo) AppendPath(segment string) AttributeInfo {
	return AttributeInfo{
		Scope: ai.Scope,
		path:  append(ai.path, segment),
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
