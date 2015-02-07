'use strict';

// Declare app level module which depends on views, and components
angular.module('ccj16reg', [
	'ngRoute',
	'ngMaterial',
	'ccj16reg.registration',
	'ccj16reg.view.admin',
	'ccj16reg.view.login',
	'ccj16reg.view.register',
	'ccj16reg.view.registration',
	'ccj16reg.view.emailConfirmation',
])
.config(['$routeProvider', '$locationProvider', function($routeProvider, $locationProvider) {
	$routeProvider.otherwise({redirectTo: '/register'});
	$locationProvider.html5Mode(true).hashPrefix("!");
}])
.config(function($mdThemingProvider) {
	$mdThemingProvider.theme('default')
		.primaryPalette('yellow', {
			'default': '500',
			'hue-1': 'A200',
			'hue-2': 'A400',
			'hue-3': '400',
		})
		.accentPalette('red', {
			'default': '900',
			'hue-1': '500',
		})
		.backgroundPalette('grey', {
			'default': '100',
		});
})
.controller('CtrlApp', function($scope) {
	$scope.packDisplayName = '';
	$scope.$on('CurrentGroupInformationChanged', function(event, registration) {
		$scope.packDisplayName = " - " + registration.groupName + " of " + registration.council;
		if (registration.packName) {
			$scope.packDisplayName += " (" + registration.packName + ")";
		}
	});
	$scope.$on('$locationChangeSuccess', function() {
		$scope.packDisplayName = '';
	});
})
.factory('xsrfFailureFixer',function($q, $injector) {
	var attempts = 5;
	return {
		'responseError': function(response) {
			if (response.status === 400 && response.data == "Invalid XSRF token\n" && attempts > 0) {
				attempts--;
				var $http = $injector.get('$http');
				return $http(response.config);
			}
			return $q.reject(response);
		},
		'response': function(response) {
			if (response.status === 200 && /^\/api/.test(response.config.url)) {
				attempts = 5;
			}
			return response;
		},
	};
})
.config(function($httpProvider) {
	$httpProvider.interceptors.push('xsrfFailureFixer');
});

function loginRequired($location, $q, authentication) {
	return $q(function(resolve, reject) {
		authentication.isLoggedIn().then(function(res) {
			if (res) {
				resolve();
			} else {
				reject();
				$location.path('/login');
			}
		});
	});
}
