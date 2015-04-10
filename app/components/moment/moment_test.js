"use strict";

describe("ccj16reg.moment module", function() {

	beforeEach(module("ccj16reg.moment"));

	describe("moment filter", function() {
		var moment

		beforeEach(inject(function(_momentFilter_) {
			moment = _momentFilter_;
		}));
		it("when given the a date in a string, it should output correctly", function() {
			expect(moment("2015-04-03T12:34:56-04:00", "MMMM D, YYYY", "America/Toronto")).toBe("April 3, 2015")
		})
		it("when given the a date close to midnight in a different timezone, output the correct date in the correct timezone", function() {
			expect(moment("2015-04-04T00:34:56-03:00", "MMMM D, YYYY", "America/Toronto")).toBe("April 3, 2015")
		})
	})
});
