'use strict';

describe('ccj16reg.registration module', function() {
	var $httpBackend, registration;

	beforeEach(module('ccj16reg.registration'));

	beforeEach(inject(function($injector, _registration_) {
		$httpBackend = $injector.get('$httpBackend');

		registration = _registration_;
	}))

	afterEach(function() {
		$httpBackend.verifyNoOutstandingExpectation();
		$httpBackend.verifyNoOutstandingRequest();
	});

	describe('registration service', function() {
		describe('registration model', function() {
			it('should give a useful model', function() {
				var newReg = new registration();
				expect(newReg).toBeDefined();
				expect(newReg.$save).not.toBeNull();
			});

			it('that will save itself to a new object on save()', function() {
				var newReg = new registration();
				newReg.council = "A council";
				newReg.groupName = "4th Group";
				newReg.packName = "Pack G";
				newReg.contactLeaderEmail = "test@example.com";

				$httpBackend.expectPOST('/api/preregistration', angular.toJson(newReg)).respond(201, '');

				newReg.$save().then(function() {}, function(message) {
					console.log(message);
					expect(false).toBe(true);
				});
				$httpBackend.flush()
			});
		});
	});
});
