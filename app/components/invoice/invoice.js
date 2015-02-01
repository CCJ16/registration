'use strict';

angular.module('ccj16reg.invoice', ['ngResource'])
.factory('invoice', function($resource, $http) {
	return $resource('/api/invoice/:invoiceId', null, {
		'getPreregistration': { method: 'GET', url: '/api/preregistration/:securityKey/invoice' },
	});
});
