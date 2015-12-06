"use strict";

var request = require("request");

exports.config = {
	sauceUser: process.env.SAUCE_USERNAME,
	sauceKey: process.env.SAUCE_ACCESS_KEY,

	allScriptsTimeout: 11000,

	specs: [
		"e2e-tests/*.js"
	],

	multiCapabilities: [
		addCommon({browserName: "chrome", version: 46}),
		addCommon({browserName: "firefox", version: 42}),
		addCommon({browserName: "internet explorer", version: 11}),
		addCommon({browserName: "safari", version: "9.0", platform: "OS X 10.11"}),
	],

	baseUrl: "http://localhost:9090",
	onPrepare: function() {
		// Disable animations so e2e tests run more quickly
		browser.addMockModule("disableNgAnimate", function() {
			angular.module("disableNgAnimate", []).run(["$animate", function($animate) {
				$animate.enabled(false);
			}]);
		});

		var defer = protractor.promise.defer();
		request(browser.baseUrl + "/test_is_integration", function(error, response, body) {
			if (!error && response.statusCode === 418 && body === "true") {
				defer.fulfill(body);
			} else {
				defer.reject("Not running against integration!");
			}
		});
		return defer.promise;
	},

	framework: "jasmine2",

	jasmineNodeOpts: {
		defaultTimeoutInterval: 60000
	}
};
function addCommon(capabilities) {
	var buildLabel = "TRAVIS #" + process.env.TRAVIS_BUILD_NUMBER + " (" + process.env.TRAVIS_BUILD_ID + ")";
	if (process.env.TRAVIS) {
		return {
			"tunnel-identifier": process.env.TRAVIS_JOB_NUMBER,

			"name": "CCJ16 Registration (e2e)",
			"build": buildLabel,

			"browserName": capabilities.browserName,
			"platform": capabilities.platform,
			"version": capabilities.version,
			"device-orientation": capabilities["device-orientation"],
		};
	} else {
		return {
			"browserName": capabilities.browserName,
			"platform": capabilities.platform,
			"version": capabilities.version,
			"device-orientation": capabilities["device-orientation"],
		};
	}
}
