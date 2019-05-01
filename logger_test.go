package echo

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"bou.ke/monkey"
	"github.com/labstack/gommon/log"
	. "github.com/smartystreets/goconvey/convey"
)

func TestLogger(t *testing.T) {
	Convey("Try logger", t, func(c C) {
		var (
			b        = new(bytes.Buffer)
			z        = newTestLogger(testBuffer{Buffer: b})
			l        = NewLogger(z)
			exitCode = 0
		)

		monkey.Patch(os.Exit, func(code int) { exitCode = code })

		c.Convey("try nullWriter", func(c C) {
			out := l.Output()
			size, err := out.Write(make([]byte, 10))
			c.So(out, ShouldEqual, ioutil.Discard)
			c.So(size, ShouldEqual, 10)
			c.So(err, ShouldBeNil)
		})

		c.Convey("try customize", func(c C) {
			l.SetLevel(log.ERROR)  // do nothing
			l.SetOutput(os.Stdout) // do nothing
			l.SetPrefix("prefix")  // do nothing
			l.SetHeader("header")  // do nothing

			c.So(l.Prefix(), ShouldEqual, "")
			c.So(l.Level(), ShouldEqual, log.DEBUG)
		})

		c.Convey("try Print", func(c C) {
			l.Print("")
			c.So(b.String(), ShouldContainSubstring, "info")
		})
		c.Convey("try Printf", func(c C) {
			l.Printf("")
			c.So(b.String(), ShouldContainSubstring, "info")
		})
		c.Convey("try Printj", func(c C) {
			l.Printj(log.JSON{})
			c.So(b.String(), ShouldContainSubstring, "info")
		})
		c.Convey("try Debug", func(c C) {
			l.Debug("")
			c.So(b.String(), ShouldContainSubstring, "debug")
		})
		c.Convey("try Debugf", func(c C) {
			l.Debugf("")
			c.So(b.String(), ShouldContainSubstring, "debug")
		})
		c.Convey("try Debugj", func(c C) {
			l.Debugj(log.JSON{})
			c.So(b.String(), ShouldContainSubstring, "debug")
		})
		c.Convey("try Info", func(c C) {
			l.Info("")
			c.So(b.String(), ShouldContainSubstring, "info")
		})
		c.Convey("try Infof", func(c C) {
			l.Infof("")
			c.So(b.String(), ShouldContainSubstring, "info")
		})
		c.Convey("try Infoj", func(c C) {
			l.Infoj(log.JSON{})
			c.So(b.String(), ShouldContainSubstring, "info")
		})
		c.Convey("try Warn", func(c C) {
			l.Warn("")
			c.So(b.String(), ShouldContainSubstring, "warn")
		})
		c.Convey("try Warnf", func(c C) {
			l.Warnf("")
			c.So(b.String(), ShouldContainSubstring, "warn")
		})
		c.Convey("try Warnj", func(c C) {
			l.Warnj(log.JSON{})
			c.So(b.String(), ShouldContainSubstring, "warn")
		})
		c.Convey("try Error", func(c C) {
			l.Error("")
			c.So(b.String(), ShouldContainSubstring, "error")
		})
		c.Convey("try Errorf", func(c C) {
			l.Errorf("")
			c.So(b.String(), ShouldContainSubstring, "error")
		})
		c.Convey("try Errorj", func(c C) {
			l.Errorj(log.JSON{})
			c.So(b.String(), ShouldContainSubstring, "error")
		})

		c.Convey("try Fatal", func(c C) {
			l.Fatal("")
			c.So(exitCode, ShouldNotEqual, 2)
			c.So(b.String(), ShouldContainSubstring, "fatal")
		})
		c.Convey("try Fatalf", func(c C) {
			l.Fatalf("")
			c.So(exitCode, ShouldNotEqual, 2)
			c.So(b.String(), ShouldContainSubstring, "fatal")
		})
		c.Convey("try Fatalj", func(c C) {
			l.Fatalj(log.JSON{})
			c.So(exitCode, ShouldNotEqual, 2)
			c.So(b.String(), ShouldContainSubstring, "fatal")
		})

		c.Convey("try Panic", func(c C) {
			c.So(func() {
				l.Panic("")
			}, ShouldPanic)
			c.So(exitCode, ShouldNotEqual, 2)
			c.So(b.String(), ShouldContainSubstring, "panic")
		})
		c.Convey("try Panicf", func(c C) {
			c.So(func() {
				l.Panicf("")
			}, ShouldPanic)
			c.So(exitCode, ShouldNotEqual, 2)
			c.So(b.String(), ShouldContainSubstring, "panic")
		})
		c.Convey("try Panicj", func(c C) {
			c.So(func() {
				l.Panicj(log.JSON{})
			}, ShouldPanic)
			c.So(exitCode, ShouldNotEqual, 2)
			c.So(b.String(), ShouldContainSubstring, "panic")
		})
	})
}
