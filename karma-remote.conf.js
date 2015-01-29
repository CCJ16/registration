module.exports = function(config){
	if (!process.env.SAUCE_USERNAME || !process.env.SAUCE_ACCESS_KEY) {
		console.log('Make sure the SAUCE_USERNAME and SAUCE_ACCESS_KEY environment variables are set.');
		process.exit(1);
	}
	var customLaunchers = {
		'SL_Chrome': {
			base: 'SauceLabs',
			browserName: 'chrome',
			version: '39',
		},
		'SL_Firefox': {
			base: 'SauceLabs',
			browserName: 'firefox',
			version: '35',
		},
		'SL_Safari': {
			base: 'SauceLabs',
			browserName: 'safari',
			platform: 'OS X 10.10',
			version: '8',
		},
		'SL_iOS': {
			base: 'SauceLabs',
			browserName: 'iphone',
			platform: 'OS X 10.10',
			version: '8.1',
			'device-orientation': 'portrait',
		},
		'SL_Android': {
			base: 'SauceLabs',
			browserName: 'android',
			version: '4.4',
		},
		'SL_IE': {
			base: 'SauceLabs',
			browserName: 'internet explorer',
		},
	};

	config.set({
		basePath : './',

		files : [
			'app/bower_components/angular/angular.js',
			'app/bower_components/hammerjs/hammer.js',
			'app/bower_components/angular-*/angular-*.js',
			'app/components/**/*.js',
			'app/views/**/*.js'
		],

		singleRun : true,

		frameworks: ['jasmine'],

		browsers: Object.keys(customLaunchers),
		customLaunchers: customLaunchers,
		sauceLabs: {
			testName: 'CCJ16 Registration',
			recordScreenshots: false,
			connectOptions: {
				port: 5757,
				logfile: 'sauce_connect.log'
			}
		},
		captureTimeout: 0,
		browserDisconnectTimeout: 10000,
		browserDisconnectTolerance: 2,
		browserNoActivityTimeout: 30000,

		plugins : [
			'karma-sauce-launcher',
			'karma-jasmine',
			'karma-junit-reporter'
		],

		reporters: ['progress', 'saucelabs'],
		junitReporter : {
			outputFile: 'test_out/unit.xml',
			suite: 'unit'
		}
	});
};
