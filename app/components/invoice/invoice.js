angular.module("ccj16reg.invoice", ["ngResource"])
.factory("invoice", function($resource) {
	"use strict";
	return $resource("/api/invoice/:invoiceId", null, {
		"getPreregistration": { method: "GET", url: "/api/preregistration/:securityKey/invoice" },
	});
});
