'use strict';

angular.module('ccj16reg.view.register', ['ngRoute', 'ngMaterial', 'ccj16reg.registration'])

.config(['$routeProvider', function($routeProvider) {
	$routeProvider.when('/register', {
		templateUrl: 'views/register/register.html',
		controller: 'RegisterCtrl'
	});
}])

.controller('RegisterCtrl', ['$scope', '$location', '$mdDialog', 'registration', function($scope, $location, $mdDialog, registration) {
	$scope.registration = registration.new();

	$scope.submitRegistration = function(ev) {
		var progressDialog = $mdDialog.show({
			templateUrl: 'views/register/pending_submit.html',
			targetEvent: ev,
			clickOutsideToClose: false,
		});
		$scope.registration.save().then(function(data) {
			$mdDialog.hide();
			$location.path('/registration/' + data.securityKey)
		}, function(msg) {
			$mdDialog.hide();
			$mdDialog.show(
				$mdDialog.alert()
					.title('Failed to insert record')
					.content('Server message: ' + msg)
					.ok('OK')
			);
		});
	};
}]);
