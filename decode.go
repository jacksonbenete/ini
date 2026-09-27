package ini

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

func getDefaultOpts() DecodeOptions {
	return DecodeOptions{
		CommentSyntax:      []byte{';'},
		AllowInlineComment: false,
		DuplicateAbort:     true,
		DuplicateOverride:  false,
		QuotedValues:       false,
	}
}

func Unmarshal(data []byte, v any) error {
	return UnmarshalWithOpt(data, v, getDefaultOpts())
}

func UnmarshalIO(data io.Reader, v any) error {
	return UnmarshalIOWithOpt(data, v, getDefaultOpts())
}

// DecodeOptions define your parsed INI dialect of choice.
// CommentSyntax allows defining extra or alternative runes for comment apart from ';'.
// AllowInlineComment will remove any comment syntax and won't interpret that as part of value.
// DuplicateAbort returns error when duplicate key or section is found.
// DuplicateOverride chose to override previous declared key/section with the same name.
// QuotedValues won't trim quotes " or ' from value.
type DecodeOptions struct {
	CommentSyntax      []byte
	AllowInlineComment bool
	DuplicateAbort     bool
	DuplicateOverride  bool
	QuotedValues       bool
}

func UnmarshalWithOpt(data []byte, v any, opt DecodeOptions) error {
	var d decodeState

	d.init(bytes.NewReader(data), opt)
	return d.unmarshal(v)
}

func UnmarshalIOWithOpt(data io.Reader, v any, opt DecodeOptions) error {
	var d decodeState

	d.init(data, opt)
	return d.unmarshal(v)
}

func (d *decodeState) unmarshal(v any) error {
	val := reflect.ValueOf(v)
	if val.IsNil() || val.Kind() != reflect.Pointer {
		return errors.New("invalid type to unmarshal: " + val.Type().String())
	}

	d.decodedSections[d.nextSection] = make(map[string]any)
	reSection := regexp.MustCompile(`(?m)^\[(.*)\]$`)

	for d.scan.Scan() {
		line := d.scan.Text()

		// For now won't distinguish between empty or no match, otherwise: use Regexp.FindStringIndex or Regexp.FindStringSubmatch.
		matchSection := reSection.FindStringSubmatch(line)
		if len(matchSection) > 0 {
			// Found section.
			section := matchSection[1]
			// TODO signal (perhaps on *decodeState) that we're in a section, so that every next Scan attribute k/v to the inner section related struct.
			d.decodedSections[section] = make(map[string]any)
			// TODO include guardrails such as if section already exist, check options for what to do.
			d.nextSection = section
			continue
		} else {
			matchKV := strings.Split(line, "=")
			if len(matchKV) == 1 {
				// Didn't find key/value.
				// TODO check standard to see if we ScanContinue or ScanAbort.
				continue
			}
			// Found key/value.
			matchKey := strings.TrimSpace(matchKV[0])
			matchVal := strings.TrimSpace(matchKV[1])

			m := d.decodedSections[d.nextSection]

			// TODO include guardrails such as if key already exist, check options for what to do.
			m[matchKey] = matchVal
		}
	}

	// TODO should return my decodeState error instead?
	if err := d.scan.Err(); err != nil {
		return err
	}

	d.nextSection = ""
	err := d.unmarshalStruct(v)
	if err != nil {
		return err
	}

	return nil
}

type errorContext struct {
	Err    error
	Struct reflect.Type
	Stack  []string
}

type decodeState struct {
	scan            *bufio.Scanner
	errorContext    *errorContext
	decodedSections map[string]map[string]any
	nextSection     string
	opt             DecodeOptions
}

func (d *decodeState) init(data io.Reader, opt DecodeOptions) *decodeState {
	d.scan = bufio.NewScanner(data)
	d.errorContext = new(errorContext)
	d.decodedSections = make(map[string]map[string]any)
	d.opt = opt

	return d
}

