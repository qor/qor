'use strict';

var gulp = require('gulp');
var eslint = require('gulp-eslint');
var plugins = require('gulp-load-plugins')();

var fs = require('fs');
var path = require('path');
var es = require('event-stream');
var rename = require('gulp-rename');

var moduleName = (function () {
      var args = process.argv;
      var length = args.length;
      var i = 0;
      var name;
      var subName;

      while (i++ < length) {
        if ((/^--+(\w+)/i).test(args[i])){
          name = args[i].split('--')[1];
          subName = args[i].split('--')[2];
          break;
        }
      }
      return {
        'name': name,
        'subName': subName
      };
    })();

// Admin Module
// Command: gulp [task]
// Admin is default task
// Watch Admin module: gulp
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
// Command: gulp [task] --moduleName--subModuleName
// Watch Worker module: gulp --worker
// Watch Worker inline_edit subModule: gulp --worker--inline_edit
// -----------------------------------------------------------------------------

function moduleTasks(moduleNames) {
  var moduleName = moduleNames.name;
  var subModuleName = moduleNames.subName;

  var pathto = function (file) {
    if(moduleName && subModuleName) {
      return '../' + moduleName + '/' + subModuleName + '/views/themes/' + moduleName + '/assets/' + file;
    }
    return '../' + moduleName + '/views/themes/' + moduleName + '/assets/' + file;
  };

  var scripts = {
        src: pathto('javascripts/'),
        watch: pathto('javascripts/**/*.js')
      };
  var styles = {
        src: pathto('stylesheets/'),
        watch: pathto('stylesheets/**/*.scss')
      };

  function getFolders(dir){
    return fs.readdirSync(dir).filter(function(file){
      return fs.statSync(path.join(dir, file)).isDirectory();
    })
  }


  gulp.task('concat', function () {
    var scriptPath = scripts.src;
    var folders = getFolders(scriptPath);
    var task = folders.map(function(folder){

      return gulp.src(path.join(scriptPath, folder, '/*.js'))
        .pipe(eslint({configFile: '.eslintrc'}))
        .pipe(eslint.format())
        .pipe(plugins.sourcemaps.init())
        .pipe(plugins.concat(folder + '.js'))
        .pipe(plugins.uglify())
        .pipe(plugins.sourcemaps.write('./'))
        .pipe(gulp.dest(scriptPath));
    });

    return es.concat.apply(null, task);

  });

  gulp.task('css', function () {

    var stylePath = styles.src;
    var folders = getFolders(stylePath);
    var task = folders.map(function(folder){

      return gulp.src(path.join(stylePath, folder, '/*.scss'))

        .pipe(plugins.sourcemaps.init())
        .pipe(plugins.sass({outputStyle: 'compressed'}))
        .pipe(plugins.sourcemaps.write('./'))
        .pipe(plugins.minifyCss())
        .pipe(rename(folder + '.css'))
        .pipe(gulp.dest(stylePath))
    });

    return es.concat.apply(null, task);

  });

  gulp.task('watch', function () {
    gulp.watch(scripts.watch, ['concat']);
    gulp.watch(styles.watch, ['css']);
  });

  gulp.task('default', ['watch']);
}


// Init
// -----------------------------------------------------------------------------

if (moduleName.name) {
  var runModuleName = 'Running "' + moduleName.name + '" module task...';

  if (moduleName.subName){
    runModuleName = 'Running "' + moduleName.name + ' > ' + moduleName.subName + '" module task...';
  }

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
