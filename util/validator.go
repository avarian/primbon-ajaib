package util

import (
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
)

type Validator struct {
	Validate *validator.Validate
	Trans    ut.Translator
}

func ValidatorTranslate() *Validator {
	// NOTE: ommitting allot of error checking for brevity
	var uni *ut.UniversalTranslator
	var validate *validator.Validate

	en := en.New()
	uni = ut.New(en, en)

	// this is usually know or extracted from http 'Accept-Language' header
	// also see uni.FindTranslator(...)
	trans, _ := uni.GetTranslator("en")

	validate = validator.New()
	en_translations.RegisterDefaultTranslations(validate, trans)

	validator := Validator{
		Validate: validate,
		Trans:    trans,
	}
	return &validator
}
