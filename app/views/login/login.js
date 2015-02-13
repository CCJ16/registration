angular.module("ccj16reg.view.login", ["ngRoute", "ngMaterial", "ccj16reg.authentication"])

.config(function($routeProvider) {
	"use strict";
	$routeProvider.when("/login", {
		templateUrl: "views/login/login.html",
		controller: "LoginCtrl"
	});
})

.controller("LoginCtrl", function($scope, $location, $mdDialog, authentication) {
	"use strict";
	gapi.signin.render("myButton", {
		callback: function(authResult) {
			$scope.$apply(function() {
				callback(authResult);
			});
		},
	});
	function callback(authResult) {
		if (authResult.code) {
			authentication.tryGoogleToken(authResult.code).then(function(loggedIn) {
				if (!loggedIn) {
					$mdDialog.show(
						$mdDialog.alert()
							.title("Failed to login")
							.content("Server refused your account, please try again.")
							.ok("OK")
					);
				} else {
					$location.path("/admin/");
				}
			}, function() {
				$mdDialog.show(
					$mdDialog.alert()
						.title("Failed to login")
						.content("Server messed up, please complain loudly.")
						.ok("OK")
				);
			});
		}
	}
});
