'use strict';

angular.module('ccj16reg.registration', [])
.factory('registration', ['$q', '$http', function($q, $http) {
	return {
		new: function() {
			return {
				$$unsaved: true,

				save: function() {
					var obj = this;
					if (this.$$unsaved) {
						return $q(function(resolve, reject) {
							$http.post('/api/preregistration', angular.toJson(obj)).success(function(data, status) {
								if (status == 201) {
									resolve();
								} else {
									reject("Failed to create object.");
								}
							}).error(function() {
								reject("Failed to save object.");
							})
						})
					} else {
						console.log("IMPLEMENT OTHER SAVE");
					}
				},
			}
		}
	}
}]);
