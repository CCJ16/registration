angular.module("ccj16reg.view.registration", ["ngRoute", "ngMaterial", "ccj16reg.registration"])

.config(function($routeProvider) {
	"use strict";
	$routeProvider.when("/registration/:securityKey", {
		templateUrl: "views/registration/registration.html",
		controller: "RegistrationCtrl",
		resolve: {
			"registrationData": function($route, $location, Registration) {
				return Registration.get({securityKey: $route.current.params.securityKey}).$promise;
			},
		},
	});
})

.controller("RegistrationCtrl", function($scope, $routeParams, $window, registrationData) {
	"use strict";
	$scope.registration = registrationData;
	$scope.$emit("CurrentGroupInformationChanged", registrationData);
	$scope.printPage = function() {
		$window.print();
	};
});
