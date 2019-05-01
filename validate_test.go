package echo

import (
	"reflect"
	"testing"

	"github.com/labstack/echo/v4"
	. "github.com/smartystreets/goconvey/convey"
)

func someValidatorFunc(fl FieldLevel) bool {
	field := fl.Field()

	switch field.Kind() {
	case reflect.String:
		return field.String() == "true"
	default:
		return false
	}
}

type someValidStruct struct {
	A string `validate:"required,some" message:"some error"`
}

type someInvalidStruct struct {
	A int `validate:"required,some" message:"some error"`
}

type someSimpleStruct struct {
	A int
}

var cases = []struct {
	item  interface{}
	error error
}{
	{
		item:  &someValidStruct{A: "true"},
		error: nil,
	},

	{
		item:  new(someSimpleStruct),
		error: nil,
	},

	{
		item:  &someValidStruct{A: "false"},
		error: echo.NewHTTPError(400, "`a` some error"),
	},

	{
		item:  &someInvalidStruct{A: 1},
		error: echo.NewHTTPError(400, "`a` some error"),
	},
}

func TestValidate_Register(t *testing.T) {
	Convey("Validator", t, func(c C) {
		var (
			v   = NewValidator()
			err = v.Register("some", someValidatorFunc)
		)

		c.So(err, ShouldBeNil)

		c.Convey("try cases", func(c C) {
			for _, test := range cases {
				err = v.Validate(test.item)
				switch test.error {
				case nil:
					c.So(err, ShouldBeNil)
				default:
					if ok, vErr := CheckErrors(ValidateParams{
						Struct: test.item,
						Errors: err,
					}); ok {
						c.So(vErr, ShouldBeError, test.error)
					} else {
						t.FailNow()
					}
				}
			}
		})
	})
}
