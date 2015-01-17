'use strict';

angular.module('ccj16reg.registration', ['ngResource'])
.factory('registration', function($resource, $http, currentDateFetch) {
	var res = $resource('/api/preregistration/:securityKey');
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
	res.confirmEmail = function(email, token) {
		return $http.put('/api/confirmpreregistration?email=' + email, token);
	}
	return res;
})
.value('currentDateFetch', function() {
	return new Date().toISOString();
});
