'use strict';

angular.module('ccj16reg.view.admin', ['ngRoute', 'ngMaterial', 'ccj16reg.authentication'])

.config(['$routeProvider', function($routeProvider) {
	$routeProvider.when('/admin/', {
		templateUrl: 'views/admin/admin.html',
		controller: 'AdminCtrl',
		resolve: {
			checkAuth: loginRequired,
		},
	});
}])

.controller('AdminCtrl', function() {
});
