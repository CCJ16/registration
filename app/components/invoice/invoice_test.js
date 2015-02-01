"use strict";

describe("ccj16reg.invoice module", function() {

	beforeEach(module("ccj16reg.invoice"));

	describe("invoice service", function() {
		var $httpBackend, invoice;

		beforeEach(inject(function($injector, _invoice_) {
			$httpBackend = $injector.get("$httpBackend");

			invoice = _invoice_;
		}))

		afterEach(function() {
			$httpBackend.verifyNoOutstandingExpectation();
			$httpBackend.verifyNoOutstandingRequest();
		});
		it("should properly request the preregistration invoice", function() {
			$httpBackend.expectGET("/api/preregistration/MyKey/invoice").respond(200, {prop: "value"})

			var newReg = invoice.getPreregistration({securityKey: "MyKey"});
			$httpBackend.flush();
			expect(newReg).toBeDefined();
			expect(newReg.prop).toBe("value");
		});
	});

	describe("centToDollars filter", function() {
		var centToDollars;

		beforeEach(inject(function($injector, _centToDollarsFilter_) {
			centToDollars = _centToDollarsFilter_;
		}));
		it("should work with a number greater then 100 cents", function() {
			expect(centToDollars(1562)).toBe("15.62");
		});
		it("should work with a number 10 <= n <= 100", function() {
			expect(centToDollars(52)).toBe("0.52");
		});
		it("should work with a number 1 <= n <= 10", function() {
			expect(centToDollars(2)).toBe("0.02");
		});
		it("should work with 0", function() {
			expect(centToDollars(0)).toBe("0.00");
		});
	});

	describe("invoiceSum filter", function() {
		var invoiceSum

		beforeEach(inject(function(_invoiceSumFilter_) {
			invoiceSum = _invoiceSumFilter_;
		}));
		it("should return zero for an empty array", function() {
			expect(invoiceSum([])).toBe(0)
		})
		it("should return the exact amount for a single entry with multiple units", function() {
			var items = [{ unitPrice: 20, count: 5 }]
			expect(invoiceSum(items)).toBe(100)
		})
		it("should return the exact amount for multiple entries with varying units", function() {
			var items = [{ unitPrice: 20, count: 5 }, { unitPrice:500, count: 1 }]
			expect(invoiceSum(items)).toBe(600)
		})
	})
});
