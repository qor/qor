'use strict';

var gulp = require('gulp'),
    plugins = require('gulp-load-plugins')(),
    scripts = {
      src: ['admin/views/assets/javascripts/components/*.js'],
      dest: 'admin/views/assets/javascripts',
      lib: 'admin/views/assets/javascripts/lib',
      main: 'admin/views/assets/javascripts/main.js'
    },
    styles = {
      src: ['admin/views/assets/stylesheets/scss/**/*.scss'],
      dest: 'admin/views/assets/stylesheets',
      lib: 'admin/views/assets/stylesheets/lib',
      main: 'admin/views/assets/stylesheets/scss/main.scss'
    };

gulp.task('jshint', function () {
  return gulp.src(scripts.src)
  .pipe(plugins.jshint('admin/views/assets/javascripts/.jshintrc'))
  .pipe(plugins.jshint.reporter('default'));
});

gulp.task('jscs', function () {
  return gulp.src(scripts.src)
  .pipe(plugins.jscs('admin/views/assets/javascripts/.jscsrc'));
});

gulp.task('js', ['jshint', 'jscs'], function () {
  return gulp.src(scripts.src)
  .pipe(plugins.concat('main.js'))
  .pipe(plugins.uglify())
  .pipe(gulp.dest(scripts.dest));
});

gulp.task('concat', function () {
  return gulp.src(scripts.src)
  .pipe(plugins.sourcemaps.init())
  .pipe(plugins.concat('main.js'))
  .pipe(plugins.sourcemaps.write('./'))
  .pipe(gulp.dest(scripts.dest))
});

gulp.task('jslib', function () {
  return gulp.src([
    'bower_components/jquery/dist/jquery.js',
    'bower_components/jquery/dist/jquery.min.js',
    'bower_components/bootstrap/dist/js/bootstrap.js',
    'bower_components/bootstrap/dist/js/bootstrap.min.js',
    'bower_components/select2/dist/js/select2.min.js'
  ])
  .pipe(gulp.dest(scripts.lib))
});

gulp.task('css', function () {
  return gulp.src(styles.main)
  .pipe(plugins.sass())
  .pipe(plugins.csslint())
  .pipe(plugins.autoprefixer())
  .pipe(plugins.minifyCss())
  .pipe(gulp.dest(styles.dest));
});

gulp.task('sass', function () {
  return gulp.src(styles.main)
  .pipe(plugins.sass())
  .pipe(gulp.dest(styles.dest))
});

gulp.task('fonts', function () {
  return gulp.src([
    'bower_components/bootstrap/fonts/*',
  ])
  .pipe(gulp.dest('admin/views/assets/fonts'))
});

gulp.task('csslib', ['fonts'], function () {
  return gulp.src([
    'bower_components/bootstrap/dist/css/bootstrap.css',
    'bower_components/bootstrap/dist/css/bootstrap.css.map',
    'bower_components/bootstrap/dist/css/bootstrap.min.css',
    'bower_components/select2/dist/css/select2.min.css'
  ])
  .pipe(gulp.dest(styles.lib))
});

gulp.task('watch', function () {
  gulp.watch(scripts.src, ['concat']);
  gulp.watch(styles.src, ['sass']);
});

gulp.task('default', ['watch']);
