package cliarg

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

func ParseArgs[T any]() (T, []string, error) {
	return Parse[T](os.Args[1:])
}

func Parse[T any](input []string) (result T, positionalArg []string, err error) {
	val := reflect.ValueOf(&result).Elem()
	typ := val.Type()
	data := newParseData()
	passedRequiredFieldMap := make(map[int]bool)
	for i := range typ.NumField() {
		var info fieldInfo
		info.Index = i
		field := typ.Field(i)
		value := val.Field(i)
		info.Type = field.Type.Kind()
		tag := field.Tag.Get("cliarg")
		if tag == "" {
			continue
		}
		var isSet bool
		for _, tagField := range splitTag(tag) {
			a, err := parseKeyValue(tagField)
			if err != nil {
				return returnParseError[T](err)
			}

			switch a[0] {
			case "long":
				fallthrough
			case "short":
				isSet = true
				ok := data.AddName(a[1], i)
				if !ok {
					return result, []string{}, conflictFlagNameError(a[1])
				}
			case "default":
				err = setDefaultValue(&info, &value, a[1])
				if err != nil {
					return returnParseError[T](fmt.Errorf("invalid value \"%s\" for %s", a[1], field.Name))
				}
			case "required":
				info.Required = true
				passedRequiredFieldMap[i] = false
			}

			if !isSet {
				return returnParseError[T](errors.New("either Long or Short must be set"))
			}
		}
		data.InfoMap[i] = info
	}

	for i, size := 0, len(input); i < size; i++ {
		arg := splitByEqual(input[i])
		info, ok := data.GetInfo(arg[0])
		if info.Required {
			passedRequiredFieldMap[info.Index] = true
		}
		if !ok {
			positionalArg = append(positionalArg, arg[0])
			continue
		}
		value := val.Field(info.Index)
		if info.Type == reflect.Bool {
			value.SetBool(!info.DefaultBool)
		} else if len(arg) == 2 {
			err = setValue(info, &value, arg[1], arg[0])
			if err != nil {
				return returnParseError[T](err)
			}
		} else if i+1 < size {
			i++
			err = setValue(info, &value, input[i], arg[0])
			if err != nil {
				return returnParseError[T](err)
			}
		} else {
			return returnParseError[T](fmt.Errorf("flag needs an argument: %s", arg[0]))
		}
	}

	for k, v := range passedRequiredFieldMap {
		var str string
		if !v {
			tag := typ.Field(k).Tag.Get("cliarg")
			for _, s := range splitTag(tag) {
				a, _ := parseKeyValue(s)
				if a[0] == "long" || a[0] == "short" {
					str += a[1] + " "
				}
			}
			return result, positionalArg, fmt.Errorf("required arguments were not passed: %s", str)
		}
	}

	return result, positionalArg, nil
}

func PrintHelp[T any]() {
	var d T
	typ := reflect.ValueOf(&d).Elem().Type()
	for i := range typ.NumField() {
		tag := typ.Field(i).Tag.Get("cliarg")
		if tag == "" {
			continue
		}
		for _, tagField := range splitTag(tag) {
			a := splitByEqual(tagField)
			if a[0] == "help" && len(a) == 2 {
				help := strings.Replace(a[1], `\\t`, "\t", -1)
				help = strings.Replace(help, `\\n`, "\n", -1)
				fmt.Println(help)
			}
		}
	}
}

func returnParseError[T any](err error) (T, []string, error) {
	var d T
	return d, []string{}, err
}

func setDefaultValue(info *fieldInfo, value *reflect.Value, d string) error {
	switch info.Type {
	case reflect.String:
		value.SetString(d)
	case reflect.Bool:
		if d == "true" {
			value.SetBool(true)
			info.DefaultBool = true
		}
	case reflect.Int:
		iv, err := strconv.Atoi(d)
		if err != nil {
			return err
		}
		value.SetInt(int64(iv))
	case reflect.Uint:
		uv, err := strconv.ParseUint(d, 10, 64)
		if err != nil {
			return err
		}
		value.SetUint(uv)
	}
	return nil
}

func setValue(info fieldInfo, value *reflect.Value, s, flagName string) error {
	switch info.Type {
	case reflect.String:
		value.SetString(s)
	case reflect.Int:
		iv, err := strconv.Atoi(s)
		if err != nil {
			return fmt.Errorf("invalid value \"%s\" for flag: %s", s, flagName)
		}
		value.SetInt(int64(iv))
	case reflect.Uint:
		uv, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid value \"%s\" for flag: %s", s, flagName)
		}
		value.SetUint(uv)
	}
	return nil
}

func parseKeyValue(s string) ([]string, error) {
	arr := splitByEqual(s)

	if len(arr[0]) == 0 {
		return []string{}, continuousSemicolonError()
	}

	if !slices.Contains(tagKeys, arr[0]) {
		return []string{}, noSuchKeyError(arr[0])
	}

	switch arr[0] {
	case "short":
		if !strings.HasPrefix(arr[1], "-") {
			arr[1] = "-" + arr[1]
		}
	case "long":
		if !strings.HasPrefix(arr[1], "--") {
			arr[1] = "--" + arr[1]
		}
	}

	return arr, nil
}

func splitByEqual(s string) []string {
	index := strings.Index(s, "=")
	if index == -1 {
		return []string{s}
	}
	return []string{
		s[:index],
		strings.Trim(s[index+1:], "'"),
	}
}

func splitTag(s string) []string {
	indices := make([]int, 0, 5)
	var isInside bool
	length := len(s)
	for i := 0; i < length; i++ {
		switch s[i] {
		case '\'':
			isInside = !isInside
		case ';':
			if !isInside {
				indices = append(indices, i)
			}
		}
	}

	if len(indices) == 0 {
		return []string{s}
	}

	result := make([]string, 0, 5)
	var current int
	for _, e := range indices {
		result = append(result, strings.TrimSpace(s[current:e]))
		current = e + 1
	}
	if current != length {
		result = append(result, strings.TrimSpace(s[current:length]))
	}
	return result
}
