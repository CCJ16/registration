'use strict';

angular.module('ccj16reg.view.recordlist', ['ngRoute', 'ccj16reg.registration'])

.config(['$routeProvider', function($routeProvider) {
	$routeProvider.when('/admin/recordlist', {
		templateUrl: 'views/recordlist/list.html',
		controller: 'RecordListCtrl',
		resolve: {
			checkAuth: loginRequired,
		},
	});
}])

.controller('RecordListCtrl', function($scope, registration) {
	$scope.registrations = registration.query();
});
