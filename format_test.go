package echo

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/labstack/echo/v4"
	. "github.com/smartystreets/goconvey/convey"
)

type test1 struct {
	A int `json:"a_custom" validate:"gt=0" message:"must be greater than 0"`
	B int `form:"b_custom" validate:"required" message:"is required"`
	C int
	D int `query:"someValue" validate:"required"`
	E int `json:"-" form:"-" query:"-" param:"-" xml:"-" yaml:"-" validate:"required"`
	F int `json:"-" param:"f_custom" xml:"-" yaml:"-" validate:"required"`
	G int `xml:"g_custom" yaml:"-" validate:"required"`
	H int `yaml:"h_custom" validate:"required"`
}

type testCase struct {
	Name   string
	Struct interface{}
	Error  error
}

var testCases = []testCase{
	{
		Name:   "validate errors for all fields",
		Struct: test1{A: -1},
		Error:  echo.NewHTTPError(http.StatusBadRequest, "bad value of `someValue`, `e`, `f_custom`, `g_custom`, `h_custom`; `a_custom` must be greater than 0; `b_custom` is required"),
	},

	{
		Name:   "validate errors for A field",
		Struct: test1{A: 0, B: 1, D: 1, E: 1, F: 1, G: 1, H: 1},
		Error:  echo.NewHTTPError(http.StatusBadRequest, "`a_custom` must be greater than 0"),
	},

	{
		Name:   "validate errors for B field",
		Struct: test1{A: 1, D: 1, E: 1, F: 1, G: 1, H: 1},
		Error:  echo.NewHTTPError(http.StatusBadRequest, "`b_custom` is required"),
	},

	{
		Name:   "validate errors for D, E, F, G, H fields",
		Struct: test1{A: 1, B: 1},
		Error:  echo.NewHTTPError(http.StatusBadRequest, "bad value of `someValue`, `e`, `f_custom`, `g_custom`, `h_custom`"),
	},

	{
		Name:   "validate errors must be empty",
		Struct: test1{A: 1, B: 1, D: 1, E: 1, F: 1, G: 1, H: 1},
		Error:  nil,
	},
}

func TestCheckErrors(t *testing.T) {
	Convey("Prepare validator", t, func(c C) {
		v := NewValidator()

		c.So(v, ShouldNotBeNil)

		c.So(func() {
			AddTagParsers(func(reflect.StructTag) string {
				return "-"
			})
		}, ShouldNotPanic)

		c.Convey("check custom formatter", func(c C) {
			test := testCases[0]
			errValidate := v.Validate(test.Struct)

			if test.Error == nil {
				c.So(errValidate, ShouldBeNil)
			} else {
				c.So(errValidate, ShouldBeError)
			}

			ok, err := CheckErrors(ValidateParams{
				Struct: test.Struct,
				Errors: errValidate,
				Formatter: func(fields []*FieldError) string {
					for _, item := range fields {
						c.So(item, ShouldNotBeNil)
						c.So(item, ShouldBeError)
					}
					return defaultFormatter(fields)
				},
			})

			if test.Error == nil {
				c.So(ok, ShouldBeFalse)
				c.So(err, ShouldBeNil)
			} else {
				c.So(ok, ShouldBeTrue)
				c.So(err, ShouldNotBeNil)
				c.So(err.Error(), ShouldEqual, test.Error.Error())
			}
		})

		c.So(len(testCases) > 0, ShouldBeTrue)
		for _, test := range testCases {
			c.Convey(test.Name, func(c C) {
				errValidate := v.Validate(test.Struct)

				if test.Error == nil {
					c.So(errValidate, ShouldBeNil)
				} else {
					c.So(errValidate, ShouldBeError)
				}

				ok, err := CheckErrors(ValidateParams{
					Struct: test.Struct,
					Errors: errValidate,
				})

				if test.Error == nil {
					c.So(ok, ShouldBeFalse)
					c.So(err, ShouldBeNil)
				} else {
					c.So(ok, ShouldBeTrue)
					c.So(err, ShouldNotBeNil)
					c.So(err.Error(), ShouldEqual, test.Error.Error())
				}
			})
		}
	})
}
