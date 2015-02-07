'use strict';

angular.module('ccj16reg.view.login', ['ngRoute', 'ngMaterial'])

.config(['$routeProvider', function($routeProvider) {
	$routeProvider.when('/login', {
		templateUrl: 'views/login/login.html',
		controller: 'LoginCtrl'
	});
}])

.controller('LoginCtrl', function($scope) {
	gapi.signin.render('myButton', {
		callback: function(data) {
			$scope.$apply(callback)
		},
	});
	function callback(authResult) {
	}
});
