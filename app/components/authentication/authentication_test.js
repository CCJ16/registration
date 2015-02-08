'use strict';

describe('ccj16reg.authentication module', function() {

	beforeEach(module('ccj16reg.authentication'));

	describe('authentication service', function() {
		var $httpBackend, $rootScope, authentication;

		beforeEach(inject(function($injector, _authentication_) {
			$httpBackend = $injector.get('$httpBackend');
			$rootScope = $injector.get('$rootScope');

			authentication = _authentication_;
		}))

		afterEach(function() {
			$httpBackend.verifyNoOutstandingExpectation();
			$httpBackend.verifyNoOutstandingRequest();
		});
		it('should return unauthenticated when session is blank', function() {
			$httpBackend.expectGET('/api/authentication/isLoggedIn').respond(200, 'false')
			var gspy = jasmine.createSpy('gspy');
			var bspy = jasmine.createSpy('bspy');

			authentication.isLoggedIn().then(gspy, bspy);
			$httpBackend.flush();
			expect(gspy).toHaveBeenCalled();
			expect(bspy).not.toHaveBeenCalled();

			expect(gspy).toHaveBeenCalledWith(false);
		});
		it('should return authenticated when session reports success', function() {
			$httpBackend.expectGET('/api/authentication/isLoggedIn').respond(200, 'true')
			var gspy = jasmine.createSpy('gspy');
			var bspy = jasmine.createSpy('bspy');

			authentication.isLoggedIn().then(gspy, bspy);
			$httpBackend.flush();
			expect(gspy).toHaveBeenCalled();
			expect(bspy).not.toHaveBeenCalled();

			expect(gspy).toHaveBeenCalledWith(true);
		});
		it('should send a valid request to the backend and succeed.', function() {
			$httpBackend.expectPOST('/api/authentication/googletoken', 'goodToken').respond(200, 'true')

			var responseP = authentication.tryGoogleToken('goodToken');
			var gspy = jasmine.createSpy('gspy');
			var bspy = jasmine.createSpy('bspy');
			responseP.then(gspy, bspy);
			$httpBackend.flush();
			expect(gspy).toHaveBeenCalled();
			expect(bspy).not.toHaveBeenCalled();

			expect(gspy).toHaveBeenCalledWith(true);

			gspy = jasmine.createSpy('gspy');
			bspy = jasmine.createSpy('bspy');

			authentication.isLoggedIn().then(gspy, bspy);
			$rootScope.$digest();
			expect(gspy).toHaveBeenCalled();
			expect(bspy).not.toHaveBeenCalled();

			expect(gspy).toHaveBeenCalledWith(true);
		});
	});
});
