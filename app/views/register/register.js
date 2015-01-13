'use strict';

angular.module('ccj16reg.view.register', ['ngRoute', 'ngMaterial', 'ccj16reg.registration'])

.config(['$routeProvider', function($routeProvider) {
	$routeProvider.when('/register', {
		templateUrl: 'views/register/register.html',
		controller: 'RegisterCtrl'
	});
}])

.controller('RegisterCtrl', ['$scope', '$location', '$mdDialog', 'registration', function($scope, $location, $mdDialog, registration) {
	$scope.registration = new registration();

	$scope.submitRegistration = function(ev) {
		$mdDialog.show({
			templateUrl: 'views/register/pending_submit.html',
			targetEvent: ev,
			clickOutsideToClose: false,
			escapeToClose: false,
			onComplete: submitSaveRequest,
		});
		function submitSaveRequest() {
			$scope.registration.$save().then(function(reg) {
				$mdDialog.hide();
				$location.path('/registration/' + reg.securityKey)
			}, function(msg) {
				$mdDialog.hide();
				$mdDialog.show(
					$mdDialog.alert()
						.title('Failed to insert record')
						.content('Server message: ' + msg)
						.ok('OK')
				);
			});
		}
	};
}]);
