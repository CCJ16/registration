"use strict";

describe("ccj16reg.registration module", function() {

	beforeEach(module("ccj16reg.registration"));

	describe("registration service", function() {
		var $httpBackend, $rootScope, Registration, curDate = new Date().toISOString();

		beforeEach(function() {
			module(function($provide) {
				$provide.value("currentDateFetch", function() {
					return curDate;
				});
			});
		});

		beforeEach(inject(function($injector, _Registration_) {
			$httpBackend = $injector.get("$httpBackend");
			$rootScope = $injector.get("$rootScope");

			Registration = _Registration_;
		}))

		afterEach(function() {
			$httpBackend.verifyNoOutstandingExpectation();
			$httpBackend.verifyNoOutstandingRequest();
		});
		it("should give a useful model", function() {
			var newReg = new Registration();
			expect(newReg).toBeDefined();
			expect(newReg.$save).not.toBeNull();
		});

		describe("should have a working agreedToEmailTerms", function() {

			it("should exist as a function", function() {
				var newReg = new Registration();
				expect(newReg.agreedToEmailTerms).toBeDefined();
			});

			it("should return false for an empty object", function() {
				var newReg = new Registration();
				expect(newReg.agreedToEmailTerms()).toBe(false);
			});

			it("should return false for a zero date from Go", function() {
				var newReg = new Registration();
				newReg.emailApprovalGivenAt = "0001-01-01T00:00:00Z";
				expect(newReg.agreedToEmailTerms()).toBe(false);
			});

			it("should return true when set", function() {
				var newReg = new Registration();
				newReg.emailApprovalGivenAt = new Date().toISOString();
				expect(newReg.agreedToEmailTerms()).toBe(true);
			});

			it("should set the current (injected) date when set to true.", function() {
				var newReg = new Registration();
				expect(newReg.agreedToEmailTerms(true)).toBe(true);
				expect(newReg.emailApprovalGivenAt).toBe(curDate);
			});

			it("should keep the original time if asked to be true again.", function() {
				var newReg = new Registration();
				expect(newReg.agreedToEmailTerms(true)).toBe(true);
				expect(newReg.emailApprovalGivenAt).toBe(curDate);

				var oldDate = curDate;
				curDate = new Date().toISOString();
				expect(newReg.agreedToEmailTerms(true)).toBe(true);
				expect(newReg.emailApprovalGivenAt).toBe(oldDate);
			});

			it("should get the current time if asked to be true after being false.", function() {
				var newReg = new Registration();
				expect(newReg.agreedToEmailTerms(true)).toBe(true);
				expect(newReg.emailApprovalGivenAt).toBe(curDate);

				curDate = new Date().toISOString();

				expect(newReg.agreedToEmailTerms(false)).toBe(false);
				expect(newReg.agreedToEmailTerms(true)).toBe(true);
				expect(newReg.emailApprovalGivenAt).toBe(curDate);
			});

			it("should stay undefined when set to being false.", function() {
				var newReg = new Registration();
				expect(newReg.agreedToEmailTerms(false)).toBe(false);
				expect(newReg.emailApprovalGivenAt).not.toBeDefined();
			});

			it("should set it back to undefined when set to being false.", function() {
				var newReg = new Registration();
				expect(newReg.agreedToEmailTerms(true)).toBe(true);
				expect(newReg.emailApprovalGivenAt).toBe(curDate);
				expect(newReg.agreedToEmailTerms(false)).toBe(false);
				expect(newReg.emailApprovalGivenAt).not.toBeDefined();
			});
		});

		it("that will save itself to a new object on save()", function() {
			var newReg = new Registration();
			newReg.council = "A council";
			newReg.groupName = "4th Group";
			newReg.packName = "Pack G";
			newReg.contactLeaderEmail = "test@example.com";

			$httpBackend.expectPOST("/api/preregistration", angular.toJson(newReg)).respond(201, "");

			newReg.$save().then(function() {}, function(message) {
				console.log(message);
				expect(false).toBe(true);
			});
			$httpBackend.flush()
		});

		describe("should have a working validatedEmail", function() {

			it("should exist as a function", function() {
				var newReg = new Registration();
				expect(newReg.validatedEmail).toBeDefined();
			});

			it("should return false for an empty object", function() {
				var newReg = new Registration();
				expect(newReg.validatedEmail()).toBe(false);
			});

			it("should return false for a zero date from Go", function() {
				var newReg = new Registration();
				newReg.validatedOn = "0001-01-01T00:00:00Z";
				expect(newReg.validatedEmail()).toBe(false);
			});

			it("should return true when set", function() {
				var newReg = new Registration();
				newReg.validatedOn = new Date().toISOString();
				expect(newReg.validatedEmail()).toBe(true);
			});
		});

		describe("should have a working promote function", function() {

			it("should exist as a function", function() {
				var newReg = new Registration();
				expect(newReg.promote).toBeDefined();
			});

			it("should return a failing promise for a regular registration", function() {
				var newReg = new Registration()
				newReg.securityKey = "key"
				newReg.isOnWaitingList = false

				var good
				newReg.promote().then(function() {
					good = false
				}, function(error) {
					good = error
				})

				$rootScope.$apply();
				expect(good).toBe("Group is not on the waiting list")
			});

			it("should return poke the right endpoint and succeed on 200", function() {
				var newReg = new Registration()
				newReg.securityKey = "key"
				newReg.isOnWaitingList = true

				$httpBackend.expectPOST("/api/preregistration/key/promote").respond(200, "")

				var good

				newReg.promote().then(function() {
					good = true
				}, function() {
					good = false
				})

				$httpBackend.flush()
				expect(good).toBe(true)
			});

			it("should return poke the right endpoint and fail with the error message on !2xx", function() {
				var newReg = new Registration()
				newReg.securityKey = "key"
				newReg.isOnWaitingList = true

				$httpBackend.expectPOST("/api/preregistration/key/promote").respond(500, "Error message")

				var good

				newReg.promote().then(function() {
					good = false
				}, function(error) {
					good = error
				})

				$httpBackend.flush()
				expect(good).toBe("Error message")
			});
		});

		describe("will have an interface for confirming emails", function() {
			it("that succeeds with valid information", function() {
				var good;

				$httpBackend.expectPUT("/api/confirmpreregistration?email=test@example.com", "validToken").respond(204, "");

				Registration.confirmEmail("test@example.com", "validToken").then(function() {
					good = true;
				}, function() {
					good = false;
				})

				$httpBackend.flush();

				expect(good).toBe(true);
			});
			it("that fails with bad information", function() {
				var good;

				$httpBackend.expectPUT("/api/confirmpreregistration?email=test@example.com", "badToken").respond(400, "");

				Registration.confirmEmail("test@example.com", "badToken").then(function() {
					good = true;
				}, function() {
					good = false;
				})

				$httpBackend.flush();

				expect(good).toBe(false);
			});
		});
	});

	describe("currentDateFetch service", function() {
		var oldDateFunction, currentDateFetch;
		var curDate = new Date();
		var newDateFunction = function() {
			return curDate;
		}

		beforeEach(function() {
			oldDateFunction = window.Date;
			window.Date = newDateFunction;
		});
		afterEach(function() {
			window.Date = oldDateFunction;
		});

		beforeEach(inject(function(_currentDateFetch_) {
			currentDateFetch = _currentDateFetch_;
		}))

		it("should return the current date", function() {
			expect(currentDateFetch()).toBe(curDate.toISOString());
		});
	});
});
