"use strict";

var InvoicePage = function() {
}

InvoicePage.prototype = Object.create({}, {
	get: { value: function(securityKey) {
		browser.setLocation("/registration/" + securityKey + "/invoice")
	}},
	header: { get: function() {
		return element(by.css("h2")).getText()
	}},
	id: { get: function() {
		return element(by.css(".right.element p")).getText()
	}},
	date: { get: function() {
		return element(by.css(".left.element p")).getText()
	}},
	items: { get: function() {
		return element.all(by.repeater("item in invoice.lineItems")).map(function(element) {
			var elements = element.all(by.css("td"))
			return {
				description: elements.get(0).getText(),
				count: elements.get(1).getText(),
				unitPrice: elements.get(2).getText(),
				total: elements.get(3).getText(),
			}
		})
	}},
	total: { get: function() {
		return element(by.css("tr.total td.numeric")).getText()
	}}
})

module.exports = InvoicePage
