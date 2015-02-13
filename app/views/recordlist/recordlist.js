angular.module("ccj16reg.view.recordlist", ["ngRoute", "ccj16reg.registration", "ccj16reg.common"])

.config(function($routeProvider, resolveLoginRequired) {
	"use strict";
	$routeProvider.when("/admin/recordlist", {
		templateUrl: "views/recordlist/list.html",
		controller: "RecordListCtrl",
		resolve: {
			checkAuth: resolveLoginRequired,
		},
	});
})

.controller("RecordListCtrl", function($scope, Registration) {
	"use strict";
	$scope.registrations = Registration.query();
});
