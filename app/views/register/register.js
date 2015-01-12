'use strict';

angular.module('ccj16reg.view.register', ['ngRoute', 'ccj16reg.registration'])

.config(['$routeProvider', function($routeProvider) {
	$routeProvider.when('/register', {
		templateUrl: 'views/register/register.html',
		controller: 'RegisterCtrl'
	});
}])

.controller('RegisterCtrl', ['$scope', 'registration', function($scope, registration) {
	$scope.registration = registration.new();

	$scope.submitRegistration = function() {
		$scope.registration.save();
	};
}]);
