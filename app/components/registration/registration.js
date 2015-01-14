'use strict';

angular.module('ccj16reg.registration', ['ngResource'])
.factory('registration', function($resource, currentDateFetch) {
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
		return angular.isDefined(this.emailApprovalGivenAt);
	}
	return res;
})
.value('currentDateFetch', function() {
	return new Date();
});
