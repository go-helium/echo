# Echo module for Helium

![Codecov](https://img.shields.io/codecov/c/github/go-helium/echo.svg?style=flat-square)
[![Build Status](https://travis-ci.com/go-helium/echo.svg?branch=master)](https://travis-ci.com/go-helium/echo)
[![Report](https://goreportcard.com/badge/github.com/go-helium/echo)](https://goreportcard.com/report/github.com/go-helium/echo)
[![GitHub release](https://img.shields.io/github/release/go-helium/echo.svg)](https://github.com/go-helium/echo)
![GitHub](https://img.shields.io/github/license/go-helium/echo.svg?style=popout)

Module provides boilerplate that preconfigured echo.Engine for you with custom Binder / Logger / Validator / ErrorHandler

- bind - simple replacement for echo.Binder
- validate - simple replacement for echo.Validate
- logger - provides echo.Logger that pass calls to **zap.Logger**