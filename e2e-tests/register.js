"use strict"

function testRegPage(headerText) {
	it("should render the registration form without user information in the main header", function() {
		expect(element.all(by.css("h1")).first().getText()).
			toBe("CCJ'16 Registration")
	})
	it("should render the registration form in the view", function() {
		expect(element.all(by.css("[ng-view] h2")).first().getText()).
			toBe(headerText)
	})
	it("should have the submit button disabled by default", function() {
		var button = element(by.css("button.md-button.md-primary"))
		expect(button.isEnabled()).toBe(false)
	})

	describe("with the agreement checked", function() {
		beforeEach(function() {
			element.all(by.css("md-checkbox")).first().click()
		})

		it("should now have an enabled submit button", function() {
			var button = element(by.css("button.md-button.md-primary"))
			expect(button.isEnabled()).toBe(true)
		})

		describe("the form should be submittable when filled in", function() {
			beforeEach(function() {
				element(by.model("registration.council")).sendKeys("Test Council")
				element(by.model("registration.groupName")).sendKeys("Test Group")

				element(by.model("registration.contactLeaderFirstName")).sendKeys("FirstN")
				element(by.model("registration.contactLeaderLastName")).sendKeys("LastN")

				element(by.model("registration.contactLeaderEmail")).sendKeys("test@invalid")
				element(by.model("registration.contactLeaderPhoneNumber")).sendKeys("123-456-7890")

				var button = element(by.css("button.md-button.md-primary"))
				button.click()
			})

			it("should then redirect to /registration/...", function() {
				expect(browser.getLocationAbsUrl()).toMatch("/registration/")
			}, 60000)
		})
	})
}

describe("Initial registration process", function() {
	beforeEach(function() {
		browser.setLocation("/")
	})

	it("should automatically redirect to /register when location hash/fragment is empty", function() {
		expect(browser.getLocationAbsUrl()).toMatch("/register")
	})

	describe("by registering with open registration", function() {
		beforeEach(function() {
			browser.get("/")
		})
		testRegPage("Group pre-registration")
	})

	describe("by registering within waiting list mode", function() {
		beforeEach(function() {
			var flow = protractor.promise.controlFlow()
			flow.execute(function() {
				var defer = protractor.promise.defer()
				browser.executeAsyncScript(function(baseUrl, callback) {
					var $http = angular.element(document.body).injector().get("$http")
					$http({ url: baseUrl + "/integration/config" }).then(function(resp) {
						resp.data.General.EnableWaitingList = true
						$http({ url: baseUrl + "/integration/config", method: "POST", data: resp.data }).then(function() {
							callback([true])
						}, function () {
							callback([false])
						})
					}, function () {
						callback([false])
					})
				}, browser.baseUrl).then(function(data) {
					var success = data[0]
					if (success) {
						defer.fulfill()
					} else {
						defer.reject("Failed to enter waiting list mode")
					}
				})
				defer.promise.then(function() {
					browser.get("/")
				})
				return defer.promise
			})
		})
		testRegPage("CCJ'16 Wait List")
	})

	describe("by attempting registration when disabed", function() {
		beforeEach(function() {
			var flow = protractor.promise.controlFlow()
			flow.execute(function() {
				var defer = protractor.promise.defer()
				browser.executeAsyncScript(function(baseUrl, callback) {
					var $http = angular.element(document.body).injector().get("$http")
					$http({ url: baseUrl + "/integration/config" }).then(function(resp) {
						resp.data.General.EnableGroupReg = false
						$http({ url: baseUrl + "/integration/config", method: "POST", data: resp.data }).then(function() {
							callback([true])
						}, function () {
							callback([false])
						})
					}, function () {
						callback([false])
					})
				}, browser.baseUrl).then(function(data) {
					var success = data[0]
					if (success) {
						defer.fulfill()
					} else {
						defer.reject("Failed to enter waiting list mode")
					}
				})
				defer.promise.then(function() {
					browser.get("/")
				})
				return defer.promise
			})
		})
		it("should render without user information in the main header", function() {
			expect(element.all(by.css("h1")).first().getText()).
				toBe("CCJ'16 Registration")
		})
		it("should render the sorry message in the view", function() {
			expect(element.all(by.css("[ng-view] h2")).first().getText()).
				toBe("CCJ'16 Cub Pack Registration Closed")
		})
	})
})
