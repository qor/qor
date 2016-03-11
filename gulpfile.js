'use strict';

var gulp = require('gulp');
var plugins = require('gulp-load-plugins')();
var moduleName = (function () {
      var args = process.argv;
      var length = args.length;
      var i = 0;
      var name;

      while (i++ < length) {
        if ((/^--+(\w+)/i).test(args[i])){
          name = args[i].split('--')[1];
          break;
        }
      }
      return name;
    })();

// Admin Module
// Command: gulp [task]
// Admin is default task
// Watch Admin module: gulp
// Release Admin module: gulp release
// -----------------------------------------------------------------------------

function adminTasks() {
  var pathto = function (file) {
        return ('../admin/views/assets/' + file);
      };
  var scripts = {
        src: pathto('javascripts/app/*.js'),
        dest: pathto('javascripts'),
        qor: pathto('javascripts/qor/*.js'),
        qorInit: pathto('javascripts/qor/qor-init.js'),
        all: [
          'gulpfile.js',
          pathto('javascripts/qor/*.js')
        ]
      };
  var styles = {
        src: pathto('stylesheets/scss/{app,qor}.scss'),
        dest: pathto('stylesheets'),
        vendors: pathto('stylesheets/vendors'),
        main: pathto('stylesheets/{qor,app}.css'),
        scss: pathto('stylesheets/scss/**/*.scss')
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
    return gulp.src([scripts.qorInit,scripts.qor])
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
    return gulp.src([scripts.qorInit,scripts.qor])
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

  gulp.task('watch', function () {
    gulp.watch(scripts.qor, ['qor+']);
    gulp.watch(scripts.src, ['js+']);
    gulp.watch(styles.scss, ['sass']);
  });

  gulp.task('release', ['js', 'css']);

  gulp.task('default', ['watch']);
}


// Other Modules
// Command: gulp [task] --moduleName
// Watch Worker module: gulp --worker
// Release Worker module: gulp release --worker
// -----------------------------------------------------------------------------

function moduleTasks(moduleNames) {
  var pathto = function (file) {
    return '../' + moduleNames + '/views/themes/' + moduleNames + '/assets/' + file;
  };

  var scripts = {
        src: pathto('javascripts/app/*.js'),
        dest: pathto('javascripts/')
      };
  var styles = {
        src: pathto('stylesheets/scss/*.scss'),
        dest: pathto('stylesheets/'),
        main: pathto('stylesheets/' + moduleNames + '.css'),
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

  gulp.task('js', ['jshint', 'jscs'], function () {
    return gulp.src(scripts.src)
    .pipe(plugins.concat(moduleNames + '.js'))
    .pipe(plugins.uglify())
    .pipe(gulp.dest(scripts.dest));
  });

  gulp.task('concat', function () {
    return gulp.src(scripts.src)
    .pipe(plugins.sourcemaps.init())
    .pipe(plugins.concat(moduleNames + '.js'))
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
}


// Init
// -----------------------------------------------------------------------------

if (moduleName) {
  var runModuleName = 'Running "' + moduleName + '" module task...';
  console.log(runModuleName);
  moduleTasks(moduleName);
} else {
  console.log('Running "admin" module task...');
  adminTasks();
}

// Task for compress js and css vendor assets
gulp.task('compressJavaScriptVendor', function () {
  return gulp.src(['!../admin/views/assets/javascripts/vendors/jquery.min.js','../admin/views/assets/javascripts/vendors/*.js'])
  .pipe(plugins.concat('vendors.js'))
  .pipe(gulp.dest('../admin/views/assets/javascripts'));
});

gulp.task('compressCSSVendor', function () {
  return gulp.src('../admin/views/assets/stylesheets/vendors/*.css')
  .pipe(plugins.concat('vendors.css'))
  .pipe(gulp.dest('../admin/views/assets/stylesheets'));
});
