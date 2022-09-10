package validation

import (
	"bitbucket.org/itskovich/goava/pkg/goava/errs"
	"bitbucket.org/itskovich/goava/pkg/goava/utils"
	"fmt"
	"github.com/spf13/cast"
	"net/mail"
	"reflect"
	"regexp"
	"strings"
	"time"
)

const (
	Empty          = "EMPTY"
	Null           = "NULL"
	InvalidInt     = "INVALID_INT"
	InvalidInt64   = "INVALID_LONG"
	InvalidFloat   = "INVALID_FLOAT"
	InvalidDate    = "INVALID_DATE"
	ViolatesRegexp = "VIOLATES_REGEXP"
	Unexpectable   = "UNEXPECTABLE"
	InvalidLength  = "INVALID_LENGTH"
	InvalidEmail   = "INVALID_EMAIL"
	InvalidBoolean = "INVALID_BOOLEAN"
)

type ValidationError struct {
	errs.BaseError

	Reason       string
	Param        string
	InvalidValue interface{}
}

func CheckInt64(param string, v interface{}) (int64, error) {
	r, err := CheckCondition(func() (interface{}, bool) {
		r, e := cast.ToInt64E(v)
		return r, e == nil
	}, param, InvalidInt64, v, func() string {
		return "Параметр должен быть целым числом"
	})

	if err != nil {
		return 0, err
	}
	return cast.ToInt64(r), nil
}

func CheckFloat32(param string, v interface{}) (float32, error) {
	r, err := CheckCondition(func() (interface{}, bool) {
		r, e := cast.ToFloat32E(v)
		return r, e == nil
	}, param, InvalidInt, v, func() string {
		return "Параметр должен быть вещественным числом"
	})

	if err != nil {
		return 0, err
	}
	return cast.ToFloat32(r), nil
}

func CheckFloat64(param string, v interface{}) (float64, error) {
	r, err := CheckCondition(func() (interface{}, bool) {
		r, e := cast.ToFloat64E(v)
		return r, e == nil
	}, param, InvalidInt, v, func() string {
		return "Параметр должен быть вещественным числом"
	})

	if err != nil {
		return 0, err
	}
	return cast.ToFloat64(r), nil
}

func CheckInt(param string, v interface{}) (int, error) {
	r, err := CheckCondition(func() (interface{}, bool) {
		r, e := cast.ToIntE(v)
		return r, e == nil
	}, param, InvalidInt, v, func() string {
		return "Параметр должен быть целым числом"
	})

	if err != nil {
		return 0, err
	}
	return cast.ToInt(r), nil
}

func CheckCondition(condition func() (interface{}, bool), param string, reason string, value interface{}, errMsgProvider func() string) (interface{}, error) {
	res, ok := condition()
	if !ok {
		return res, &ValidationError{
			BaseError:    *errs.NewBaseError(errMsgProvider()),
			Reason:       reason,
			Param:        param,
			InvalidValue: value,
		}
	}
	return res, nil
}

func CheckNotEmptyStr(param string, v string) (string, error) {
	r, err := CheckCondition(func() (interface{}, bool) {
		return v, len(v) > 0
	}, param, Empty, v, func() string {
		return "Параметр должен быть непустой строкой"
	})

	if err != nil {
		return "", err
	}
	return r.(string), nil
}

func CheckNotEmpty(param string, value interface{}) (interface{}, error) {
	return CheckCondition(func() (interface{}, bool) {
		switch v := value.(type) {
		case string:
			return v, len(v) > 0
		}
		return value, value != nil
	}, param, Empty, value, func() string {
		return "Параметр не должен быть пустым"
	})
}

func CheckBool(param string, value interface{}) (bool, error) {
	r, err := CheckCondition(func() (interface{}, bool) {
		b, err := cast.ToBoolE(value)
		return b, err == nil
	}, param, InvalidBoolean, value, func() string {
		return "Параметр должен иметь значение true/false"
	})

	return cast.ToBool(r), err
}

func CheckMatchRegexp(param string, v string, pattern string) (string, error) {
	r, err := CheckCondition(func() (interface{}, bool) {
		matches, _ := regexp.MatchString(pattern, v)
		return v, matches
	}, param, ViolatesRegexp, v, func() string {
		return "Параметр не соответствует формату"
	})

	if err != nil {
		return "", err
	}
	return cast.ToString(r), nil
}

func CheckDate(param string, v string) (*time.Time, error) {
	r, err := CheckCondition(func() (interface{}, bool) {
		r, err := time.Parse("02.01.2006", v)
		if err != nil {
			r, err = time.Parse("2006-01-02", v)
		}
		return r, err == nil
	}, param, InvalidDate, v, func() string {
		return ""
	})

	if err != nil {
		return nil, err
	}
	rr := r.(time.Time)
	return &rr, nil
}

func CheckMaxLen(param string, v string, maxLen int) (interface{}, error) {
	return CheckCondition(func() (interface{}, bool) {
		return v, len(v) <= maxLen
	}, param, InvalidLength, v, func() string {
		return fmt.Sprintf("Длина строки превышает %v символов", maxLen)
	})
}

func CheckMinLen(param string, v string, minLen int) (interface{}, error) {
	return CheckCondition(func() (interface{}, bool) {
		return v, len(v) >= minLen
	}, param, InvalidLength, v, func() string {
		return fmt.Sprintf("Длина строки меньше %v символов", minLen)
	})
}

func CheckEmail(param string, v string) (interface{}, error) {
	return CheckCondition(func() (interface{}, bool) {
		addr, err := mail.ParseAddress(cast.ToString(v))
		return addr, err == nil
	}, param, InvalidEmail, v, func() string {
		return "Некорректный e-mail"
	})
}

func CheckFirst(param string, a interface{}) error {
	r := Check(param, a)
	if len(r) > 0 {
		return r[0]
	}
	return nil
}

func Check(param string, a interface{}) []error {
	var checkErrors []error
	t := reflect.TypeOf(a).Elem()
	v := reflect.ValueOf(a).Elem()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		fv := v.Field(i)
		checksStr := f.Tag.Get("check")
		if len(checksStr) == 0 {
			continue
		}
		checks := strings.Split(checksStr, ",")
		if len(v.String()) == 0 && !utils.ContainsStr(checks, "notempty", true) {
			continue
		}
		for _, checkType := range checks {
			checkType = strings.ToLower(checkType)
			err := checkField(f, checkType, param, fv)
			if err != nil {
				checkErrors = append(checkErrors, err)
			}
		}
	}
	return checkErrors
}

func checkField(f reflect.StructField, checkType string, param string, fv reflect.Value) error {
	param += "." + f.Name
	var err error
	switch checkType {
	case "email":
		_, err = CheckEmail(param, fv.String())
		break
	case "notempty":
		_, err = CheckNotEmpty(param, fv.Interface())
		break
	}
	return err
}
