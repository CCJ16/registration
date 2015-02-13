"use strict";

var request = require("request");

beforeEach(function () {
	var flow = protractor.promise.controlFlow();
	flow.execute(function() {
		var defer = protractor.promise.defer();
		request(browser.baseUrl + "/test_is_integration", function(error, response, body) {
			if (!error && response.statusCode === 418 && body === "true") {
				defer.fulfill(body);
			} else {
				defer.reject("Not running against integration!");
			}
		});
		return defer.promise;
	});
	flow.execute(function() {
		var defer = protractor.promise.defer();
		browser.executeAsyncScript(function(baseUrl, callback) {
			var $http = angular.injector(["ccj16reg"]).get("$http");
			$http({ url: baseUrl + "/integration/wipe_database" }).then(function() {
				callback(true)
			}, function () {
				callback(false)
			});
		}, browser.baseUrl).then(function(success) {
			if (success) {
				defer.fulfill();
			} else {
				defer.reject("Failed to clear database");
			}
		});
		return defer.promise;
	});
});
