"use strict";

var ConfirmEmail = function() {
}

ConfirmEmail.prototype = Object.create({}, {
	get: { value: function(email, token) {
		browser.setLocation("/confirmpreregistration?email=" + email + "&token=" + token)
	}},
	content: { get: function() {
		return element(by.css("md-card-content > div")).getText()
	}},
})

module.exports = ConfirmEmail
