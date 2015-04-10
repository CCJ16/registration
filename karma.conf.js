"use strict";

module.exports = function(config){
	config.set({
		basePath : "./",

		files : [
			"app/bower_components/angular/angular.js",
			"app/bower_components/angular-*/angular-*.js",
			"app/bower_components/moment/moment.js",
			"app/bower_components/moment-timezone/builds/moment-timezone-with-data.js",
			"app/components/**/*.js",
			"app/views/**/*.js"
		],

		autoWatch : true,

		frameworks: ["jasmine"],

		browsers : ["Chrome", "Firefox"],

		plugins : [
			"karma-chrome-launcher",
			"karma-firefox-launcher",
			"karma-jasmine",
			"karma-notify-send-reporter",
			"karma-junit-reporter"
		],

		junitReporter : {
			outputFile: "test_out/unit.xml",
			suite: "unit"
		},

		reporters : [
			"notify-send",
			"progress"
		]
	});
};
