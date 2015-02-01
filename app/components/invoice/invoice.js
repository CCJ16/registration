angular.module("ccj16reg.invoice", ["ngResource", "ccj16reg.invoice.filters"])
.factory("invoice", function($resource) {
	"use strict";
	return $resource("/api/invoice/:invoiceId", null, {
		"getPreregistration": { method: "GET", url: "/api/preregistration/:securityKey/invoice" },
	});
});
