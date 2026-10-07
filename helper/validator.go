package helper

import (
	"errors"
	"reflect"
	"strings"
	"unicode"

	"github.com/Alan-00280/piibl-e-commerce.git/app/model"
	"github.com/go-playground/validator/v10"
)

type AppValidator struct {
	validate          *validator.Validate
	passwordCommonSet *PasswordCommonSet
}

func NewValidator(passwordCommonSet *PasswordCommonSet) *AppValidator {
	v := validator.New()

	v.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]

		if name == "" || name == "-" {
			return field.Name
		}

		return name
	})

	// nospace --> cek space / tab / new-line / return
	_ = v.RegisterValidation("nospace", func(fl validator.FieldLevel) bool {
		return !strings.Contains(fl.Field().String(), " \t\n\r")
	})

	// username --> cek hanya berupa letter / nomor / titik / underscore
	_ = v.RegisterValidation("username", func(fl validator.FieldLevel) bool {
		for _, r := range fl.Field().String() {
			if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '.' && r != '_' {
				return false
			}
		}

		return true
	})

	// role --> pengecekan role harus berupa enum CUSTOMER | TENANT
	_ = v.RegisterValidation("role", func(fl validator.FieldLevel) bool {
		return fl.Field().String() == string(model.RoleCustomer) || fl.Field().String() == string(model.RoleTenant)
	})

	_ = v.RegisterValidation("strongpassword", func(fl validator.FieldLevel) bool {
		return checkPasswordStrength(fl.Field().String(), *passwordCommonSet) == ""
	})

	return &AppValidator{
		validate:          v,
		passwordCommonSet: passwordCommonSet,
	}
}

func messageFor(fe validator.FieldError, appValidator AppValidator) string {
	switch fe.Tag() {
	case "required":
		return "wajib diisi"
	case "email":
		return "format email tidak valid"
	case "min":
		if fe.Kind() == reflect.String {
			return "minimal " + fe.Param() + " karakter"
		}
		return "minimal " + fe.Param()
	case "max":
		if fe.Kind() == reflect.String {
			return "maksimal " + fe.Param() + " karakter"
		}
		return "maksimal " + fe.Param()
	case "alphanum":
		return "wajib berisi huruf dan angka"
	case "nospace":
		return "tidak boleh berisi spasi"
	case "username":
		return "hanya boleh huruf, angka, titik, garis bawah"
	case "strongpassword":
		if value, ok := fe.Value().(string); ok {
			return checkPasswordStrength(value, *appValidator.passwordCommonSet)
		}
		return "password tidak memenuhi syarat"
	case "oneof":
		return "harus salah satu dari " + strings.ReplaceAll(fe.Param(), " ", ", ")
	case "len":
		if fe.Kind() == reflect.String {
			return "panjang harus tepat " + fe.Param() + " karakter"
		}
		return "panjang harus tepat " + fe.Param()
	case "numeric":
		return "karakter harus berupa angka"
	case "role":
		return "role tidak valid antara CUSTOMER | TENANT"
	default:
		return "tidak memenuhi aturan " + fe.Tag()
	}
}

func ValidateStruct(s any, appValidator AppValidator) map[string]string {
	validate := appValidator.validate

	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	var invalid *validator.InvalidValidationError
	if errors.As(err, &invalid) {
		return map[string]string{"_": "objek yang divalidasi tidak sah"}
	}

	var fieldErrors validator.ValidationErrors
	if !errors.As(err, &fieldErrors) {
		return map[string]string{"_": "validasi gagal"}
	}

	result := make(map[string]string, len(fieldErrors))
	for _, fe := range fieldErrors {
		if _, exists := result[fe.Field()]; !exists {
			result[fe.Field()] = messageFor(fe, appValidator)
		}
	}

	return result
}

func checkPasswordStrength(password string, pwCommon PasswordCommonSet) string {
	if len(password) < minPasswordLength {
		return "minimum 8 character of password"
	}

	var hasLetter, hasDigit bool = false, false
	for _, c := range password {
		switch {
		case unicode.IsLetter(c):
			hasLetter = true
		case unicode.IsDigit(c):
			hasDigit = true
		}
	}

	if !hasLetter || !hasDigit {
		return "passsword must contain mix of letter and numbers"
	}

	commonSet := pwCommon.PasswordSet
	if _, exists := commonSet[password]; exists {
		return "password's too common"
	}

	return ""

}
