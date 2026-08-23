package telejoon

import (
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

// Codec encodes and decodes a typed callback payload. The same codec is used
// on the button side (encoding) and the handler side (decoding) of a Route,
// so the two can never drift.
type Codec[A any] interface {
	Encode(A) ([]byte, error)
	Decode([]byte) (A, error)
}

// JSONCodec returns a codec that encodes payloads as JSON. Convenient for
// nested payloads, but verbose — mind Telegram's 64-byte callback_data limit.
func JSONCodec[A any]() Codec[A] {
	return jsonCodec[A]{}
}

type jsonCodec[A any] struct{}

func (jsonCodec[A]) Encode(args A) ([]byte, error) {
	return json.Marshal(args)
}

func (jsonCodec[A]) Decode(data []byte) (A, error) {
	var args A
	err := json.Unmarshal(data, &args)

	return args, err
}

// positionalCodec is the default codec: it encodes structs field-by-field
// (or a single primitive value) as ":"-separated, percent-escaped fields.
// Supported field kinds: string, bool, all int/uint sizes, float32/64.
type positionalCodec[A any] struct{}

func (positionalCodec[A]) Encode(args A) ([]byte, error) {
	raw, err := flattenArgs(reflect.ValueOf(args))
	if err != nil {
		return nil, err
	}

	escaped := make([]string, len(raw))
	for i, field := range raw {
		escaped[i] = url.QueryEscape(field)
	}

	return []byte(strings.Join(escaped, ":")), nil
}

func (positionalCodec[A]) Decode(data []byte) (A, error) {
	var args A

	var raw []string
	if len(data) > 0 {
		raw = strings.Split(string(data), ":")
	}

	fields := make([]string, len(raw))

	for i, field := range raw {
		unescaped, err := url.QueryUnescape(field)
		if err != nil {
			return args, fmt.Errorf("unescape field %d: %w", i, err)
		}

		fields[i] = unescaped
	}

	if err := fillArgs(reflect.ValueOf(&args).Elem(), fields); err != nil {
		return args, err
	}

	return args, nil
}

func flattenArgs(v reflect.Value) ([]string, error) {
	if v.Kind() == reflect.Struct {
		var fields []string

		for i := 0; i < v.NumField(); i++ {
			if !v.Type().Field(i).IsExported() {
				continue
			}

			raw, err := argToString(v.Field(i))
			if err != nil {
				return nil, fmt.Errorf("field %s: %w", v.Type().Field(i).Name, err)
			}

			fields = append(fields, raw)
		}

		if len(fields) == 0 {
			// Payload-less structs (e.g. NoData) encode as the empty payload.
			return []string{}, nil
		}

		return fields, nil
	}

	raw, err := argToString(v)
	if err != nil {
		return nil, err
	}

	return []string{raw}, nil
}

func fillArgs(v reflect.Value, fields []string) error {
	if v.Kind() == reflect.Struct {
		exported := 0

		for i := 0; i < v.NumField(); i++ {
			if !v.Type().Field(i).IsExported() {
				continue
			}

			if exported >= len(fields) {
				return fmt.Errorf("not enough fields for struct %s", v.Type())
			}

			if err := stringToArg(v.Field(i), fields[exported]); err != nil {
				return fmt.Errorf("field %s: %w", v.Type().Field(i).Name, err)
			}

			exported++
		}

		if exported != len(fields) {
			return fmt.Errorf("field count mismatch for struct %s", v.Type())
		}

		return nil
	}

	if len(fields) != 1 {
		return fmt.Errorf("expected 1 field, got %d", len(fields))
	}

	return stringToArg(v, fields[0])
}

func argToString(v reflect.Value) (string, error) {
	switch v.Kind() {
	case reflect.String:
		return v.String(), nil
	case reflect.Bool:
		return strconv.FormatBool(v.Bool()), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10), nil
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', -1, 64), nil
	default:
		return "", fmt.Errorf("unsupported kind %s (use JSONCodec for complex payloads)", v.Kind())
	}
}

func stringToArg(v reflect.Value, raw string) error {
	switch v.Kind() {
	case reflect.String:
		v.SetString(raw)
	case reflect.Bool:
		parsed, err := strconv.ParseBool(raw)
		if err != nil {
			return err
		}

		v.SetBool(parsed)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return err
		}

		v.SetInt(parsed)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		parsed, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			return err
		}

		v.SetUint(parsed)
	case reflect.Float32, reflect.Float64:
		parsed, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return err
		}

		v.SetFloat(parsed)
	default:
		return fmt.Errorf("unsupported kind %s", v.Kind())
	}

	return nil
}
