'use strict';

describe('ccj16reg.invoice module', function() {

	beforeEach(module('ccj16reg.invoice'));

	describe('invoice service', function() {
		var $httpBackend, invoice;

		beforeEach(inject(function($injector, _invoice_) {
			$httpBackend = $injector.get('$httpBackend');

			invoice = _invoice_;
		}))

		afterEach(function() {
			$httpBackend.verifyNoOutstandingExpectation();
			$httpBackend.verifyNoOutstandingRequest();
		});
		it('should properly request the preregistration invoice', function() {
			$httpBackend.expectGET('/api/preregistration/MyKey/invoice').respond(200, {prop: 'value'})

			var newReg = invoice.getPreregistration({securityKey: 'MyKey'});
			$httpBackend.flush();
			expect(newReg).toBeDefined();
			expect(newReg.prop).toBe('value');
		});
	});
});
