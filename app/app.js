'use strict';

// Declare app level module which depends on views, and components
angular.module('ccj16reg', [
	'ngRoute',
	'ngMaterial',
	'ccj16reg.registration',
	'ccj16reg.view.register',
	'ccj16reg.view.registration',
]).
config(['$routeProvider', function($routeProvider) {
	$routeProvider.otherwise({redirectTo: '/register'});
}])
.config(function($mdThemingProvider) {
	$mdThemingProvider.theme('default')
		.primaryColor('yellow', {
			'default': '500',
			'hue-1': 'A200',
			'hue-2': 'A400',
			'hue-3': '400',
		})
		.accentColor('red', {
			'default': '900',
		})
		.backgroundColor('grey', {
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
});
