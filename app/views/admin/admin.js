angular.module("ccj16reg.view.admin", ["ngRoute", "ngMaterial", "ccj16reg.authentication", "ccj16reg.common"])

.config(function($routeProvider, resolveLoginRequired) {
	"use strict";
	$routeProvider.when("/admin/", {
		templateUrl: "views/admin/admin.html",
		controller: "AdminCtrl",
		resolve: {
			checkAuth: resolveLoginRequired,
		},
	});
})

.controller("AdminCtrl", function() {
	"use strict";
});