func (d *decodeState) unmarshalStruct(v any) error {
	val := reflect.ValueOf(v)
	t := val.Elem().Type()

	structFields := t.NumField()

	for i := range structFields {
		structField := t.Field(i)
		structValue := val.Elem().Field(i)

		tag := structField.Tag.Get("ini")

		// Ignore unset tag.
		if tag == "" || tag == "-" {
			continue
		}

		value, ok := d.decodedSections[tag]
		if !ok {
			// If not found, could still be on root. Continue.
			value = d.decodedSections[d.nextSection]
		}

		if structValue.Kind() == reflect.Struct {
			for f, sv := range structValue.Fields() {
				fTag := f.Tag.Get("ini")

				err := setValue(sv, value[fTag])
				if err != nil {
					return fmt.Errorf("field %s: %w", structField.Name, err)
				}
			}
		} else {
			err := setValue(structValue, value[tag])
			if err != nil {
				return fmt.Errorf("field %s: %w", structField.Name, err)
			}
		}
	}

	return nil
}

func setValue(dst reflect.Value, src any) error {
	if src == nil {
		return nil
	}

	switch dst.Kind() {
	case reflect.String:
		s, ok := src.(string)
		if !ok {
			return fmt.Errorf("expected string, got %T", src)
		}
		dst.SetString(s)

	case reflect.Bool:
		b, ok := src.(bool)
		if !ok {
			s, ok := src.(string)
			if !ok {
				return fmt.Errorf("expected string 'true' or 'false', got %T", src)
			}

			sb := false

			switch s {
			case "true", "t", "yes", "y", "on":
				sb = true
			case "false", "f", "no", "n", "off":
				sb = false
			}

			b = sb
		}
		dst.SetBool(b)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, ok := src.(int64)
		if !ok {
			s, ok := src.(string)
			if !ok {
				return fmt.Errorf("expected string integer, got %T", src)
			}
			si, err := strconv.Atoi(s)
			if err != nil {
				return err
			}
			n = int64(si)
		}
		dst.SetInt(n)

	case reflect.Float32, reflect.Float64:
		n, ok := src.(float64)
		if !ok {
			s, ok := src.(string)
			if !ok {
				return fmt.Errorf("expected string float, got %T", src)
			}
			si, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return err
			}
			n = float64(si)
		}
		dst.SetFloat(n)

	case reflect.Slice:
		s, ok := src.(string)
		if !ok {
			return fmt.Errorf("expected a comma separated list written in a string format to get parsed, got %T", src)
		}
		list := strings.Split(s, ",")
		trimmed := make([]string, 0)
		for _, l := range list {
			trimmed = append(trimmed, strings.TrimSpace(l))
		}
		if len(trimmed) == 0 {
			return fmt.Errorf("expected a comma separated list, got empty: %v", src)
		}
		err := setSliceValue(dst, trimmed)
		if err != nil {
			return err
		}

	default:
		return fmt.Errorf("unsupported type: %s", dst.Type())
	}

	return nil
}

func setSliceValue(dst reflect.Value, src []string) error {
	list := reflect.MakeSlice(dst.Type(), 0, len(src))
	t := dst.Type().Elem()
	innerType := dst.Type().Elem().Kind()

	for _, s := range src {
		var parsed any

		switch innerType {
		case reflect.String:
			parsed = s

		case reflect.Bool:
			sb := false
			switch s {
			case "true", "t", "yes", "y", "on":
				sb = true
			case "false", "f", "no", "n", "off":
				sb = false
			}

			parsed = sb

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			res, err := strconv.Atoi(s)
			if err != nil {
				return fmt.Errorf("cannot convert list to int: %v", src)
			}
			parsed = res

		case reflect.Float32, reflect.Float64:
			res, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return fmt.Errorf("cannot convert list to float: %v", src)
			}
			parsed = res

		default:
			return fmt.Errorf("cannot convert list of unknown type: %v", src)
		}

		val := reflect.ValueOf(parsed)
		con := val.Convert(t)
		list = reflect.Append(list, con)
	}

	dst.Set(list)
	return nil
}
