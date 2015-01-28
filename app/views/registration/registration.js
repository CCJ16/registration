'use strict';

angular.module('ccj16reg.view.registration', ['ngRoute', 'ngMaterial', 'ccj16reg.registration'])

.config(['$routeProvider', function($routeProvider) {
	$routeProvider.when('/registration/:securityKey', {
		templateUrl: 'views/registration/registration.html',
		controller: 'RegistrationCtrl',
		resolve: {
			'registrationData': function($route, $location, registration) {
				return registration.get({securityKey: $route.current.params.securityKey}).$promise;
			},
		},
	});
}])

.controller('RegistrationCtrl', function($scope, $routeParams, $window, registrationData) {
	$scope.registration = registrationData;
	$scope.$emit('CurrentGroupInformationChanged', registrationData);
	$scope.printPage = function() {
		$window.print();
	};
});
