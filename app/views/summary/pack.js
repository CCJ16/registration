angular.module("ccj16reg.view.summary.pack", ["ngRoute", "ngMaterial", "ccj16reg.summary"])

.config(function($routeProvider) {
	"use strict";
	$routeProvider.when("/summary/pack", {
		templateUrl: "views/summary/pack.html",
		controller: "SummaryPackCtrl",
		resolve: {
			packSummary: function(summary) {
				return summary.getPack()
			}
		},
	});
})

.controller("SummaryPackCtrl", function($scope, packSummary) {
	"use strict";
	$scope.summary = packSummary
});
