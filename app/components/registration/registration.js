'use strict';

angular.module('ccj16reg.registration', ['ngResource'])
.factory('registration', function($resource) {
	return $resource('/api/preregistration/:securityKey')
});
