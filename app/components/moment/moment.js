angular.module("ccj16reg.moment", [])
.filter("moment", function() {
	"use strict";
	return function(input, formatString, timezone) {
		return moment(input).tz(timezone).format(formatString)
	}
})
