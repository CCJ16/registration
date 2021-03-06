angular.module("ccj16reg.view.invoice", ["ngRoute", "ngMaterial", "ccj16reg.invoice", "ccj16reg.moment"])

.config(["$routeProvider", function($routeProvider) {
	"use strict";
	$routeProvider.when("/registration/:securityKey/invoice", {
		templateUrl: "views/invoice/invoice.html",
		controller: "InvoiceCtrl",
		resolve: {
			"invoiceData": function($route, Invoice) {
				return Invoice.getPreregistration({securityKey: $route.current.params.securityKey}).$promise;
			},
		},
	});
}])

.controller("InvoiceCtrl", function($scope, invoiceData) {
	"use strict";
	$scope.invoice = invoiceData;
});
