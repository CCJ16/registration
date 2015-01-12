'use strict';

angular.module('ccj16reg.view.registration', ['ngRoute', 'ngMaterial', 'ccj16reg.registration'])

.config(['$routeProvider', function($routeProvider) {
	$routeProvider.when('/registration/:securityKey', {
		templateUrl: 'views/registration/registration.html',
		controller: 'RegistrationCtrl'
	});
}])

.controller('RegistrationCtrl', function($scope, $routeParams, registration) {
	$scope.registration = registration.get({securityKey: $routeParams.securityKey})
});
