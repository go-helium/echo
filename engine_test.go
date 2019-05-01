package echo

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/im-kulikov/helium/module"
	"github.com/labstack/echo/v4"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"go.uber.org/dig"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func newTestViper() *viper.Viper {
	v := viper.New()
	v.SetDefault("api.debug", true)
	return v
}

func jsonUnmarshalError() error {
	return &json.UnmarshalTypeError{
		Value:  "",
		Type:   reflect.TypeOf("something"),
		Offset: 0,
		Struct: "",
		Field:  "",
	}
}

func jsonSyntaxError() error {
	return &json.SyntaxError{
		Offset: 50,
	}
}

func xmlUnsuportTypeError() error {
	return &xml.UnsupportedTypeError{
		Type: reflect.TypeOf("something"),
	}
}

func xmlSyntaxError() error {
	return &xml.SyntaxError{}
}

type myTestCustomError struct {
	Message string
	Fail    bool
}

func (e myTestCustomError) Error() string {
	return e.Message
}

func (e myTestCustomError) FormatResponse(ctx echo.Context) error {
	if e.Fail {
		return errors.New(e.Message)
	}
	return ctx.String(http.StatusBadRequest, e.Message)
}

func customError() error {
	return myTestCustomError{Message: "this is custom error"}
}

func customErrorFail() error {
	return myTestCustomError{
		Message: "this is custom error",
		Fail:    true,
	}
}

func httpError() error {
	return echo.NewHTTPError(http.StatusBadRequest, "some bad request")
}

func httpInternalError() error {
	return echo.NewHTTPError(http.StatusInternalServerError, "some internal error")
}

func unknownError() error {
	return errors.New("unknown error")
}

type testBuffer struct {
	*bytes.Buffer
}

func (testBuffer) Sync() error {
	return nil
}

type respWriter struct{}

func newTestWriter() http.ResponseWriter {
	return &respWriter{}
}

func (*respWriter) Header() http.Header {
	return make(http.Header)
}

func (*respWriter) Write([]byte) (int, error) {
	return 0, errors.New("respWriter error")
}

func (*respWriter) WriteHeader(statusCode int) {}

func newTestLogger(rw zapcore.WriteSyncer) *zap.Logger {
	encoderCfg := zapcore.EncoderConfig{
		MessageKey:     "msg",
		LevelKey:       "level",
		NameKey:        "logger",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
	}

	core := zapcore.NewCore(zapcore.NewJSONEncoder(encoderCfg), rw, zap.DebugLevel)
	return zap.New(core)
}

func TestLoggerMiddleware(t *testing.T) {
	var (
		err error

		va  = NewValidator()
		b   = NewBinder(va)
		vi  = newTestViper()
		dic = dig.New()
	)

	err = module.Provide(dic, module.Module{
		{Constructor: NewEngine},
		{Constructor: func() echo.Binder { return b }},
		{Constructor: func() *viper.Viper { return vi }},
		{Constructor: func() echo.Validator { return va }},
		{Constructor: func() *zap.Logger { return zap.L() }},
	})

	require.NoError(t, err)

	err = dic.Invoke(func(e *echo.Echo, z *zap.Logger) {
		require.True(t, e.Debug)
		require.Equal(t, b, e.Binder)
		require.Equal(t, va, e.Validator)
		require.IsType(t, (*echoLogger)(nil), e.Logger)

		e.Use(LoggerMiddleware(z))

		z.Warn("test")

		{ // Fail
			req, err := http.NewRequest(echo.GET, "http://localhost", nil)
			require.NoError(t, err)

			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
		}

		{ // OK
			req, err := http.NewRequest(echo.GET, "http://localhost", nil)
			require.NoError(t, err)

			rec := httptest.NewRecorder()

			e.GET("/", func(ctx echo.Context) error { return ctx.String(http.StatusOK, "OK") })
			e.ServeHTTP(rec, req)
		}
	})

	require.NoError(t, err)
}

