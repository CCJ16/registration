"use strict";

module.exports = function(config){
	if (!process.env.SAUCE_USERNAME || !process.env.SAUCE_ACCESS_KEY) {
		console.log("Make sure the SAUCE_USERNAME and SAUCE_ACCESS_KEY environment variables are set.");
		process.exit(1);
	}
	var customLaunchers = {
		"SL_Chrome": {
			base: "SauceLabs",
			browserName: "chrome",
			version: "46",
		},
		"SL_Firefox": {
			base: "SauceLabs",
			browserName: "firefox",
			version: "42",
		},
		"SL_IE_11": {
			base: "SauceLabs",
			browserName: "internet explorer",
			version: 11,
		},
	};

	config.set({
		basePath : "./",

		files : [
			"app/bower_components/angular/angular.js",
			"app/bower_components/angular-*/angular-*.js",
			"app/bower_components/moment/moment.js",
			"app/bower_components/moment-timezone/builds/moment-timezone-with-data.js",
			"app/components/**/*.js",
			"app/views/**/*.js"
		],

		singleRun : true,

		frameworks: ["jasmine"],

		browsers: Object.keys(customLaunchers),
		customLaunchers: customLaunchers,
		sauceLabs: {
			testName: "CCJ16 Registration (unit)",
			recordScreenshots: false,
			connectOptions: {
				port: 5757,
				logfile: "sauce_connect.log"
			}
		},
		captureTimeout: 0,
		browserDisconnectTimeout: 10000,
		browserDisconnectTolerance: 2,
		browserNoActivityTimeout: 30000,

		plugins : [
			"karma-sauce-launcher",
			"karma-jasmine",
			"karma-junit-reporter"
		],

		reporters: ["progress", "saucelabs"],
		junitReporter : {
			outputFile: "test_out/unit.xml",
			suite: "unit"
		}
	});

	if (process.env.TRAVIS) {
		var buildLabel = "TRAVIS #" + process.env.TRAVIS_BUILD_NUMBER + " (" + process.env.TRAVIS_BUILD_ID + ")";

		// Karma (with socket.io 1.x) buffers by 50 and 50 tests can take a long time on IEs;-)
		config.browserNoActivityTimeout = 120000;

		config.sauceLabs.build = buildLabel;
		config.sauceLabs.startConnect = false;
		config.sauceLabs.tunnelIdentifier = process.env.TRAVIS_JOB_NUMBER;
	}
};
