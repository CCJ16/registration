"use strict"

var page = new (require("./po/summarypack"))()

describe("Pack summary page", function() {
	describe("with no pack information submitted", function() {
		it("should display 0 for both", function() {
			page.get()

			expect(page.youthCount).toBe("0")
			expect(page.leaderCount).toBe("0")
		})
	})

	describe("with two packs reporting", function() {
		beforeEach(function() {
			var flow = protractor.promise.controlFlow()
			flow.execute(function() {
				var p1 = protractor.promise.defer()
				browser.executeAsyncScript(function(callback) {
					var Registration = angular.injector(["ccj16reg"]).get("Registration")
					var reg = new Registration()
					reg.council = "Test council"
					reg.groupName = "Test group"
					reg.contactLeaderFirstName = "FirstN"
					reg.contactLeaderLastName = "LastN"
					reg.contactLeaderEmail = "test@invalid"
					reg.contactLeaderPhoneNumber= "123-456-7890"
					reg.estimatedYouth = 20
					reg.estimatedLeaders = 12
					reg.agreedToEmailTerms(true)

					reg.$save().then(function() {
						callback(true)
					}, function() {
						callback(false)
					})
				}).then(function(result) {
					if (result === true) {
						p1.fulfill()
					} else {
						p1.reject()
					}
				})
				var p2 = protractor.promise.defer()
				browser.executeAsyncScript(function(callback) {
					var Registration = angular.injector(["ccj16reg"]).get("Registration")
					var reg = new Registration()
					reg.council = "Test council 2"
					reg.groupName = "Test group 2"
					reg.contactLeaderFirstName = "FirstN"
					reg.contactLeaderLastName = "LastN"
					reg.contactLeaderEmail = "test2@invalid"
					reg.contactLeaderPhoneNumber= "123-456-7890"
					reg.estimatedYouth = 4
					reg.estimatedLeaders = 3
					reg.agreedToEmailTerms(true)

					reg.$save().then(function() {
						callback(true)
					}, function() {
						callback(false)
					})
				}).then(function(result) {
					if (result === true) {
						p2.fulfill()
					} else {
						p2.reject()
					}
				})
				return protractor.promise.all(p1.promise, p2.promise)
			})
		})
		it("should display correct totals", function() {
			page.get()

			expect(page.youthCount).toBe("24")
			expect(page.leaderCount).toBe("15")
		})
	})
})
