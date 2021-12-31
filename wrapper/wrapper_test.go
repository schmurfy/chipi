package wrapper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/franela/goblin"
	"github.com/go-chi/chi/v5"
	"github.com/schmurfy/chipi/response"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type someData struct {
	N   uint
	Str string
}

type sharedDecoder struct{}

func (r *sharedDecoder) DecodeBody(body io.ReadCloser, target interface{}, obj interface{}) error {
	data := target.(*someData)
	data.Str = "some great string !"

	return nil
}

type createTestUser struct {
	sharedDecoder
	response.ErrorEncoder

	Path struct{}
	Body *someData
}

func (r *createTestUser) Handle(ctx context.Context, w http.ResponseWriter) error {
	encoder := json.NewEncoder(w)
	return encoder.Encode(r.Body)
}

func TestWrapper(t *testing.T) {
	g := goblin.Goblin(t)

	g.Describe("Wrapper", func() {

		g.Describe("setFieldValue", func() {
			type loc struct {
				Type string
			}

			type st struct {
				Int    int
				IntPtr *int

				Int8    int8
				Int8Ptr *int8

				Int16 int16
				Int32 int32
				Int64 int64

				Uint   uint
				Uint32 uint64
				Uint64 uint64

				Float32 float32
				Float64 float64

				ArrString []string
				ArrUint   []uint

				Str    string
				StrPtr *string

				Bool    bool
				BoolPtr *bool

				Loc    loc
				LocPtr *loc
			}
			ctx := context.Background()

			tests := []struct {
				Field    string
				Value    string
				Expected interface{}
			}{
				{"Int", "34", 34},
				{"IntPtr", "23", 23},

				{"Int8", "127", int8(127)},
				{"Int8Ptr", "-127", int8(-127)},

				{"Int16", "45", int16(45)},
				{"Int32", "274", int32(274)},
				{"Int64", strconv.FormatInt(math.MaxInt64, 10), int64(math.MaxInt64)},

				{"Uint", "579", uint(579)},
				{"Uint32", strconv.FormatUint(math.MaxUint32, 10), uint64(math.MaxUint32)},
				{"Uint64", strconv.FormatUint(math.MaxUint64, 10), uint64(math.MaxUint64)},

				{"ArrString", "a,b,toto", []string{"a", "b", "toto"}},
				{"ArrString", "[a,b,toto]", []string{"a", "b", "toto"}},
				{"ArrString", `["a","b","toto"]`, []string{"a", "b", "toto"}},
				{"ArrUint", "3,567,900", []uint{3, 567, 900}},
				{"ArrUint", "3,  567,  900", []uint{3, 567, 900}},

				{"Float32", "3.1415927", float32(3.1415927)},

				{"Str", "a few words", "a few words"},
				{"Bool", "true", true},

				{"LocPtr", `{"Type": "toto"}`, loc{Type: "toto"}},
				{"Loc", `{"Type": "titi"}`, loc{Type: "titi"}},
			}

			for _, tt := range tests {
				// bind to local value
				tt := tt
				g.It(fmt.Sprintf("should set direct field value (%s)", tt.Field), func() {
					st := st{}
					vv := reflect.ValueOf(&st).Elem().FieldByName(tt.Field)

					err := setFValue(ctx, "unused", vv, tt.Value)
					require.NoError(g, err)

					if strings.HasSuffix(tt.Field, "Ptr") {
						require.NotNil(g, vv.Interface())
						assert.Equal(g, tt.Expected, vv.Elem().Interface())
					} else {
						assert.Equal(g, tt.Expected, vv.Interface())
					}
				})
			}

		})

		g.Describe("incoming request", func() {
			type testRequest struct {
				Path struct {
					Id      int
					AString string
					B       bool
				}
				Query struct {
					Count *int
					Unset *string
				}

				PrivateString string
			}

			var req *http.Request
			var rctx *chi.Context
			var reqObject *testRequest

			g.BeforeEach(func() {
				var ok bool

				req = httptest.NewRequest("GET", "/user", nil)
				rctx = chi.NewRouteContext()
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				// path
				rctx.URLParams.Add("Id", "42")
				rctx.URLParams.Add("AString", "toto")
				rctx.URLParams.Add("B", "true")

				// query
				query := req.URL.Query()
				query.Set("count", "2")
				req.URL.RawQuery = query.Encode()

				m := &testRequest{
					PrivateString: "some private string",
				}

				parsingErrors := map[string]string{}
				vv, hasResponse, err := createFilledRequestObject(req, m, parsingErrors)
				require.NoError(g, err)

				require.IsType(g, &testRequest{}, vv.Interface())

				assert.False(g, hasResponse.IsValid())

				reqObject, ok = vv.Interface().(*testRequest)
				require.True(g, ok)

			})

			g.It("should fill wrapper with path variables", func() {
				assert.Equal(g, 42, reqObject.Path.Id)
				assert.Equal(g, "toto", reqObject.Path.AString)
				assert.Equal(g, true, reqObject.Path.B)
			})

			g.It("should fill wrapper with query variables", func() {
				require.NotNil(g, reqObject.Query.Count)
				assert.Equal(g, 2, *reqObject.Query.Count)
			})

			g.It("should include private data", func() {
				assert.Equal(g, "some private string", reqObject.PrivateString)
			})

		})

		g.Describe("custom body decoder", func() {

			g.It("should be called", func() {
				rctx := chi.NewRouteContext()
				ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)

				r := httptest.NewRequest("POST", "/", nil).WithContext(ctx)

				w := httptest.NewRecorder()
				writtenbody := bytes.NewBufferString("")
				w.Body = writtenbody

				handler := WrapRequest(&createTestUser{})

				handler(w, r)

				assert.JSONEq(g, `{"N": 0, "Str": "some great string !"}`, writtenbody.String())
			})
		})
	})
}

func BenchmarkDecoding(b *testing.B) {
	b.Run("int32", func(b *testing.B) {
		var n int32 = 42

		b.Run("direct", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				n, err := strconv.ParseInt("42", 10, 32)
				if err != nil {
					b.Fatalf("err: %s", err.Error())
				}

				if n != 42 {
					b.Fatalf("wrong value: %d", n)
				}
			}
		})

		b.Run("reflect", func(b *testing.B) {
			typ := reflect.TypeOf(n)
			for i := 0; i < b.N; i++ {
				_, err := convertValue(typ, "42")
				if err != nil {
					b.Fatalf("err: %s", err.Error())
				}
			}
		})

	})

}
