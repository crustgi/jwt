[![GoDoc](https://godoc.org/github.com/pascaldekloe/jwt?status.svg)](https://godoc.org/github.com/pascaldekloe/jwt)
[![Build Status](https://travis-ci.org/pascaldekloe/jwt.svg?branch=master)](https://travis-ci.org/pascaldekloe/jwt)
[![Coverage Status](https://coveralls.io/repos/github/pascaldekloe/jwt/badge.svg?branch=master)](https://coveralls.io/github/pascaldekloe/jwt)

[![JWT.io](http://jwt.io/img/badge.svg)](https://jwt.io/)

Lighteight JSON Web Token (JWT) library for the Go programming language.

* Small API with lightweight implementation
* No third party dependencies
* No support for (ECDSA) encryption

```
goos: darwin
goarch: amd64
pkg: github.com/pascaldekloe/jwt
BenchmarkHMACCheck-12    	  100000	     11822 ns/op
BenchmarkRSACheck-12     	   20000	     74142 ns/op
BenchmarkHMACSign-12     	  500000	      3680 ns/op
BenchmarkRSASign-12      	     500	   2579161 ns/op
PASS
ok  	github.com/pascaldekloe/jwt	7.004s
```

This is free and unencumbered software released into the
[public domain](http://creativecommons.org/publicdomain/zero/1.0).
