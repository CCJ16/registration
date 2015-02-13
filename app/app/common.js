angular.module("ccj16reg.common", [])
.constant("resolveLoginRequired", function ($location, $q, authentication) {
	"use strict";
	return $q(function(resolve, reject) {
		authentication.isLoggedIn().then(function(res) {
			if (res) {
				resolve();
			} else {
				reject();
				$location.path("/login");
			}
		});
	});
})
