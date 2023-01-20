package wrapper

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/schmurfy/chipi/schema"
	"github.com/schmurfy/chipi/shared"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var (
	_tracer  = otel.Tracer("chipiii")
	_noValue = reflect.Value{}
)

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
		param := strings.Split(
			strings.Trim(value, `[]`),
			",")
		sliceType := fieldType.Elem()
		setValue := reflect.New(reflect.SliceOf(sliceType)).Elem()
		for _, v := range param {
			vv, err := convertValue(sliceType, strings.TrimSpace(v))
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
		return reflect.ValueOf(
			strings.Trim(value, `"`),
		).Convert(fieldType), nil

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

func createFilledRequestObject(r *http.Request, obj interface{}, parsingErrors map[string]string) (ret reflect.Value, response reflect.Value, err error) {
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

	hasParamsErrors := false

	// path
	pathValue := ret.Elem().FieldByName("Path")
	rctx := chi.RouteContext(r.Context())
	for _, k := range rctx.URLParams.Keys {
		fieldValue := pathValue.FieldByName(k)
		if fieldValue.IsValid() {
			path := "request.path." + k
			err = setFValue(ctx,
				path,
				fieldValue,
				rctx.URLParam(k),
			)
			if err != nil {
				parsingErrors[path] = err.Error()
				hasParamsErrors = true
			}
		}
	}

	// query
	queryValue := ret.Elem().FieldByName("Query")
	if queryValue.IsValid() {
		for i := 0; i < queryValue.NumField(); i++ {

			queryFieldName := queryValue.Type().Field(i).Name
			structField, _ := queryValue.Type().FieldByName(queryFieldName)

			// Tag "json" overwrite the key
			parsedQueryFieldName := schema.ParseJsonTag(structField).Name
			if parsedQueryFieldName == structField.Name {
				parsedQueryFieldName = shared.ToSnakeCase(structField.Name)
			}
			path := "request.query." + parsedQueryFieldName

			if value, ok := r.URL.Query()[parsedQueryFieldName]; ok {
				err = setFValue(ctx,
					path,
					queryValue.Field(i),
					value[0],
				)
				if err != nil {
					parsingErrors[path] = err.Error()
					hasParamsErrors = true
				}
			}
		}
	}

	// header
	headerValue := ret.Elem().FieldByName("Header")
	if headerValue.IsValid() {
		for i := 0; i < headerValue.NumField(); i++ {

			attributeName := headerValue.Type().Field(i).Name
			structField, _ := headerValue.Type().FieldByName(attributeName)

			// Tag "name" overwrite the key
			name := structField.Tag.Get("name")
			headerName := attributeName
			if name != "" {
				headerName = name
			}
			path := "request.header." + attributeName
			if r.Header.Get(headerName) != "" {
				err = setFValue(ctx,
					path,
					headerValue.Field(i),
					r.Header.Get(headerName),
				)
				if err != nil {
					parsingErrors[path] = err.Error()
					hasParamsErrors = true
				}
			}
		}
	}

	if hasParamsErrors {
		err = errors.New("input parsing error")
		return
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

		path := "request.body"
		// call the request method if it implements a custom decoder
		if decoder, ok := ret.Interface().(BodyDecoder); ok {
			err = decoder.DecodeBody(r.Body, bodyObject, ret)
			if err != nil {
				parsingErrors[path] = err.Error()
				return
			}
		} else {
			err = fmt.Errorf(
				"structure %s needs to implement BodyDecoder interface",
				typ.Name(),
			)
			if err != nil {
				parsingErrors[path] = err.Error()
			}
			return
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

		parsingErrors := map[string]string{}

		vv, response, err = createFilledRequestObject(r, obj, parsingErrors)
		if err != nil {
			data, err := json.Marshal(parsingErrors)
			if err != nil {
				data = []byte(`{}`)
			}
			w.Header().Set("content-type", "application/json")
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, string(data))
			return
		}

		if rr, ok := vv.Interface().(HandlerWithRequestInterface); ok {
			err = rr.Handle(ctx, r, w)
		} else if rr, ok := vv.Interface().(HandlerInterface); ok {
			err = rr.Handle(ctx, w)
		}

		if err != nil {
			if rr, ok := vv.Interface().(ErrorHandlerInterface); ok {
				rr.HandleError(ctx, w, err)
			}

		} else if response.IsValid() {
			// encode response if any
			if encoder, ok := obj.(ResponseEncoder); ok {
				encoder.EncodeResponse(ctx, w, response.Interface())
			} else {
				err = fmt.Errorf(
					"structure %s needs to implement ResponseEncoder interface",
					vv.Type().Name(),
				)
				return

			}
		}

	}
}
