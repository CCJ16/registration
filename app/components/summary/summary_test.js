"use strict";

describe("ccj16reg.summary module", function() {

	beforeEach(module("ccj16reg.summary"));

	describe("summary service", function() {
		var $httpBackend, summary;

		beforeEach(inject(function($injector, _summary_) {
			$httpBackend = $injector.get("$httpBackend");

			summary = _summary_;
		}))

		afterEach(function() {
			$httpBackend.verifyNoOutstandingExpectation();
			$httpBackend.verifyNoOutstandingRequest();
		});
		it("should request pack summary information for getPack function", function() {
			$httpBackend.expectGET("/api/summary/pack").respond(200, {prop: "value"})

			var summaryInfo = summary.getPack();
			$httpBackend.flush();
			expect(summaryInfo).toBeDefined();
			expect(summaryInfo.prop).toBe("value");
		});
	});
});
