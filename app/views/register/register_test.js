"use strict";

describe("In the register view module", function() {

	beforeEach(module("ccj16reg.view.register"));

	describe("with the register controller", function(){

		it("should be creatable", inject(function($controller, $rootScope) {
			var RegisterCtrl = $controller("RegisterCtrl", {"$scope": $rootScope });
			expect(RegisterCtrl).toBeDefined();
		}));

	});
});
