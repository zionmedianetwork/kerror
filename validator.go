package kerror

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"

	_ "github.com/go-playground/locales"
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	enTranslations "github.com/go-playground/validator/v10/translations/en"
)

type FieldValidationError struct {
	FieldName string `json:"field"`
	Error     error  `json:"message"`
}

func (f FieldValidationError) MarshalJSON() ([]byte, error) {
	var errMessage string

	if f.Error != nil {
		errMessage = f.Error.Error()
	}

	temp := struct {
		FieldName string `json:"field"`
		Error     string `json:"message"`
	}{
		FieldName: f.FieldName,
		Error:     errMessage,
	}

	return json.Marshal(temp)

}

type Validator struct {
	*validator.Validate
	trans ut.Translator
}

func New() *Validator {
	validate := validator.New()

	// Register custom validation tags
	_ = validate.RegisterValidation("tel", validateTelephoneNumber)
	_ = validate.RegisterValidation("password", validatePassword)

	// Initialize englise locales
	english := en.New()
	uni := ut.New(english, english)
	trans, _ := uni.GetTranslator("en")

	// Set default translation
	_ = enTranslations.RegisterDefaultTranslations(validate, trans)

	vl := &Validator{
		validate,
		trans,
	}
	// Add translation to validation tags
	vl.addTranslation("tel", errInvalidTelephone)
	vl.addTranslation("password", errInvalidPassword)

	// Registering func to return tags
	vl.RegisterTagNameFunc(GetJSONTag)

	return vl
}

func (v *Validator) VStruct(s interface{}) []FieldValidationError {
	validateStruct := v.Struct(s)
	return v.translateError(validateStruct)
}

func validateTelephoneNumber(fl validator.FieldLevel) bool {
	telephone := fl.Field().String()

	result, _ := RegexpPhone.MatchString(telephone)

	return result
}

func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	result, _ := RegexPassword.MatchString(password)

	return result
}

func (v Validator) addTranslation(tag string, errMessage string) {

	transFn := func(ut ut.Translator, fe validator.FieldError) string {
		param := fe.Param()
		tag := fe.Tag()

		t, err := ut.T(tag, fe.Field(), param)
		if err != nil {
			return fe.(error).Error()
		}
		return t
	}

	registerFn := func(ut ut.Translator) error {
		return ut.Add(tag, errMessage, false)
	}
	_ = v.RegisterTranslation(tag, v.trans, registerFn, transFn)
}

func (v Validator) translateError(err error) []FieldValidationError {
	var fErr []FieldValidationError

	if err == nil {
		return nil
	}
	validatorErrs := err.(validator.ValidationErrors)
	for _, e := range validatorErrs {
		translatedErr := e.Translate(v.trans)
		fErr = append(fErr, FieldValidationError{
			FieldName: e.Field(),
			Error:     errors.New(translatedErr)})
	}

	return fErr
}

// GetJSONTag will return the value of the
// json tag of the field. This func can
// be passed as argument to validate.RegisterTagNameFunc
func GetJSONTag(fld reflect.StructField) string {
	name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
	if name == "-" {
		return ""
	}
	return name
}
