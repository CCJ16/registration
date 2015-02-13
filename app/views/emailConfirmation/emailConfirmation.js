angular.module("ccj16reg.view.emailConfirmation", ["ngRoute", "ngMaterial", "ccj16reg.registration"])

.config(function($routeProvider) {
	"use strict";
	$routeProvider.when("/confirmpreregistration", {
		templateUrl: "views/emailConfirmation/emailConfirmation.html",
		controller: "EmailConfirmationCtrl",
	});
})

.controller("EmailConfirmationCtrl", function($scope, $location, $mdDialog, $routeParams, registration) {
	"use strict";
	$scope.verifying = true;
	$scope.error = false;

	$scope.email = $routeParams.email;
	var token = $routeParams.token;
	if (!angular.isDefined($scope.email) || !angular.isDefined(token)) {
		$location.path("/register")
	} else {
		registration.confirmEmail($scope.email, token).then(function() {
			$scope.verifying = false;
		}, function(resp) {
			$scope.verifying = false;
			$scope.error = resp.data;
		});
	}
});
