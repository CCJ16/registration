# Registration application for CCJ16
[![Build Status](https://travis-ci.org/CCJ16/registration.svg?branch=master)](https://travis-ci.org/CCJ16/registration)

This project contains the entire registration system for CCJ16.  This project can be used for other registration systems if wanted.

## Requirements

All requirements are available using the given language's package manager.  Currently this includes:

### Golang
 - [BoltDB](https://github.com/CCJ16/registration/regbackend)
 - [Gorilla toolkit](http://www.gorillatoolkit.org/)
 - [Space Monkey Go errors](https://github.com/spacemonkeygo/errors)
 - [GoConvey](https://github.com/smartystreets/goconvey)

### Javascript
 - [AngularJS](http://angularjs.org)
 - [Karma](http://karma-runner.github.io/0.12/index.html)
 - [Jasmine](http://jasmine.github.io/)
 - [HTML5 Boilerplate](https://html5boilerplate.com/)
 - [Material Design Icon](https://github.com/google/material-design-icons)

## Installing

1. Download this project into your GOPATH: `go get github.com/CCJ16/registration`
2. Change into the repository location.
3. Install the project dependencies for the client side: `npm install`.

## Running

The project is currently missing a configuration file, so main.go will need adjusting.  Modify the relevant strings as desired.  To run the project, run in the repository location:

1. Change to the go folder `cd regbackend`
2. Build the go project again in the current folder: `go build`
3. Run the application in the current folder.  Note the path to the html/js is hardcoded, so run the command in the current folder: `./regbackend`

The executable will log the http requests and some errors.  You can run the server directly against the Internet, or proxied through another web server like lighttpd or nginx.

## Tests

To run the server side tests, run `go test ./...` in the repository folder.  To run the client side tests, run `npm test` in the repository folder.

## Contributing

Bug reports are appreciated.  Pull requests for any code, documentation, or other changes are appreciated.
