angular.module("ccj16reg.view.recordlist", ["ngRoute", "ngMaterial", "ccj16reg.registration", "ccj16reg.common"])

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

.controller("WaitingRecordListCtrl", function($scope, $mdDialog, Registration) {
	"use strict";
	$scope.registrations = Registration.query({select: "waiting"});
	$scope.promote = function(index, ev) {
		var reg = $scope.registrations[index]

		$mdDialog.show({
			templateUrl: "views/recordlist/pending_promote.html",
			targetEvent: ev,
			clickOutsideToClose: false,
			escapeToClose: false,
			onComplete: submitPromotion,
		});

		function submitPromotion() {
			reg.promote().then(function() {
				$mdDialog.hide()
				$scope.registrations.splice(index, 1)
			}, function(msg) {
				$mdDialog.hide()
				$mdDialog.show(
					$mdDialog.alert()
						.title("Failed to promote group")
						.content("Server message: " + msg.data)
						.ok("OK")
				)
			})
		}
	}
});
