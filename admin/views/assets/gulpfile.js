'use strict';

var gulp = require('gulp'),
    sass = require('gulp-sass'),
    jshint = require('gulp-jshint'),
    uglify = require('gulp-uglify'),
    minifycss = require('gulp-minify-css'),
    imagemin = require('gulp-imagemin'),
    concat = require('gulp-concat'),
    notify = require('gulp-notify'),
    cache = require('gulp-cache'),
    livereload = require('gulp-livereload'),
    del = require('del');

var dist = './';

gulp.task('css', function() {
  return gulp.src('stylesheets/scss/app/*.sass')
  .pipe(sass({
    outputStyle: 'compressed',
    sourceComments: 'normal',
    errLogToConsole: true,
    indentedSyntax: true
  }))
  .pipe(minifycss())
  .pipe(gulp.dest(dist + 'stylesheets/'))
  // .pipe(notify({ message: 'Stylesheets task complete' }));
});

gulp.task('js', function() {
  return gulp.src(['javascripts/jquery-2.1.3.min.js', 'javascripts/lib/*.js', 'javascripts/app/*.js'])
  // .pipe(jshint('.jshintrc'))
  .pipe(jshint.reporter('default'))
  .pipe(concat('app.js'))
  .pipe(uglify())
  .pipe(gulp.dest(dist + 'javascripts/'))
  .pipe(notify({ message: 'Javascripts task complete' }));
});

gulp.task('img', function() {
  return gulp.src('images/*')
  .pipe(imagemin({ optimizationLevel: 3, progressive: true, interlaced: true }))
  .pipe(gulp.dest(dist + 'images/'))
  .pipe(notify({ message: 'Images task complete' }));
});

gulp.task('watch', function() {
  gulp.watch('stylesheets/scss/app/*.sass', ['css']);
  // gulp.watch('javascripts/**/*.js', ['js']);
  gulp.watch('images/**/*', ['img']);
});

gulp.task('sass', function() {
  return gulp.watch('stylesheets/**/*.scss', ['css'])
  .pipe(notify({ message: 'sass watch start' }));
});

gulp.task('hint', function () {
  return gulp.src(['javascripts/*.js', 'javascripts/test/*.js'])
  // .pipe(jshint('.jshintrc'))
  .pipe(jshint())
  .pipe(jshint.reporter('default'));
});

gulp.task('test', function () {
  return gulp.src('javascripts/test/*.js', { read: false })
  .pipe(mocha());
});

gulp.task('clean', function (cb) {
  return del(['./public/css', './public/js', './public/img'], cb);
});

// live reload

gulp.task('default', ['clean'], function() {
  gulp.start('watch');
});
