'use strict';

angular.module('ccj16reg.authentication', [])
.factory('authentication', function($http, $q) {
	var loggedInP = $q(function(resolve) {
		resolve(false);
	});
	return {
		isLoggedIn: function() {
			return loggedInP;
		},
		tryGoogleToken: function(token) {
			loggedInP = $q(function(resolve, reject) {
				$http.post('/api/authentication/googletoken', token).then(function(response) {
					if (response.data === 'true') {
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