func TestEngine(t *testing.T) {
	Convey("Engine", t, func(c C) {
		c.Convey("try create and check new engine", func(c C) {
			var err error

			log, err := zap.NewDevelopment()
			c.So(err, ShouldBeNil)

			var (
				va = NewValidator()
				b  = NewBinder(va)
				z  = log // zap.L()
				l  = NewLogger(z)
				vi = newTestViper()

				_   = vi
				dic = dig.New()
			)

			c.Convey("create engine with empty params", func(c C) {
				err = module.Provide(dic, module.Module{
					{Constructor: NewEngine},
				})

				c.So(err, ShouldBeNil)

				err = dic.Invoke(func(e *echo.Echo) {
					c.So(e.Binder, ShouldNotEqual, b)
					c.So(e.Logger, ShouldNotEqual, l)
					c.So(e.Validator, ShouldNotEqual, va)
					c.So(e.Debug, ShouldBeFalse)
				})

				c.So(err, ShouldBeNil)
			})

			c.Convey("create engine with binder, zap.logger, validate and debug", func(c C) {
				err = module.Provide(dic, module.Module{
					{Constructor: func() echo.Validator { return va }},
					{Constructor: func() echo.Binder { return b }},
					{Constructor: func() *zap.Logger { return z }},
					{Constructor: func() *viper.Viper { return vi }},
					{Constructor: NewEngine},
				})

				c.So(err, ShouldBeNil)

				err = dic.Invoke(func(e *echo.Echo) {
					c.So(e.Binder, ShouldEqual, b)
					_, ok := e.Logger.(*echoLogger)
					c.So(ok, ShouldBeTrue)
					c.So(e.Validator, ShouldEqual, va)
					c.So(e.Debug, ShouldBeTrue)
				})

				c.So(err, ShouldBeNil)
			})

			c.Convey("create engine with binder, logger, validate and debug", func(c C) {
				err = module.Provide(dic, module.Module{
					{Constructor: func() echo.Validator { return va }},
					{Constructor: func() echo.Binder { return b }},
					{Constructor: func() echo.Logger { return l }},
					{Constructor: func() *viper.Viper { return vi }},
					{Constructor: NewEngine},
				})

				c.So(err, ShouldBeNil)

				err = dic.Invoke(func(e *echo.Echo) {
					c.So(e.Binder, ShouldEqual, b)
					c.So(e.Logger, ShouldEqual, l)
					c.So(e.Validator, ShouldEqual, va)
					c.So(e.Debug, ShouldBeTrue)
				})

				c.So(err, ShouldBeNil)
			})
		})

		c.Convey("try to capture errors", func(c C) {
			var (
				buf      = new(bytes.Buffer)
				z        = newTestLogger(testBuffer{Buffer: buf})
				e        = NewEngine(EngineParams{Logger: z})
				req, err = http.NewRequest("POST", "/some-url", nil)
				rec      = httptest.NewRecorder()
				ctx      = e.NewContext(req, rec)
			)

			c.Convey("try to capture json.Unmarshal errors", func(c C) {
				c.So(err, ShouldBeNil)
				err = jsonUnmarshalError()
				c.So(err, ShouldBeError)
				captureError(z)(err, ctx)
				c.So(rec.Body.Len(), ShouldBeGreaterThan, 0)
				c.So(rec.Body.String(), ShouldContainSubstring, "JSON parse error: expected=")
			})

			c.Convey("try to capture json.Syntax errors", func(c C) {
				c.So(err, ShouldBeNil)
				err = jsonSyntaxError()
				c.So(err, ShouldBeError)
				captureError(z)(err, ctx)
				c.So(rec.Body.Len(), ShouldBeGreaterThan, 0)
				c.So(rec.Body.String(), ShouldContainSubstring, "JSON parse error: offset=")
			})

			c.Convey("try to capture xml.Unmarshal errors", func(c C) {
				c.So(err, ShouldBeNil)
				err = xmlUnsuportTypeError()
				c.So(err, ShouldBeError)
				captureError(z)(err, ctx)
				c.So(rec.Body.Len(), ShouldBeGreaterThan, 0)
				c.So(rec.Body.String(), ShouldContainSubstring, "XML parse error: type=")
			})

			c.Convey("try to capture xml.Syntax errors", func(c C) {
				c.So(err, ShouldBeNil)
				err = xmlSyntaxError()
				c.So(err, ShouldBeError)
				captureError(z)(err, ctx)
				c.So(rec.Body.Len(), ShouldBeGreaterThan, 0)
				c.So(rec.Body.String(), ShouldContainSubstring, "XML parse error: line=")
			})

			c.Convey("try to capture custom errors", func(c C) {
				c.So(err, ShouldBeNil)
				err = customError()
				c.So(err, ShouldBeError)
				captureError(z)(err, ctx)
				c.So(rec.Body.Len(), ShouldBeGreaterThan, 0)
				c.So(rec.Body.String(), ShouldContainSubstring, "this is custom error")
			})

			c.Convey("try to capture custom errors and fail", func(c C) {
				c.So(err, ShouldBeNil)
				err = customErrorFail()
				c.So(err, ShouldBeError)
				captureError(z)(err, ctx)
				c.So(rec.Body.Len(), ShouldEqual, 0)
				c.So(buf.String(), ShouldContainSubstring, "Capture error")
				c.So(buf.String(), ShouldContainSubstring, "this is custom error")
			})

			c.Convey("try to capture http.Error", func(c C) {
				c.So(err, ShouldBeNil)
				err = httpError()
				c.So(err, ShouldBeError)
				captureError(z)(err, ctx)
				c.So(rec.Body.Len(), ShouldBeGreaterThan, 0)
				c.So(rec.Body.String(), ShouldContainSubstring, "some bad request")
			})

			c.Convey("try to capture http.Error 500", func(c C) {
				c.So(err, ShouldBeNil)
				err = httpInternalError()
				c.So(err, ShouldBeError)
				captureError(z)(err, ctx)
				c.So(rec.Body.Len(), ShouldBeGreaterThan, 0)
				c.So(rec.Body.String(), ShouldContainSubstring, "some internal error")
				c.So(buf.String(), ShouldContainSubstring, "Request error")
			})

			c.Convey("try to capture unknown Error", func(c C) {
				c.So(err, ShouldBeNil)
				err = unknownError()
				c.So(err, ShouldBeError)
				captureError(z)(err, ctx)
				c.So(rec.Body.Len(), ShouldBeGreaterThan, 0)
				c.So(rec.Body.String(), ShouldContainSubstring, http.StatusText(http.StatusBadRequest))
			})

			c.Convey("try to capture ctx.JSON Error", func(c C) {
				c.So(err, ShouldBeNil)
				err = unknownError()
				c.So(err, ShouldBeError)

				ctx.Reset(
					ctx.Request(),
					echo.NewResponse(newTestWriter(), e),
				)

				captureError(z)(err, ctx)

				c.So(rec.Body.Len(), ShouldBeZeroValue)
				c.So(buf.Len(), ShouldBeGreaterThan, 0)
				c.So(buf.String(), ShouldContainSubstring, "respWriter error")
			})
		})
	})
}
