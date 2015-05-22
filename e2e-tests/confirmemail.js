"use strict";

var page = new (require("./po/confirmemail"))()

describe("Initial registration process", function() {
	it("should fail when given a bad email/token combo", function() {
		page.get("notanemail@invalid", "notatoken")
		browser.waitForAngular()
		expect(page.content).toBe("Failed to verify notanemail@invalid! (Received error: Failed to verify token)")
	});
});
