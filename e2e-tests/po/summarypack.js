"use strict";

var SummaryPackPage = function() {
}

SummaryPackPage.prototype = Object.create({}, {
	get: { value: function() {
		browser.setLocation("/summary/pack")
	}},
	leaderCount: { get: function() {
		return element(by.id("leaderCount")).getText()
	}},
	youthCount: { get: function() {
		return element(by.id("youthCount")).getText()
	}},
})

module.exports = SummaryPackPage
