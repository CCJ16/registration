angular.module("ccj16reg.view.register", ["ngRoute", "ngSanitize", "ngMaterial", "ccj16reg.registration"])

.config(function($routeProvider) {
	"use strict";
	$routeProvider.when("/register", {
		templateUrl: "views/register/register.html",
		controller: "RegisterCtrl"
	});
})

.controller("RegisterCtrl", function($scope, $location, $mdDialog, Registration) {
	"use strict";
	$scope.registration = new Registration();

	$scope.$watch("registration.agreedToEmailTerms()", function(checked) {
		$scope.registrationTosAccepted = checked;
	});

	$scope.showEmailTos = function(ev) {
		ev.preventDefault()
		ev.stopPropagation();
		var dialog = $mdDialog.alert()
			.title("Email usage agreement")
			.content("All email addresses received through the registration process have been  added to the <span class=\"title-text\">CCJ'16</span>  registration database as well as the <span class=\"title-text\">CCJ'16</span>  mailing list.  During the coming months, you will receive mails from time to time with information about our upcoming CCJ, planning for attending the camp, and other information related to this camp.  Your address will not be distributed to others, nor used for matters not directly connected with <span class=\"title-text\">CCJ'16</span>.  Should you wish to be removed from this distribution list, please send an email to info@cubjamboree.ca")
			.ok("Done")
			.targetEvent(ev)
		dialog._options.template = [
			"<md-dialog aria-label=\"{{ dialog.ariaLabel }}\">",
			"<md-content>",
				"<h2>{{ dialog.title }}</h2>",
				"<p><span ng-bind-html=\"dialog.content\"></span></p>",
			"</md-content>",
			"<div class=\"md-actions\">",
				"<md-button ng-if=\"dialog.$type == 'confirm'\" ng-click=\"dialog.abort()\">",
				"{{ dialog.cancel }}",
				"</md-button>",
				"<md-button ng-click=\"dialog.hide()\" class=\"md-primary\">",
				"{{ dialog.ok }}",
				"</md-button>",
			"</div>",
			"</md-dialog>",
		].join("")
		$mdDialog.show(dialog);
	}

	$scope.submitRegistration = function(ev) {
		$mdDialog.show({
			templateUrl: "views/register/pending_submit.html",
			targetEvent: ev,
			clickOutsideToClose: false,
			escapeToClose: false,
			onComplete: submitSaveRequest,
		});
		function submitSaveRequest() {
			$scope.registration.$save().then(function(reg) {
				$mdDialog.hide();
				$location.path("/registration/" + reg.securityKey)
			}, function(msg) {
				$mdDialog.hide();
				$mdDialog.show(
					$mdDialog.alert()
						.title("Failed to create registration")
						.content("Server message: " + msg.data)
						.ok("OK")
				);
			});
		}
	};
});
