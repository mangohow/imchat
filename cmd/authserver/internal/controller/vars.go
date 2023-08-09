package controller

import (
	"github.com/go-playground/validator/v10"
	"github.com/mangohow/imchat/pkg/utils"
)

var validate = validator.New()

func init() {
	validate.RegisterValidation("phone", func(fl validator.FieldLevel) bool {
		return utils.ValidatePhoneNumber(fl.Field().String())
	}, false)
}