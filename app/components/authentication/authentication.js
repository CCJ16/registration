angular.module("ccj16reg.authentication", [])
.factory("authentication", function($http, $q) {
	"use strict";
	var loggedInP = null;
	return {
		isLoggedIn: function() {
			if (loggedInP === null) {
				loggedInP = $q(function(resolve, reject) {
					$http.get("/api/authentication/isLoggedIn").then(function(response) {
						if (response.data === "true") {
							resolve(true);
						} else {
							resolve(false);
						}
					}, function(response) {
						reject(response);
					});
				});
			}
			return loggedInP;
		},
		tryGoogleToken: function(token) {
			loggedInP = $q(function(resolve, reject) {
				$http.post("/api/authentication/googletoken", token).then(function(response) {
					if (response.data === "true") {
						resolve(true);
					} else {
						resolve(false);
					}
				}, function(response) {
					reject(response);
				});
			});
			return loggedInP;
		},
	};
});
