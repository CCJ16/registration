angular.module("ccj16reg.view.recordlist", ["ngRoute", "ccj16reg.registration", "ccj16reg.common"])

.config(function($routeProvider, resolveLoginRequired) {
	"use strict";
	$routeProvider.when("/admin/recordlist", {
		templateUrl: "views/recordlist/list.html",
		controller: "AllRecordListCtrl",
		resolve: {
			checkAuth: resolveLoginRequired,
		},
	});
	$routeProvider.when("/admin/registeredlist", {
		templateUrl: "views/recordlist/list.html",
		controller: "RegisteredRecordListCtrl",
		resolve: {
			checkAuth: resolveLoginRequired,
		},
	});
	$routeProvider.when("/admin/waitinglist", {
		templateUrl: "views/recordlist/waitinglist.html",
		controller: "WaitingRecordListCtrl",
		resolve: {
			checkAuth: resolveLoginRequired,
		},
	});
})

.controller("AllRecordListCtrl", function($scope, Registration) {
	"use strict";
	$scope.registrations = Registration.query({select: "all"});
})

.controller("RegisteredRecordListCtrl", function($scope, Registration) {
	"use strict";
	$scope.registrations = Registration.query({select: "registered"});
})

.controller("WaitingRecordListCtrl", function($scope, Registration) {
	"use strict";
	$scope.registrations = Registration.query({select: "waiting"});
});
