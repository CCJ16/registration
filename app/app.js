'use strict';

// Declare app level module which depends on views, and components
angular.module('ccj16reg', [
  'ngRoute',
  'ngMaterial',
  'ccj16reg.registration',
  'ccj16reg.view.register',
]).
config(['$routeProvider', function($routeProvider) {
  $routeProvider.otherwise({redirectTo: '/register'});
}]);
