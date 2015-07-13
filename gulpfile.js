'use strict';

var gulp = require('gulp'),
    plugins = require('gulp-load-plugins')(),
    task = (function (args) {
      var length = args.length,
          i = 0,
          n;

      while (i++ < length) {
        n = String(args[i]).match(/^-*(i18n|l10n|publish)$/i);

        if (n) {
          n = n[1];
          break;
        }
      }

      return n;
    })(process.argv),
    tasks = {};


// Default
// Command: gulp [task]
// -----------------------------------------------------------------------------

tasks.base = function () {
  var pathto = function (file) {
        return ('admin/views/assets/' + file);
      },
      scripts = {
        src: pathto('javascripts/app/*.js'),
        dest: pathto('javascripts'),
        qor: pathto('javascripts/qor/*.js'),
        all: [
          'gulpfile.js',
          pathto('javascripts/qor/*.js'),
          pathto('javascripts/app/*.js')
        ]
      },
      styles = {
        src: pathto('stylesheets/scss/*.scss'),
        dest: pathto('stylesheets'),
        main: pathto('stylesheets/{qor,app}.css'),
        scss: pathto('stylesheets/scss/**/*.scss')
      },
      fonts = {
        dest: pathto('fonts')
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
      'bower_components/bootstrap/dist/js/bootstrap.min.js',
      'bower_components/bootstrap-material-design/dist/js/material.min.js',
      'bower_components/bootstrap-material-design/dist/js/ripples.min.js',
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
      'bower_components/bootstrap/fonts/*',
      'bower_components/bootstrap-material-design/dist/fonts/Material*',
      'bower_components/bootstrap-material-design/dist/fonts/Roboto*'
    ])
    .pipe(gulp.dest(fonts.dest));
  });

  gulp.task('csslib', ['fonts'], function () {
    return gulp.src([
      'bower_components/bootstrap/dist/css/bootstrap.min.css',
      'bower_components/bootstrap-material-design/dist/css/material.min.css',
      'bower_components/bootstrap-material-design/dist/css/ripples.min.css',
      'bower_components/bootstrap-material-design/dist/css/roboto.min.css',
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

  gulp.task('release', ['js', 'css']);

  gulp.task('default', ['watch']);
};


// I18n (Internationalization)
// Command: gulp [task] --i18n
// -----------------------------------------------------------------------------

tasks.i18n = function () {
  var namespace = 'i18n',
      pathto = function (file) {
        return (namespace + '/views/themes/' + namespace + '/assets/' + file);
      },
      scripts = {
        src: pathto('javascripts/app/*.js'),
        dest: pathto('javascripts/')
      },
      styles = {
        src: pathto('stylesheets/scss/*.scss'),
        dest: pathto('stylesheets/'),
        main: pathto('stylesheets/' + namespace + '.css'),
        scss: pathto('stylesheets/scss/**/*.scss')
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
    .pipe(plugins.concat(namespace + '.js'))
    .pipe(plugins.uglify())
    .pipe(gulp.dest(scripts.dest));
  });

  gulp.task('concat', function () {
    return gulp.src(scripts.src)
    .pipe(plugins.sourcemaps.init())
    .pipe(plugins.concat(namespace + '.js'))
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


// L10n (Localization)
// Command: gulp [task] --l10n
// -----------------------------------------------------------------------------

tasks.l10n = function () {
  var namespace = 'l10n',
      pathto = function (file) {
        return (namespace + '/views/themes/' + namespace + '/assets/' + file);
      },
      scripts = {
        src: pathto('javascripts/app/*.js'),
        dest: pathto('javascripts/')
      },
      styles = {
        src: pathto('stylesheets/scss/*.scss'),
        dest: pathto('stylesheets/'),
        main: pathto('stylesheets/' + namespace + '.css'),
        scss: pathto('stylesheets/scss/**/*.scss')
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
    .pipe(plugins.concat(namespace + '.js'))
    .pipe(plugins.uglify())
    .pipe(gulp.dest(scripts.dest));
  });

  gulp.task('concat', function () {
    return gulp.src(scripts.src)
    .pipe(plugins.sourcemaps.init())
    .pipe(plugins.concat(namespace + '.js'))
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


// Publish
// Command: gulp [task] --publish
// -----------------------------------------------------------------------------

tasks.publish = function () {
  var namespace = 'publish',
      pathto = function (file) {
        return (namespace + '/views/themes/' + namespace + '/assets/' + file);
      },
      scripts = {
        src: pathto('javascripts/app/*.js'),
        dest: pathto('javascripts/')
      },
      styles = {
        src: pathto('stylesheets/scss/*.scss'),
        dest: pathto('stylesheets/'),
        main: pathto('stylesheets/' + namespace + '.css'),
        scss: pathto('stylesheets/scss/**/*.scss')
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
    .pipe(plugins.concat(namespace + '.js'))
    .pipe(plugins.uglify())
    .pipe(gulp.dest(scripts.dest));
  });

  gulp.task('concat', function () {
    return gulp.src(scripts.src)
    .pipe(plugins.sourcemaps.init())
    .pipe(plugins.concat(namespace + '.js'))
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

if (task && typeof tasks[task] === 'function') {
  tasks[task]();
} else {
  tasks.base();
}
