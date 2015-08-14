'use strict';

var gulp = require('gulp');
var plugins = require('gulp-load-plugins')();
var moduleName = (function () {
      var args = process.argv;
      var length = args.length;
      var i = 0;
      var matched;
      var name;

      while (i++ < length) {
        matched = String(args[i]).match(/^--+(\w+)$/i);

        if (matched) {
          name = matched[1];
          break;
        }
      }

      return name;
    })();


// Admin
// Command: gulp [task]
// -----------------------------------------------------------------------------

function adminTasks() {
  var pathto = function (file) {
        return ('admin/views/assets/' + file);
      };
  var scripts = {
        src: pathto('javascripts/app/*.js'),
        dest: pathto('javascripts'),
        qor: pathto('javascripts/qor/*.js'),
        all: [
          'gulpfile.js',
          pathto('javascripts/qor/*.js'),
          pathto('javascripts/app/*.js'),
        ],
      };
  var styles = {
        src: pathto('stylesheets/scss/*.scss'),
        dest: pathto('stylesheets'),
        main: pathto('stylesheets/{qor,app}.css'),
        scss: pathto('stylesheets/scss/**/*.scss'),
      };
  var fonts = {
        dest: pathto('fonts'),
      };

  gulp.task('jshint', function () {
    return gulp.src(scripts.all)
    .pipe(plugins.jshint())
    .pipe(plugins.jshint.reporter('default'));
  });

  gulp.task('jscs', function () {
    return gulp.src(scripts.all)
    .pipe(plugins.jscs());
  });

  gulp.task('qor', ['jshint', 'jscs'], function () {
    return gulp.src(scripts.qor)
    .pipe(plugins.concat('qor.js'))
    .pipe(plugins.uglify())
    .pipe(gulp.dest(scripts.dest));
  });

  gulp.task('js', ['qor'], function () {
    return gulp.src(scripts.src)
    .pipe(plugins.concat('app.js'))
    .pipe(plugins.uglify())
    .pipe(gulp.dest(scripts.dest));
  });

  gulp.task('qor+', function () {
    return gulp.src(scripts.qor)
    .pipe(plugins.sourcemaps.init())
    .pipe(plugins.concat('qor.js'))
    .pipe(plugins.sourcemaps.write('./'))
    .pipe(gulp.dest(scripts.dest));
  });

  gulp.task('js+', function () {
    return gulp.src(scripts.src)
    .pipe(plugins.sourcemaps.init())
    .pipe(plugins.concat('app.js'))
    .pipe(plugins.sourcemaps.write('./'))
    .pipe(gulp.dest(scripts.dest));
  });

  gulp.task('jslib', function () {
    return gulp.src([
      'bower_components/jquery/dist/jquery.min.js',
      'bower_components/jquery/dist/jquery.min.map',
      'bower_components/material-design-lite/material.min.js',
      'bower_components/material-design-lite/material.min.js.map',
      'bower_components/cropper/dist/cropper.min.js',
      'bower_components/chosen/chosen.jquery.min.js'
    ])
    .pipe(gulp.dest(scripts.dest));
  });

  gulp.task('sass', function () {
    return gulp.src(styles.src)
    .pipe(plugins.sourcemaps.init())
    .pipe(plugins.sass())
    .pipe(plugins.sourcemaps.write('./'))
    .pipe(gulp.dest(styles.dest));
  });

  gulp.task('csslint', ['sass'], function () {
    return gulp.src(styles.main)
    .pipe(plugins.csslint('.csslintrc'))
    .pipe(plugins.csslint.reporter());
  });

  gulp.task('css', ['csslint'], function () {
    return gulp.src(styles.main)
    .pipe(plugins.autoprefixer())
    .pipe(plugins.csscomb())
    .pipe(plugins.minifyCss())
    .pipe(gulp.dest(styles.dest));
  });

  gulp.task('fonts', function () {
    return gulp.src([
      'bower_components/material-design-icons/iconfont/codepoints',
      'bower_components/material-design-icons/iconfont/MaterialIcons*'
    ])
    .pipe(gulp.dest(fonts.dest));
  });

  gulp.task('csslib', ['fonts'], function () {
    return gulp.src([
      'bower_components/material-design-lite/material.min.css',
      'bower_components/material-design-lite/material.min.css.map',
      'bower_components/cropper/dist/cropper.min.css',
      'bower_components/chosen/chosen-sprite.png',
      'bower_components/chosen/chosen-sprite@2x.png',
      'bower_components/chosen/chosen.min.css'
    ])
    .pipe(gulp.dest(styles.dest));
  });

  gulp.task('watch', function () {
    gulp.watch(scripts.qor, ['qor+']);
    gulp.watch(scripts.src, ['js+']);
    gulp.watch(styles.scss, ['sass']);
  });

  gulp.task('lib', ['jslib', 'csslib']);
  gulp.task('release', ['js', 'css']);

  gulp.task('default', ['watch']);
};


// Modules
// Command: gulp [task] --moduleName
// -----------------------------------------------------------------------------

function moduleTasks(moduleName) {
  var pathto = function (file) {
        return (moduleName + '/views/themes/' + moduleName + '/assets/' + file);
      };
  var scripts = {
        src: pathto('javascripts/app/*.js'),
        dest: pathto('javascripts/'),
      };
  var styles = {
        src: pathto('stylesheets/scss/*.scss'),
        dest: pathto('stylesheets/'),
        main: pathto('stylesheets/' + moduleName + '.css'),
        scss: pathto('stylesheets/scss/**/*.scss'),
      };

  gulp.task('jshint', function () {
    return gulp.src(scripts.src)
    .pipe(plugins.jshint())
    .pipe(plugins.jshint.reporter('default'));
  });

  gulp.task('jscs', function () {
    return gulp.src(scripts.src)
    .pipe(plugins.jscs());
  });

  gulp.task('js', function () {
    return gulp.src(scripts.src)
    .pipe(plugins.concat(moduleName + '.js'))
    .pipe(plugins.uglify())
    .pipe(gulp.dest(scripts.dest));
  });

  gulp.task('concat', function () {
    return gulp.src(scripts.src)
    .pipe(plugins.sourcemaps.init())
    .pipe(plugins.concat(moduleName + '.js'))
    .pipe(plugins.sourcemaps.write('./'))
    .pipe(gulp.dest(scripts.dest));
  });

  gulp.task('sass', function () {
    return gulp.src(styles.src)
    .pipe(plugins.sass())
    .pipe(gulp.dest(styles.dest));
  });

  gulp.task('csslint', ['sass'], function () {
    return gulp.src(styles.main)
    .pipe(plugins.csslint('.csslintrc'))
    .pipe(plugins.csslint.reporter());
  });

  gulp.task('css', ['csslint'], function () {
    return gulp.src(styles.main)
    .pipe(plugins.autoprefixer())
    .pipe(plugins.csscomb())
    .pipe(plugins.minifyCss())
    .pipe(gulp.dest(styles.dest));
  });

  gulp.task('watch', function () {
    gulp.watch(scripts.src, ['concat']);
    gulp.watch(styles.scss, ['sass']);
  });

  gulp.task('release', ['js', 'css']);

  gulp.task('default', ['watch']);
};


// Init
// -----------------------------------------------------------------------------

if (moduleName) {
  console.log('Running "' + moduleName + '" task...');
  moduleTasks(moduleName);
} else {
  console.log('Running "admin" task...');
  adminTasks();
}
