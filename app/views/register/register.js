'use strict';

angular.module('ccj16reg.view.register', ['ngRoute', 'ngMaterial', 'ccj16reg.registration'])

.config(['$routeProvider', function($routeProvider) {
	$routeProvider.when('/register', {
		templateUrl: 'views/register/register.html',
		controller: 'RegisterCtrl'
	});
}])

.controller('RegisterCtrl', ['$scope', '$mdDialog', 'registration', function($scope, $mdDialog, registration) {
	$scope.registration = registration.new();

	$scope.submitRegistration = function(ev) {
		$mdDialog.show({
			templateUrl: 'views/register/pending_submit.html',
			targetEvent: ev,
			clickOutsideToClose: false,
		});
		$scope.registration.save();
	};
}]);
