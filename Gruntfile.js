// Generated on 2015-02-09 using generator-angular 0.11.0
"use strict";

// # Globbing
// for performance reasons we're only matching one level down:
// "test/spec/{,*/}*.js"
// use this if you want to recursively match all subfolders:
// "test/spec/**/*.js"

module.exports = function (grunt) {

	// Load grunt tasks automatically
	require("load-grunt-tasks")(grunt);

	// Time how long tasks take. Can help when optimizing build times
	require("time-grunt")(grunt);

	// Configurable paths for the application
	var appConfig = {
		app: require("./bower.json").appPath || "app",
		dist: "dist"
	};

	// Define the configuration for all the tasks
	grunt.initConfig({

		// Project settings
		yeoman: appConfig,

		// Watches files for changes and runs tasks based on the changed files
		watch: {
			js: {
				files: [
					"<%= yeoman.app %>/**/*.js",
					"e2e-tests/**/*.js",
					"*.js"
				],
				tasks: ["newer:jshint"]
			},
			gruntfile: {
				files: ["Gruntfile.js"]
			},
		},

		// Make sure code styles are up to par and there are no obvious mistakes
		jshint: {
			options: {
				jshintrc: ".jshintrc",
				reporter: require("jshint-stylish")
			},
			all: {
				options: {
					ignores: "**/*_test.js"
				},
				src: [
					"<%= yeoman.app %>/components/*/*.js",
					"<%= yeoman.app %>/views/*/*.js",
					"<%= yeoman.app %>/app/*.js"
				],
			},
			test: {
				options: {
					jshintrc: ".jshintrc.test"
				},
				src: [
					"<%= yeoman.app %>/{components,views}/*/*_test.js",
					"<%= yeoman.app %>/app/*_test.js",
					"e2e-tests/**/*.js",
					"Gruntfile.js",
					"karma*.js",
					"protractor*.js"
				]
			}
		},

		// Empties folders to start fresh
		clean: {
			dist: {
				files: [{
					dot: true,
					src: [
						".tmp",
						"<%= yeoman.dist %>/{,*/}*",
						"!<%= yeoman.dist %>/.git{,*/}*"
					]
				}]
			},
			server: ".tmp"
		},

		// Add vendor prefixed styles
		autoprefixer: {
			options: {
				browsers: ["last 1 version"]
			},
			dist: {
				files: [{
					expand: true,
					cwd: ".tmp/styles/",
					src: "**/*.css",
					dest: ".tmp/styles/"
				}]
			}
		},

		// Renames files for browser caching purposes
		filerev: {
			dist: {
				src: [
					"<%= yeoman.dist %>/app/*.{js,css}",
					"<%= yeoman.dist %>/images/{,*/,*/*/}*.{png,jpg,jpeg,gif,webp,svg}",
				]
			}
		},

		// Reads HTML for usemin blocks to enable smart builds that automatically
		// concat, minify and revision files. Creates configurations in memory so
		// additional tasks can operate on them
		useminPrepare: {
			html: "<%= yeoman.app %>/index.html",
			options: {
				dest: "<%= yeoman.dist %>",
				flow: {
					html: {
						steps: {
							js: ["concat", "uglifyjs"],
							css: ["cssmin"]
						},
						post: {}
					}
				}
			}
		},

		// Performs rewrites based on filerev and the useminPrepare configuration
		usemin: {
			html: ["<%= yeoman.dist %>/{,*/}*.html"],
			css: ["<%= yeoman.dist %>/styles/**/*.css"],
			options: {
				assetsDirs: [
					"<%= yeoman.dist %>",
					"<%= yeoman.dist %>/images",
					"<%= yeoman.dist %>/styles"
				]
			}
		},

		imagemin: {
			dist: {
				files: [{
					expand: true,
					cwd: "<%= yeoman.app %>/images",
					src: "{,*/}*.{png,jpg,jpeg,gif}",
					dest: "<%= yeoman.dist %>/images"
				}]
			}
		},

		svgmin: {
			dist: {
				files: [{
					expand: true,
					cwd: "<%= yeoman.app %>/images",
					src: "{,*/}*.svg",
					dest: "<%= yeoman.dist %>/images"
				}]
			}
		},

		htmlmin: {
			dist: {
				options: {
					collapseWhitespace: true,
					conservativeCollapse: true,
					collapseBooleanAttributes: true,
					removeCommentsFromCDATA: true,
					removeOptionalTags: true
				},
				files: [{
					expand: true,
					cwd: "<%= yeoman.dist %>",
					src: ["*.html", "views/{,*/}*.html"],
					dest: "<%= yeoman.dist %>"
				}]
			}
		},

		// ng-annotate tries to make the code safe for minification automatically
		// by using the Angular long form for dependency injection.
		ngAnnotate: {
			dist: {
				files: [{
					expand: true,
					cwd: ".tmp/concat",
					src: ["**/*.js", "!oldieshim.js"],
					dest: ".tmp/concat"
				}]
			}
		},

		// Copies remaining files to places other tasks can use
		copy: {
			dist: {
				files: [{
					expand: true,
					dot: true,
					cwd: "<%= yeoman.app %>",
					dest: "<%= yeoman.dist %>",
					src: [
						"*.{ico,png,txt}",
						"*.html",
						"views/{,*/}*.html",
					]
				}, {
					expand: true,
					cwd: ".tmp/images",
					dest: "<%= yeoman.dist %>/images",
					src: ["generated/*"]
				}]
			},
			styles: {
				expand: true,
				cwd: "<%= yeoman.app %>/",
				dest: ".tmp/styles/",
				src: [ "bower_components/*/*.css", "*.css", "{components, views}/{,*/}*.css"]
			}
		},

		image_debower: {
			dist: {
				expand: true,
				cwd: "<%= yeoman.dist %>/",
				src: [
					"views/*/*.html",
					"index.html"
				],
				dest: "<%= yeoman.dist %>",
				srcImages: "<%= yeoman.app %>",
				destImages: "<%= yeoman.dist %>/images"
			}
		},

		// Run some tasks in parallel to speed up the build process
		concurrent: {
			server: [
				"copy:styles"
			],
			test: [
				"copy:styles"
			],
			dist: [
				"copy:styles",
				"imagemin",
				"svgmin"
			]
		},

		// Test settings
		karma: {
			unit: {
				configFile: "karma.conf.js",
				singleRun: true
			}
		}
	});

	grunt.registerTask("test", [
		"clean:server",
		"concurrent:test",
		"autoprefixer",
		"karma"
	]);

	grunt.registerTask("build", [
		"clean:dist",
		"useminPrepare",
		"concurrent:dist",
		"autoprefixer",
		"concat",
		"ngAnnotate",
		"copy:dist",
		"cssmin",
		"uglify",
		"filerev",
		"usemin",
		"image_debower",
		"htmlmin"
	]);

	grunt.registerTask("default", [
		"newer:jshint",
		"test",
		"build"
	]);
};
