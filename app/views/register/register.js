'use strict';

angular.module('ccj16reg.view.register', ['ngRoute'])

.config(['$routeProvider', function($routeProvider) {
  $routeProvider.when('/register', {
    templateUrl: 'views/register/register.html',
    controller: 'RegisterCtrl'
  });
}])

.controller('RegisterCtrl', [function() {

}]);