package wrapper

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var (
	_tracer  = otel.Tracer("chipi")
	_noValue = reflect.Value{}
)

type BodyDecoder interface {
	DecodeBody(body io.ReadCloser, target interface{}) error
}

type ResponseEncoder interface {
	EncodeResponse(out http.ResponseWriter, obj interface{})
}

type HandlerInterface interface {
	Handle(context.Context, http.ResponseWriter) error
	HandleError(context.Context, http.ResponseWriter, error)
}

func convertValue(fieldType reflect.Type, value string) (reflect.Value, error) {
	switch fieldType.Kind() {
	case reflect.Ptr:
		fieldType := fieldType.Elem()
		setValue, err := convertValue(fieldType, value)
		if err != nil {
			return _noValue, err
		}
		setValuePtr := reflect.New(fieldType)
		setValuePtr.Elem().Set(setValue)
		return setValuePtr, nil

	case reflect.Slice:
		param := strings.Split(value, ",")
		sliceType := fieldType.Elem()
		setValue := reflect.New(reflect.SliceOf(sliceType)).Elem()
		for _, v := range param {
			vv, err := convertValue(sliceType, v)
			if err != nil {
				return _noValue, err
			}
			setValue = reflect.Append(setValue, vv)
		}
		return setValue, nil

	case reflect.Struct:
		setValue := reflect.New(fieldType)
		iface := setValue.Interface()
		err := json.Unmarshal([]byte(value), &iface)
		if err != nil {
			return _noValue, err
		}
		return setValue.Elem(), nil

	case reflect.String:
		return reflect.ValueOf(value).Convert(fieldType), nil

	case reflect.Bool:
		setValue, err := strconv.ParseBool(value)
		if err != nil {
			return _noValue, err
		}
		return reflect.ValueOf(setValue).Convert(fieldType), nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return _noValue, err
		}
		setValue := reflect.ValueOf(n).Convert(fieldType)
		return setValue, nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return _noValue, err
		}
		setValue := reflect.ValueOf(n).Convert(fieldType)
		return setValue, nil

	case reflect.Float32, reflect.Float64:
		x, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return _noValue, err
		}
		setValue := reflect.ValueOf(x).Convert(fieldType)
		return setValue, nil

	default:
		return reflect.Value{}, fmt.Errorf("invalid type: %v", fieldType.Kind())
	}
}

func setFValue(ctx context.Context, path string, f reflect.Value, value string) error {
	v, err := convertValue(f.Type(), value)

	if err != nil {
		return err
	}

	f.Set(v)

	trace.SpanFromContext(ctx).SetAttributes(attribute.String(path, value))
	return nil
}

func createFilledRequestObject(r *http.Request, obj interface{}) (ret reflect.Value, response reflect.Value, err error) {
	typ := reflect.TypeOf(obj)

	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	rr := reflect.ValueOf(obj)
	ret = reflect.New(typ)

	// copy already set field
	for i := 0; i < ret.Elem().NumField(); i++ {
		f := ret.Elem().Field(i)
		if f.CanSet() {
			f.Set(rr.Elem().Field(i))
		}
	}

	ctx := r.Context()

	// path
	pathValue := ret.Elem().FieldByName("Path")
	rctx := chi.RouteContext(r.Context())
	for _, k := range rctx.URLParams.Keys {
		fieldValue := pathValue.FieldByName(k)
		if fieldValue.IsValid() {
			err = setFValue(ctx,
				"request.path."+fieldValue.Type().Name(),
				fieldValue,
				rctx.URLParam(k),
			)
			if err != nil {
				return
			}
		}
	}

	// query
	queryValue := ret.Elem().FieldByName("Query")
	if queryValue.IsValid() {
		for k, v := range r.URL.Query() {
			f := queryValue.FieldByName(k)
			if f.IsValid() {
				err = setFValue(ctx,
					"request.query."+f.Type().Name(),
					f,
					v[0],
				)

				if err != nil {
					return
				}
			}
		}
	}

	// body
	bodyValue := ret.Elem().FieldByName("Body")
	if bodyValue.IsValid() {
		var bodyObject interface{}
		if bodyValue.Kind() == reflect.Ptr {
			body := reflect.New(bodyValue.Type().Elem())
			bodyValue.Set(body)
			bodyObject = bodyValue.Interface()
		} else {
			bodyObject = bodyValue.Addr().Interface()
		}

		// call the request method if it implements a custom decoder
		if decoder, ok := obj.(BodyDecoder); ok {
			err = decoder.DecodeBody(r.Body, bodyObject)
		}
	}

	response = ret.Elem().FieldByName("Response")

	return
}

func WrapRequest(obj interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var vv reflect.Value
		var response reflect.Value

		ctx, span := _tracer.Start(r.Context(), "WrapRequest")

		defer func() {
			if err != nil {
				span.RecordError(err)
			}
			span.End()
		}()

		vv, response, err = createFilledRequestObject(r, obj)
		if err != nil {
			return
		}

		var filledRequestObject HandlerInterface = vv.Interface().(HandlerInterface)
		err = filledRequestObject.Handle(ctx, w)
		if err != nil {
			filledRequestObject.HandleError(ctx, w, err)
		} else if response.IsValid() {
			// encode response if any
			if encoder, ok := obj.(ResponseEncoder); ok {
				encoder.EncodeResponse(w, response.Interface())
			}
		}

	}
}
