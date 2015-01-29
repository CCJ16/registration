exports.config = {
	allScriptsTimeout: 11000,

	specs: [
		'*.js'
	],

	capabilities: {
		'browserName': 'chrome'
	},

	baseUrl: 'http://localhost:9090/',

	framework: 'jasmine',

	jasmineNodeOpts: {
		defaultTimeoutInterval: 30000
	}
};
