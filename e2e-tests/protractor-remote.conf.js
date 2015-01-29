exports.config = {
	sauceUser: process.env.SAUCE_USERNAME,
	sauceKey: process.env.SAUCE_ACCESS_KEY,

	allScriptsTimeout: 11000,

	specs: [
		'*.js'
	],

	multiCapabilities: [
		{'browserName': 'chrome'},
		{'browserName': 'firefox', version: 34},
		{'browserName': 'internet explorer'},
		{'browserName': 'safari'},
	],

	baseUrl: 'http://localhost:9090/',

	framework: 'jasmine',

	jasmineNodeOpts: {
		defaultTimeoutInterval: 30000
	}
};
