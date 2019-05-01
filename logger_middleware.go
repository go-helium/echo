package echo

import (
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

var loggerMiddlewareFieldNames = []string{
	"id",
	"real_ip",
	"method",
	"status",
	"proto",
	"host",
	"uri",
	"path",
	"referer",
	"agent",
	"latency",
	"bytes_in",
	"bytes_out",
}

// LoggerMiddleware returns a middleware that logs HTTP requests.
func LoggerMiddleware(log *zap.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			var (
				req   = c.Request()
				res   = c.Response()
				start = time.Now()
				err   = next(c)
				stop  = time.Now()
			)

			id := req.Header.Get(echo.HeaderXRequestID)
			if id == "" {
				id = res.Header().Get(echo.HeaderXRequestID)
			}

			p := req.URL.Path
			if p == "" {
				p = "/"
			}

			items := []string{
				id,
				c.RealIP(),
				req.Method,
				strconv.Itoa(res.Status),
				req.Proto,
				req.Host,
				req.RequestURI,
				p,
				req.Referer(),
				req.UserAgent(),
				stop.Sub(start).String(),
				req.Header.Get(echo.HeaderContentLength),
				strconv.FormatInt(res.Size, 10),
			}

			fields := make([]zap.Field, 0, len(items))
			for i := 0; i < len(loggerMiddlewareFieldNames); i++ {
				if items[i] != "" {
					fields = append(fields,
						zap.String(loggerMiddlewareFieldNames[i], items[i]))
				}
			}

			fields = append(fields, zap.Error(err))
			log.Info("new request", fields...)

			return err
		}
	}
}
