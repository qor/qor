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

gulp.task('css', function() {
  return gulp.src('stylesheets/sass/*.sass')
  .pipe(sass({
    outputStyle: 'compressed',
    sourceComments: 'normal',
    errLogToConsole: true,
    indentedSyntax: true
  }))
  .pipe(minifycss())
  .pipe(gulp.dest('stylesheets/'))
  .pipe(notify({ message: 'Stylesheets task complete' }));
});

gulp.task('js', function() {
  return gulp.src(['javascripts/lib/jquery-2.1.3.min.js', 'javascripts/lib/*.js', 'javascripts/app/*.js'])
  // .pipe(jshint('.jshintrc'))
  .pipe(jshint.reporter('default'))
  .pipe(concat('bundle.js'))
  .pipe(uglify())
  .pipe(gulp.dest('javascripts/'))
  .pipe(notify({ message: 'Javascripts task complete' }));
});

gulp.task('img', function() {
  return gulp.src('images/**/*')
  .pipe(imagemin({ optimizationLevel: 3, progressive: true, interlaced: true }))
  .pipe(gulp.dest('dist/img'))
  .pipe(notify({ message: 'Images task complete' }));
});

gulp.task('watch', function() {
  gulp.watch('stylesheets/**/*.sass', ['css']);
  gulp.watch('javascripts/**/*.js', ['js']);
  gulp.watch('images/**/*', ['img']);
});

gulp.task('sass', function() {
  return gulp.watch('stylesheets/**/*.sass', ['css'])
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
  return del(['dist/css', 'dist/js', 'dist/img'], cb);
});

// live reload

gulp.task('default', ['clean'], function() {
  gulp.start('watch');
});
