gulp       = require 'gulp'
lazy       = require 'lazypipe'
watch      = require 'gulp-watch'
plumber    = require 'gulp-plumber'
notify     = require 'gulp-notify'
gulpif     = require 'gulp-if'
filter     = require 'gulp-filter'
debug      = require 'gulp-debug'
sourcemaps = require 'gulp-sourcemaps'
stylus     = require 'gulp-stylus'
pug        = require 'gulp-pug'
coffee     = require 'gulp-coffee'
htmlmin    = require 'gulp-htmlmin'
uglify     = require 'gulp-uglify-es'
babel      = require 'gulp-babel'
prefix     = require 'gulp-autoprefixer'
csso       = require 'gulp-csso'
fontmin    = require 'gulp-fontmin'
del        = require 'del'

process = require("process")
argv    = require("minimist")(process.argv)

gulp.task 'clean', -> del "dist"

if gulp.parallel
	clean = gulp.parallel 'clean'
else
	clean = ['clean']

# TODO: better js min
builders = (w) ->
	f = null

	if w
		f = -> watch(arguments...).pipe(plumber((err) ->
			notify.onError({
				title:   "Gulp Error"
				message: "Error: <%= error.message %>"
			})(err)
		))
	else
		f = -> gulp.src arguments...

	f "app/**/*.styl"
		.pipe stylus()
		.pipe csso()
		.pipe prefix()
		.pipe debug title: "Compiled"
		.pipe gulp.dest "dist"

	# not wrap code into "do ->" for debug purposes
	if argv.production
		f "app/**/*.coffee"
			.pipe coffee()
			.pipe babel presets: ['env']
			.pipe uglify.default()
			.pipe debug title: "Compiled"
			.pipe gulp.dest "dist"
	else
		f "app/**/*.coffee"
			.pipe coffee bare: true
			.pipe debug title: "Compiled"
			.pipe gulp.dest "dist"

	f "app/**/*.js"
		.pipe gulp.dest "dist"

	f "app/**/*.pug"
		.pipe pug()
		.pipe htmlmin collapseWhitespace: true
		.pipe debug title: "Compiled"
		.pipe gulp.dest "dist"

	f "app/**/*.ttf"
		.pipe fontmin()
		.pipe filter ["**/*.ttf"]
		.pipe debug title: "Compiled"
		.pipe gulp.dest "dist"


gulp.task 'build', clean, ->
	builders false

if gulp.parallel
	build = gulp.parallel 'build'
else
	build = ['build']


gulp.task 'watch', build, ->
	builders true

gulp.task 'default', build
