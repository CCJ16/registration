'use strict';

var request = require('request');

exports.config = {
	allScriptsTimeout: 11000,

	specs: [
		'e2e-tests/*.js'
	],

	capabilities: {
		'browserName': 'chrome'
	},

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
