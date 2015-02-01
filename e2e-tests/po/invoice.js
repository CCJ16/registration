"use strict";

var InvoicePage = function() {
}

InvoicePage.prototype = Object.create({}, {
	get: { value: function(securityKey) {
		browser.get("/registration/" + securityKey + "/invoice")
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
		var items = []
		var defer = protractor.promise.defer()
		element.all(by.repeater("item in invoice.lineItems")).each(function(element) {
			var item = {}
			var elements = element.all(by.css("td"))
			elements.get(0).getText().then(function(text) {
				item.description = text
			})
			elements.get(1).getText().then(function(text) {
				item.count = text
			})
			elements.get(2).getText().then(function(text) {
				item.unitPrice = text
			})
			elements.get(3).getText().then(function(text) {
				item.total = text
			})
			items.push(item)
		}).then(function() {
			defer.fulfill(items)
		})
		return defer.promise
	}},
	total: { get: function() {
		return element(by.css("tr.total td.numeric")).getText()
	}}
})

module.exports = InvoicePage
