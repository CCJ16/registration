"use strict";

describe("Initial registration process", function() {
	beforeEach(function() {
		browser.get("/");
	});

	it("should automatically redirect to /register when location hash/fragment is empty", function() {
		expect(browser.getLocationAbsUrl()).toMatch("/register");
	});

	describe("by registering", function() {
		it("should render the registration form without user information in the main header", function() {
			expect(element.all(by.css("h1")).first().getText()).
				toBe("CCJ16 Pre-registration");
		});
		it("should render the registration form in the view", function() {
			expect(element.all(by.css("[ng-view] h2")).first().getText()).
				toBe("Group pre-registration");
		});
		it("should have the submit button disabled by default", function() {
			var button = element(by.css("button.md-button.md-primary"));
			expect(button.isEnabled()).toBe(false);
		});

		describe("with the agreement checked", function() {
			beforeEach(function() {
				element.all(by.css("md-checkbox")).first().click();
			});

			it("should now have an enabled submit button", function() {
				var button = element(by.css("button.md-button.md-primary"));
				expect(button.isEnabled()).toBe(true);
			});

			describe("the form should be submittable when filled in", function() {
				beforeEach(function() {
					element(by.model("registration.council")).sendKeys("Test Council");
					element(by.model("registration.groupName")).sendKeys("Test Group");

					element(by.model("registration.contactLeaderFirstName")).sendKeys("FirstN");
					element(by.model("registration.contactLeaderLastName")).sendKeys("LastN");

					element(by.model("registration.contactLeaderEmail")).sendKeys("test@invalid");
					element(by.model("registration.contactLeaderPhoneNumber")).sendKeys("123-456-7890");

					var button = element(by.css("button.md-button.md-primary"));
					button.click();
				});

				it("should then redirect to /registration/...", function() {
					expect(browser.getLocationAbsUrl()).toMatch("/registration/");
				}, 60000);
			});
		});
	});
});
