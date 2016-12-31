"use strict";

var page = new (require("./po/invoice"))()
var moment = require("moment")
require("moment-timezone")

describe("Invoice information page", function() {
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
						defer.reject("Failed to store group in database");
					}
				});
				defer.promise.then(function(data) {
					savedData = data;
					page.get(data.securityKey);
				});
				return defer.promise;
			})
		})
		it("should have the right invoice number", function() {
			expect(page.id).toBe("1")
		})
		it("should have the date", function() {
			expect(page.date).toBe(moment().tz("America/Toronto").format("MMMM D, YYYY"))
		})
		it("should have an appropriate header", function() {
			expect(page.header).toBe("Invoice for Test group of Test council")
		})
		it("should have the preregistration line item", function() {
			expect(page.items).toEqual([{
				description: "Pre-registration deposit",
				count: "1",
				unitPrice: "250.00",
				total: "250.00"
			}])
		})
		it("should have the right total", function() {
			expect(page.total).toBe("250.00")
		})
	})
})
