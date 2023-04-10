package validator

import (
	"github.com/pkg/errors"
	"reflect"
	"strconv"
	"strings"
)

var ErrNotStruct = errors.New("wrong argument given, should be a struct")
var ErrInvalidValidatorSyntax = errors.New("invalid validator syntax")
var ErrValidateForUnexportedFields = errors.New("validation for unexported field is not allowed")
var ErrValidateForUnsupportedValidator = errors.New("used unsupported validator")
var ErrValidateForUnsupportedType = errors.New("used unsupported type with validator")

type ValidationError struct {
	Err error
}

type ValidationErrors []ValidationError

const (
	Max = 1
	Min = -1
)

func (v ValidationErrors) Error() string {
	resultErr := ""
	for _, err := range v {
		resultErr += err.Err.Error() + " >>= "
	}
	resultErr = strings.TrimSuffix(resultErr, " >>= ")
	return resultErr
}

func Validate(v any) error {
	varType := reflect.TypeOf(v)
	varValue := reflect.ValueOf(v)

	var errs ValidationErrors

	if varType.Kind() != reflect.Struct {
		return ErrNotStruct
	}

	for i := 0; i < varType.NumField(); i++ {
		tag := varType.Field(i).Tag.Get("validate")
		if tag == "" {
			continue
		}

		if !varType.Field(i).IsExported() {
			errs = append(errs, ValidationError{ErrValidateForUnexportedFields})
			continue
		}

		splitTag := strings.SplitN(tag, ":", 2)

		if len(splitTag) != 2 || len(splitTag[1]) == 0 {
			errs = append(errs, ValidationError{ErrInvalidValidatorSyntax})
			continue
		}
		if varValue.Field(i).Kind() != reflect.Int && varValue.Field(i).Kind() != reflect.String {
			errs = append(errs, ValidationError{ErrValidateForUnsupportedType})
			continue
		}

		tagName := splitTag[0]
		valParams := strings.Split(splitTag[1], ",")

		var err error

		switch tagName {
		case "in":
			err = ValidateIn(varValue.Field(i), valParams)
		case "max":
			err = ValidateMaxMin(varValue.Field(i), valParams, Max)
		case "min":
			err = ValidateMaxMin(varValue.Field(i), valParams, Min)
		case "len":
			err = ValidateLen(varValue.Field(i), valParams)
		default:
			err = ErrValidateForUnsupportedValidator
		}

		if err != nil {
			errs = append(errs, ValidationError{err})
		}

	}
	if len(errs) == 0 {
		return nil
	}
	return errs
}

func ValidateMaxMin(val reflect.Value, args []string, t int) error {
	if len(args) != 1 {
		return ErrInvalidValidatorSyntax
	}

	expVal, err := strconv.Atoi(args[0])
	if err != nil {
		return ErrInvalidValidatorSyntax
	}

	if val.Kind() == reflect.Int && t == Max && val.Int() > int64(expVal) {
		return errors.New("Integer " + strconv.FormatInt(val.Int(), 10) + " is greater than " + strconv.Itoa(expVal))
	} else if val.Kind() == reflect.Int && t == Min && val.Int() < int64(expVal) {
		return errors.New("Integer " + strconv.FormatInt(val.Int(), 10) + " is less than " + strconv.Itoa(expVal))
	} else if val.Kind() == reflect.String && t == Max && len(val.String()) > expVal {
		return errors.New("Length of string " + val.String() + " is greater than " + strconv.Itoa(expVal))
	} else if val.Kind() == reflect.String && t == Min && len(val.String()) < expVal {
		return errors.New("Length of string " + val.String() + " is less than " + strconv.Itoa(expVal))
	}
	return nil
}

func ValidateIn(val reflect.Value, in []string) error {
	found := false
	var err error

	if val.Kind() == reflect.Int {
		for _, v := range in {
			num, e := strconv.Atoi(v)
			if e != nil && err != nil {
				err = e
			} else if val.Int() == int64(num) {
				found = true
			}
		}
	} else if val.Kind() == reflect.String {
		for _, v := range in {
			if val.String() == v {
				found = true
			}
		}
	}
	if err != nil {
		return err
	}
	if !found {
		return errors.New("value is not found")
	}
	return nil
}

func ValidateLen(val reflect.Value, length []string) error {
	expVal, err := strconv.Atoi(length[0])
	if err != nil {
		return ErrInvalidValidatorSyntax
	}

	if val.Kind() == reflect.Int {
		return ErrValidateForUnsupportedValidator
	} else if val.Kind() == reflect.String && len(val.String()) != expVal {
		return errors.New("length of string " + val.String() + " is greater than " + strconv.Itoa(expVal))
	}
	return nil
}
