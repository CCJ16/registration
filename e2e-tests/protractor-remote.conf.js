'use strict';

var request = require('request');

exports.config = {
	sauceUser: process.env.SAUCE_USERNAME,
	sauceKey: process.env.SAUCE_ACCESS_KEY,

	allScriptsTimeout: 11000,

	specs: [
		'*.js'
	],

	multiCapabilities: [
		addCommon({browserName: 'chrome', version: 39}),
		addCommon({browserName: 'firefox', version: 34}),
		addCommon({browserName: 'internet explorer', version: 11}),
		addCommon({browserName: 'internet explorer', version: 10}),
		addCommon({browserName: 'safari', version: '8', platform: 'OS X 10.10'}),
		addCommon({browserName: 'iphone', version: '8.1', platform: 'OS X 10.10', 'device-orientation': 'portrait'}),
	],

	baseUrl: 'http://localhost:9090',
	onPrepare: function(data) {
		var defer = protractor.promise.defer();
		request(browser.baseUrl + '/test_is_integration', function(error, response, body) {
			if (!error && response.statusCode == 418 && body == 'true') {
				defer.fulfill(body);
			} else {
				defer.reject('Not running against integration!');
			}
		});
		return defer.promise;
	},

	framework: 'jasmine',

	jasmineNodeOpts: {
		defaultTimeoutInterval: 30000
	}
};
function addCommon(capabilities) {
	var buildLabel = 'TRAVIS #' + process.env.TRAVIS_JOB_NUMBER + ' (' + process.env.TRAVIS_JOB_ID + ')';
	if (process.env.TRAVIS) {
		return {
			'tunnel-identifier': process.env.TRAVIS_JOB_NUMBER,

			'name': 'CCJ16 Registration (e2e)',
			'build': buildLabel,

			'browserName': capabilities.browserName,
			'platform': capabilities.platform,
			'version': capabilities.version,
			'device-orientation': capabilities['device-orientation'],
		};
	} else {
		return {
			'browserName': capabilities.browserName,
			'platform': capabilities.platform,
			'version': capabilities.version,
			'device-orientation': capabilities['device-orientation'],
		};
	}
}
