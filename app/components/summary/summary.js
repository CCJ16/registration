angular.module("ccj16reg.summary", ["ngResource"])
.factory("summary", function($resource) {
	"use strict";
	var r = $resource("/api/summary/", null, {
		"getPack": { method: "GET", url: "/api/summary/pack" },
	})

	delete r.save
	delete r.query
	delete r.delete
	delete r.remove
	delete r.get

	delete r.prototype.$save
	delete r.prototype.$delete
	delete r.prototype.$removes

	return r
})
