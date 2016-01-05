'use strict';

var gulp = require('gulp');
var plugins = require('gulp-load-plugins')();
var moduleName = (function () {
      var args = process.argv;
      var length = args.length;
      var i = 0;
      var name;
      var isExternal;

      while (i++ < length) {
        if ((/^--+(\w+)/i).test(args[i])){
          name = args[i].split('--');
          if (name[2]){
            isExternal = name[2];
          }
          name = name[1];
          break;
        }
      }

      return { name,isExternal };

    })();

// Task for compress js and css plugin assets
gulp.task('compress_js_plugin', function () {
  return gulp.src(['!admin/views/assets/javascripts/vendors/jquery.min.js','admin/views/assets/javascripts/vendors/*.js'])
  .pipe(plugins.concat('vendors.js'))
  .pipe(gulp.dest('admin/views/assets/javascripts'));
});

gulp.task('compress_css_plugin', function () {
  return gulp.src('admin/views/assets/stylesheets/vendors/*.css')
  .pipe(plugins.concat('vendors.css'))
  .pipe(gulp.dest('admin/views/assets/stylesheets'));
});

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
    .pipe(gulp.dest(scripts.vendors));
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

  gulp.task('mdl', function () {
    return gulp.src([
      'bower_components/material-design-lite/src/_*',
    ])
    .pipe(gulp.dest(pathto('stylesheets/scss/mdl')));
  });

  gulp.task('fonts', function () {
    return gulp.src([
      'bower_components/material-design-icons/iconfont/codepoints',
      'bower_components/material-design-icons/iconfont/MaterialIcons*'
    ])
    .pipe(gulp.dest(fonts.dest));
  });

  gulp.task('csslib', ['mdl', 'fonts'], function () {
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
}


// Modules
// Command: gulp [task] --moduleName
// If Modules is external use --moduleName--external
// -----------------------------------------------------------------------------

function moduleTasks(moduleNames) {
  var pathto = function (file) {
    var moduleName = moduleNames.name;

    if (moduleNames.isExternal){
      moduleName =  '../' + moduleNames.name;
    }
    return (moduleName + '/views/themes/' + moduleNames.name + '/assets/' + file);
  };

  var scripts = {
        src: pathto('javascripts/app/*.js'),
        dest: pathto('javascripts/')
      };
  var styles = {
        src: pathto('stylesheets/scss/*.scss'),
        dest: pathto('stylesheets/'),
        main: pathto('stylesheets/' + moduleNames.name + '.css'),
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
    .pipe(plugins.concat(moduleNames.name + '.js'))
    .pipe(plugins.uglify())
    .pipe(gulp.dest(scripts.dest));
  });

  gulp.task('concat', function () {
    return gulp.src(scripts.src)
    .pipe(plugins.sourcemaps.init())
    .pipe(plugins.concat(moduleNames.name + '.js'))
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

if (moduleName.name) {
  var runModuleName = 'Running internal"' + moduleName.name + '" task...';
  if (moduleName.isExternal){
    runModuleName = 'Running external "' + moduleName.name + '" task...';
  }
  console.log(runModuleName);
  moduleTasks(moduleName);
} else {
  console.log('Running "admin" task...');
  adminTasks();
}
