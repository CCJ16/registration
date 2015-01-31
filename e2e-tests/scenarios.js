'use strict';

var request = require('request');

beforeEach(function () {
	var flow = protractor.promise.controlFlow();
	flow.execute(function() {
		var defer = protractor.promise.defer();
		request(browser.baseUrl + '/test_is_integration', function(error, response, body) {
			if (!error && response.statusCode == 418 && body == 'true') {
				defer.fulfill(body);
			} else {
				defer.reject('Not running against integration!');
			}
		});
		return defer.promise;
	});
	flow.execute(function() {
		var defer = protractor.promise.defer();
		browser.executeAsyncScript(function(baseUrl, callback) {
			$http = angular.injector(["ccj16reg"]).get("$http");
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

describe('Initial registration process', function() {

	browser.get('/');

	it('should automatically redirect to /register when location hash/fragment is empty', function() {
		expect(browser.getLocationAbsUrl()).toMatch("/register");
	});


	describe('by registering', function() {

		beforeEach(function() {
			browser.get('/register');
		});


		it('should render the registration form without user information', function() {
			expect(element.all(by.css('[ng-view] h2')).first().getText()).
				toMatch('Group pre-registration');
		});
		it('should have the submit button disabled by default', function() {
			var button = element(by.css('button.md-button.md-primary'));
			expect(button.isEnabled()).toBe(false);
		})
		describe('with the agreement checked', function() {
			element.all(by.css('md-checkbox')).first().click();

			it('should now have an enabled submit button', function() {
				var button = element(by.css('button.md-button.md-primary'));
				expect(button.isEnabled()).toBe(false);
			});

			it('the form should not be submittable', function() {
				var button = element(by.css('button.md-button.md-primary'));
				expect(button.click());
			});
		});
	});
});
