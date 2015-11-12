angular.module("ccj16reg.registration", ["ngResource"])
.factory("Registration", function($resource, $http, $q, currentDateFetch) {
	"use strict";
	var res = $resource("/api/preregistration/:securityKey");
	res.prototype.agreedToEmailTerms = function(checked) {
		if (angular.isDefined(checked)) {
			if(checked) {
				if (!angular.isDefined(this.emailApprovalGivenAt)) {
					this.emailApprovalGivenAt = currentDateFetch();
				}
			} else {
				this.emailApprovalGivenAt = undefined;
			}
		}
		return angular.isDefined(this.emailApprovalGivenAt) && this.emailApprovalGivenAt !== "0001-01-01T00:00:00Z";
	}
	res.prototype.validatedEmail = function() {
		return angular.isDefined(this.validatedOn) && this.validatedOn !== "0001-01-01T00:00:00Z";
	}
	res.prototype.promote = function() {
		var reg = this
		return $q(function(resolve, reject) {
			if (!reg.isOnWaitingList) {
				reject("Group is not on the waiting list")
			} else {
				$http.post("/api/preregistration/" + reg.securityKey + "/promote").then(function() {
					resolve()
				}, function(resp) {
					reject(resp.data)
				})
			}
		})
	}
	res.confirmEmail = function(email, token) {
		return $http.put("/api/confirmpreregistration?email=" + email, token);
	}
	return res;
})
.value("currentDateFetch", function() {
	"use strict";
	return new Date().toISOString();
});
