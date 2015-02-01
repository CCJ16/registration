angular.module("ccj16reg.invoice.filters", [])
.filter("centToDollars", function() {
	"use strict";
	return function(input) {
		var inStr = input.toString();
		if (inStr.length > 2) {
			return inStr.substring(0, inStr.length - 2) + "." + inStr.substring(inStr.length - 2);
		} else if (inStr.length > 1) {
			return "0." + inStr;
		} else {
			return "0.0" + inStr;
		}
	}
})
.filter("invoiceSum", function() {
	"use strict";
	return function(input) {
		var sum = 0
		for(var i = 0; i < input.length; ++i) {
			sum += input[i].count * input[i].unitPrice
		}
		return sum
	}
})
