"use strict";

var request = require("request");

exports.config = {
	allScriptsTimeout: 11000,

	specs: [
		"e2e-tests/*.js"
	],

	directConnect: true,
	capabilities: {
		"browserName": "chrome"
	},

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
		defaultTimeoutInterval: 30000
	}
};
