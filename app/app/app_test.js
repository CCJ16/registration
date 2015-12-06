"use strict";

beforeEach(function() {
	angular.module("ccj16reg.config", []).constant("Config", {})
})

describe("ccj16reg module", function() {
	it("should initialize", function() {
		expect(angular.module("ccj16reg")).toBeTruthy()
	})
})
