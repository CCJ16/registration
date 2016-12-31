"use strict";

describe("Registration information page", function() {
	describe("with no pack information submitted", function() {
		var savedData;
		beforeEach(function() {
			var flow = protractor.promise.controlFlow();
			flow.execute(function() {
				var defer = protractor.promise.defer();
				browser.executeAsyncScript(function(callback) {
					var Registration = angular.element(document.body).injector().get("Registration");
					var reg = new Registration();
					reg.council = "Test council";
					reg.groupName = "Test group";
					reg.contactLeaderFirstName = "FirstN";
					reg.contactLeaderLastName = "LastN";
					reg.contactLeaderEmail = "test@invalid";
					reg.contactLeaderPhoneNumber= "123-456-7890";
					reg.estimatedYouth = 20;
					reg.estimatedLeaders = 12;
					reg.agreedToEmailTerms(true);

					reg.$save().then(function(savedObject) {
						callback([true, savedObject]);
					}, function() {
						callback([false]);
					});
				}).then(function(data) {
					var success = data[0];
					if (success) {
						defer.fulfill(data[1]);
					} else {
						defer.reject("Failed to clear database");
					}
				});
				defer.promise.then(function(data) {
					savedData = data;
					browser.setLocation("/registration/" + data.securityKey);
				});
				return defer.promise;
			});
		});
		it("should be on the proper page", function() {
			expect(browser.getLocationAbsUrl()).toBe("/registration/" + savedData.securityKey);
		});
		it("should have the header include the council/group/pack name as appropriate", function() {
			expect(element.all(by.css("h1")).first().getText()).
				toBe("CCJ'16 Registration - Test group of Test council");
		});
		it("should have the correct leader information", function() {
			var leaderBlock = element.all(by.css("md-card-content")).get(0);
			expect(leaderBlock.element(by.tagName("h2")).getText()).toBe("Contact Leader");

			expect(leaderBlock.all(by.css("div.element")).get(0).element(by.tagName("p")).getText()).toBe("FirstN");
			expect(leaderBlock.all(by.css("div.element")).get(1).element(by.tagName("p")).getText()).toBe("LastN");
			expect(leaderBlock.all(by.css("div.element")).get(2).element(by.tagName("p")).getText()).toBe("test@invalid");
			expect(leaderBlock.all(by.css("div.element")).get(3).element(by.tagName("p")).getText()).toBe("123-456-7890");
		});
		it("should have the estimated participant count information", function() {
			var countBlock = element.all(by.css("md-card-content")).get(1);
			expect(countBlock.element(by.tagName("h2")).getText()).toBe("Estimated Participant counts");

			expect(countBlock.all(by.css("div.element")).get(0).element(by.tagName("p")).getText()).toBe("20");
			expect(countBlock.all(by.css("div.element")).get(1).element(by.tagName("p")).getText()).toBe("12");
		});
		it("should have the deposit information visible", function() {
			var elm = element.all(by.tagName("md-card")).get(-1);
			expect(elm.isDisplayed()).toBeTruthy();
		});
		it("should have the invoice link", function() {
			var elm = element.all(by.tagName("md-card-content")).get(-1).element(by.css("a"));
			expect(elm.getText()).toBe("invoice")
			expect(elm.getAttribute("href")).toMatch(/registration\/.*\/invoice/);
		});
	});

	describe("with no pack information submitted", function() {
		var savedData;
		beforeEach(function() {
			var flow = protractor.promise.controlFlow();
			flow.execute(function() {
				var defer = protractor.promise.defer();
				browser.executeAsyncScript(function(callback) {
					var Registration = angular.element(document.body).injector().get("Registration");
					var reg = new Registration();
					reg.council = "Test council";
					reg.groupName = "Test group";
					reg.packName = "Test pack";
					reg.contactLeaderFirstName = "FirstN";
					reg.contactLeaderLastName = "LastN";
					reg.contactLeaderEmail = "test@invalid";
					reg.contactLeaderPhoneNumber= "123-456-7890";
					reg.estimatedYouth = 20;
					reg.estimatedLeaders = 12;
					reg.agreedToEmailTerms(true);

					reg.$save().then(function(savedObject) {
						callback([true, savedObject]);
					}, function() {
						callback([false]);
					});
				}).then(function(data) {
					var success = data[0];
					if (success) {
						defer.fulfill(data[1]);
					} else {
						defer.reject("Failed to clear database");
					}
				});
				defer.promise.then(function(data) {
					savedData = data;
					browser.setLocation("/registration/" + data.securityKey);
				});
				return defer.promise;
			});
		});
		it("should be on the proper page", function() {
			expect(browser.getLocationAbsUrl()).toBe("/registration/" + savedData.securityKey);
		});
		it("should have the header include the council/group/pack name as appropriate", function() {
			expect(element.all(by.css("h1")).first().getText()).
				toBe("CCJ'16 Registration - Test group of Test council (Test pack)");
		});
	});

	describe("submitted onto the waiting list", function() {
		var savedData;
		beforeEach(function() {
			var flow = protractor.promise.controlFlow();
			flow.execute(function() {
				var defer = protractor.promise.defer();
				browser.executeAsyncScript(function(baseUrl, callback) {
					var $http = angular.element(document.body).injector().get("$http");
					$http({ url: baseUrl + "/integration/config" }).then(function(resp) {
						resp.data.General.EnableWaitingList = true
						$http({ url: baseUrl + "/integration/config", method: "POST", data: resp.data }).then(function() {
							var Registration = angular.element(document.body).injector().get("Registration");
							var reg = new Registration();
							reg.council = "Test council";
							reg.groupName = "Test group";
							reg.packName = "Test pack";
							reg.contactLeaderFirstName = "FirstN";
							reg.contactLeaderLastName = "LastN";
							reg.contactLeaderEmail = "test@invalid";
							reg.contactLeaderPhoneNumber= "123-456-7890";
							reg.estimatedYouth = 20;
							reg.estimatedLeaders = 12;
							reg.agreedToEmailTerms(true);

							reg.$save().then(function(savedObject) {
								callback([true, savedObject]);
							}, function() {
								callback([false]);
							});
						}, function () {
							callback(false)
						});
					}, function () {
						callback(false)
					});
				}, browser.baseUrl).then(function(data) {
					var success = data[0];
					if (success) {
						defer.fulfill(data[1]);
					} else {
						defer.reject("Failed to clear database");
					}
				});
				defer.promise.then(function(data) {
					savedData = data;
					browser.setLocation("/registration/" + data.securityKey);
				});
				return defer.promise;
			});
		});
		it("should have the deposit information invisible", function() {
			var elm = element.all(by.tagName("md-card")).get(-1);
			expect(elm.isDisplayed()).toBeFalsy();
		});
	});
});
