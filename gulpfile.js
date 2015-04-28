'use strict';

var gulp = require('gulp'),
    plugins = require('gulp-load-plugins')(),
    path = function (dir) {
      return ('admin/views/assets/' + (dir || ''));
    },
    scripts = {
      src: [
        path('javascripts/app/components/*.js'),
        path('javascripts/app/*.js')
      ],
      dest: path('javascripts'),
      all: path('javascripts/app/**/*.js'),
      lib: path('javascripts/lib')
    },
    styles = {
      src: path('stylesheets/scss/*.scss'),
      dest: path('stylesheets'),
      css: path('stylesheets/*.css'),
      all: path('stylesheets/scss/**/*.scss'),
      lib: path('stylesheets/lib'),
      fonts: path('stylesheets/fonts')
    };

gulp.task('jshint', function () {
  return gulp.src(scripts.src)
  .pipe(plugins.jshint(path('javascripts/app/.jshintrc')))
  .pipe(plugins.jshint.reporter('default'));
});

gulp.task('jscs', function () {
  return gulp.src(scripts.src)
  .pipe(plugins.jscs(path('javascripts/app/.jscsrc')));
});

gulp.task('js', ['jshint', 'jscs'], function () {
  return gulp.src(scripts.src)
  .pipe(plugins.concat('app.js'))
  .pipe(plugins.uglify())
  .pipe(gulp.dest(scripts.dest));
});

gulp.task('concat', function () {
  return gulp.src(scripts.src)
  .pipe(plugins.sourcemaps.init())
  .pipe(plugins.concat('app.js'))
  .pipe(plugins.sourcemaps.write('./'))
  .pipe(gulp.dest(scripts.dest));
});

gulp.task('jslib', function () {
  return gulp.src([
    'bower_components/jquery/dist/jquery.min.js',
    'bower_components/bootstrap/dist/js/bootstrap.min.js',
    'bower_components/redactor/redactor.min.js',
    'bower_components/cropper/dist/cropper.min.js',
    'bower_components/chosen/chosen.jquery.min.js'
  ])
  .pipe(gulp.dest(scripts.lib));
});

gulp.task('sass', function () {
  return gulp.src(styles.src)
  .pipe(plugins.sourcemaps.init())
  .pipe(plugins.sass())
  .pipe(plugins.sourcemaps.write('./'))
  .pipe(gulp.dest(styles.dest));
});

gulp.task('csslint', ['sass'], function () {
  return gulp.src(styles.css)
  .pipe(plugins.csslint(path('stylesheets/scss/.csslintrc')))
  .pipe(plugins.csslint.reporter());
});

gulp.task('css', ['csslint'], function () {
  return gulp.src(styles.css)
  .pipe(plugins.autoprefixer())
  .pipe(plugins.csscomb())
  .pipe(plugins.minifyCss())
  .pipe(gulp.dest(styles.dest));
});

gulp.task('fonts', function () {
  return gulp.src([
    'bower_components/bootstrap/fonts/*'
  ])
  .pipe(gulp.dest(styles.fonts));
});

gulp.task('redactor', function () {
  return gulp.src('bower_components/redactor/redactor.css')
  .pipe(plugins.rename('redactor.min.css'))
  .pipe(plugins.minifyCss())
  .pipe(gulp.dest('bower_components/redactor'));
});

gulp.task('csslib', ['fonts', 'redactor'], function () {
  return gulp.src([
    'bower_components/bootstrap/dist/css/bootstrap.min.css',
    'bower_components/redactor/redactor.min.css',
    'bower_components/cropper/dist/cropper.min.css',
    'bower_components/chosen/chosen-sprite.png',
    'bower_components/chosen/chosen-sprite@2x.png',
    'bower_components/chosen/chosen.min.css'
  ])
  .pipe(gulp.dest(styles.lib));
});

gulp.task('watch', function () {
  gulp.watch(scripts.all, ['concat']);
  gulp.watch(styles.all, ['sass']);
});

gulp.task('release', ['js', 'css']);

gulp.task('default', ['watch']);
