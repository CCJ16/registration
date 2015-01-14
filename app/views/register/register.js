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

	$scope.$watch('registration.agreedToEmailTerms()', function(checked) {
		$scope.registrationTosAccepted = checked;
	});

	$scope.showEmailTos = function(ev) {
		ev.preventDefault()
		ev.stopPropagation();
		$mdDialog.show(
			$mdDialog.alert()
				.title('Email usage agreement')
				.content('All email addresses received through the registration process have been  added to the CCJ16  registration database as well as the CCJ16  mailing list.  During the coming months, you will receive mails from time to time with information about our upcoming CCJ, planning for attending the camp, and other information related to this camp.  Your address will not be distributed to others, nor used for matters not directly connected with CCJ16.  Should you wish to be removed from this distribution list, please send an email to info@cubjamboree.ca')
				.ok('Done')
				.targetEvent(ev)
		);
	}

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
