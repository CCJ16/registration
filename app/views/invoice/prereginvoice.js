angular.module("ccj16reg.view.invoice", ["ngRoute", "ngMaterial", "ccj16reg.invoice"])

.config(["$routeProvider", function($routeProvider) {
	"use strict";
	$routeProvider.when("/registration/:securityKey/invoice", {
		templateUrl: "views/invoice/invoice.html",
		controller: "InvoiceCtrl",
		resolve: {
			"invoiceData": function($route, invoice) {
				return invoice.getPreregistration({securityKey: $route.current.params.securityKey}).$promise;
			},
		},
	});
}])

.controller("InvoiceCtrl", function($scope, invoiceData) {
	"use strict";
	$scope.invoice = invoiceData;
});
