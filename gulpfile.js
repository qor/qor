'use strict';

let gulp = require('gulp'),
  babel = require('gulp-babel'),
  eslint = require('gulp-eslint'),
  plugins = require('gulp-load-plugins')(),
  plumber = require('gulp-plumber'),
  cleanCSS = require('gulp-clean-css'),
  fs = require('fs'),
  path = require('path'),
  es = require('event-stream'),
  rename = require('gulp-rename');

let moduleName = (function() {
  let args = process.argv,
    length = args.length,
    i = 0,
    name,
    subName,
    useSubName;

  while (i++ < length) {
    if (/^--+(\w+)/i.test(args[i])) {
      name = args[i].split('--')[1];
      subName = args[i].split('--')[2];
      useSubName = args[i].split('--')[3];
      break;
    }
  }
  return {
    name: name,
    subName: subName,
    useSubName: useSubName
  };
})();

// Admin Module
// Command: gulp [task]
// Admin is default task
// Watch Admin module: gulp
// -----------------------------------------------------------------------------

function adminTasks() {
  let pathto = function(file) {
    return '../admin/views/assets/' + file;
  };
  let scripts = {
    src: pathto('javascripts/app/*.js'),
    dest: pathto('javascripts'),
    qor: pathto('javascripts/qor/*.js'),
    qorInit: pathto('javascripts/qor/qor-config.js'),
    qorCommon: pathto('javascripts/qor/qor-common.js'),
    qorAdmin: [pathto('javascripts/qor.js'), pathto('javascripts/app.js')],
    all: ['gulpfile.js', pathto('javascripts/qor/*.js')]
  };
  let styles = {
    src: pathto('stylesheets/scss/{app,qor}.scss'),
    dest: pathto('stylesheets'),
    vendors: pathto('stylesheets/vendors'),
    main: pathto('stylesheets/{qor,app}.css'),
    qorAdmin: [pathto('stylesheets/vendors.css'), pathto('stylesheets/qor.css'), pathto('stylesheets/app.css')],
    scss: pathto('stylesheets/scss/**/*.scss')
  };

  gulp.task('qor', function() {
    return gulp
      .src([scripts.qorInit, scripts.qorCommon, scripts.qor])
      .pipe(plumber())
      .pipe(plugins.concat('qor.js'))
      .pipe(plugins.uglify())
      .pipe(gulp.dest(scripts.dest));
  });

  gulp.task('js', ['qor'], function() {
    return gulp
      .src(scripts.src)
      .pipe(plumber())
      .pipe(
        eslint({
          configFile: '.eslintrc'
        })
      )
      .pipe(plugins.concat('app.js'))
      .pipe(plugins.uglify())
      .pipe(gulp.dest(scripts.dest));
  });

  gulp.task('qor+', function() {
    return gulp
      .src([scripts.qorInit, scripts.qorCommon, scripts.qor])
      .pipe(plumber())
      .pipe(
        eslint({
          configFile: '.eslintrc'
        })
      )
      .pipe(
        babel({
          presets: ['env']
        })
      )
      .pipe(eslint.format())
      .pipe(plugins.concat('qor.js'))
      .pipe(plugins.uglify())
      .pipe(gulp.dest(scripts.dest));
  });

  gulp.task('js+', function() {
    return gulp
      .src(scripts.src)
      .pipe(plumber())
      .pipe(
        babel({
          presets: ['env']
        })
      )
      .pipe(eslint.format())
      .pipe(plugins.concat('app.js'))
      .pipe(plugins.uglify())
      .pipe(gulp.dest(scripts.dest));
  });

  gulp.task('sass', function() {
    return gulp
      .src(styles.src)
      .pipe(plumber())
      .pipe(plugins.sass().on('error', plugins.sass.logError))
      .pipe(gulp.dest(styles.dest));
  });

  gulp.task('css', ['sass'], function() {
    return gulp
      .src(styles.main)
      .pipe(plumber())
      .pipe(plugins.autoprefixer())
      .pipe(plugins.csscomb())
      .pipe(cleanCSS())
      .pipe(gulp.dest(styles.dest));
  });

  gulp.task('release_js', function() {
    return gulp
      .src(scripts.qorAdmin)
      .pipe(plugins.concat('qor_admin_default.js'))
      .pipe(gulp.dest(scripts.dest));
  });

  gulp.task('release_css', function() {
    return gulp
      .src(styles.qorAdmin)
      .pipe(plugins.concat('qor_admin_default.css'))
      .pipe(gulp.dest(styles.dest));
  });

  gulp.task('release', ['qor+', 'js+', 'css', 'release_js', 'release_css']);

  let watcher = gulp.task('watch', function() {
    let watch_qor = gulp.watch(scripts.qor, ['qor+']),
      watch_js = gulp.watch(scripts.src, ['js+']),
      watch_css = gulp.watch(styles.scss, ['css']);

    gulp.watch(styles.qorAdmin, ['release_css']);
    gulp.watch(scripts.qorAdmin, ['release_js']);

    watch_qor.on('change', function(event) {
      console.log(':==> File ' + event.path + ' was ' + event.type + ', running tasks...');
    });
    watch_js.on('change', function(event) {
      console.log(':==> File ' + event.path + ' was ' + event.type + ', running tasks...');
    });
    watch_css.on('change', function(event) {
      console.log(':==> File ' + event.path + ' was ' + event.type + ', running tasks...');
    });
  });

  gulp.task('default', ['watch']);
}

// -----------------------------------------------------------------------------
// Other Modules
// Command: gulp [task] --moduleName--subModuleName
//
//  example:
// Watch Worker module: gulp --worker
//
// if module's assets just as normal path:
// moduleName/views/themes/moduleName/assets/javascripts(stylesheets)
// just use gulp --worker
//
// if module's assets in enterprise as normal path:
// moduleName/views/themes/moduleName/assets/javascripts(stylesheets)
// just use gulp --microsite--enterprise
//
// if module's assets path as Admin module:
// moduleName/views/assets/javascripts(stylesheets)
// you need set subModuleName as admin
// gulp --worker--admin
//
// if you need run task for subModule in modules
// example: worker module inline_edit subModule:
// gulp --worker--inline_edit
//
// gulp --media--media_library--true
//
// -----------------------------------------------------------------------------

function moduleTasks(moduleNames) {
  let moduleName = moduleNames.name,
    subModuleName = moduleNames.subName,
    useSubName = moduleNames.useSubName;

  let pathto = function(file) {
    if (moduleName && subModuleName) {
      if (subModuleName == 'admin') {
        return '../' + moduleName + '/views/assets/' + file;
      } else if (subModuleName == 'enterprise') {
        return '../../../enterprise.getqor.com/' + moduleName + '/views/themes/' + moduleName + '/assets/' + file;
      } else if (useSubName) {
        if (useSubName == 'admin') {
          return '../' + moduleName + '/' + subModuleName + '/views/assets/' + file;
        } else {
          return '../' + moduleName + '/' + subModuleName + '/views/themes/' + subModuleName + '/assets/' + file;
        }
      } else {
        return '../' + moduleName + '/' + subModuleName + '/views/themes/' + moduleName + '/assets/' + file;
      }
    }
    return '../' + moduleName + '/views/themes/' + moduleName + '/assets/' + file;
  };

  let scripts = {
    src: pathto('javascripts/'),
    watch: pathto('javascripts/**/*.js')
  };
  let styles = {
    src: pathto('stylesheets/'),
    watch: pathto('stylesheets/**/*.scss')
  };

  function getFolders(dir) {
    return fs.readdirSync(dir).filter(function(file) {
      return fs.statSync(path.join(dir, file)).isDirectory();
    });
  }

  gulp.task('js', function() {
    let scriptPath = scripts.src,
      folders = getFolders(scriptPath);

    let task = folders.map(function(folder) {
      return gulp
        .src(path.join(scriptPath, folder, '/*.js'))
        .pipe(plumber())
        .pipe(
          eslint({
            configFile: '.eslintrc'
          })
        )
        .pipe(
          babel({
            presets: ['env']
          })
        )
        .pipe(eslint.format())
        .pipe(plugins.concat(folder + '.js'))
        .pipe(plugins.uglify())
        .pipe(gulp.dest(scriptPath));
    });

    return es.concat.apply(null, task);
  });

  gulp.task('css', function() {
    let stylePath = styles.src,
      folders = getFolders(stylePath);

    let task = folders.map(function(folder) {
      return gulp
        .src(path.join(stylePath, folder, '/*.scss'))
        .pipe(plumber())
        .pipe(
          plugins
            .sass({
              outputStyle: 'compressed'
            })
            .on('error', plugins.sass.logError)
        )
        .pipe(cleanCSS())
        .pipe(rename(folder + '.css'))
        .pipe(gulp.dest(stylePath));
    });

    return es.concat.apply(null, task);
  });

  gulp.task('watch', function() {
    let moduleScript = gulp.watch(scripts.watch, {debounceDelay: 2000}, ['js']);
    gulp.watch(styles.watch, ['css']);

    moduleScript.on('change', function(event) {
      console.log(':==> File ' + event.path + ' was ' + event.type + ', running tasks...');
    });
  });

  gulp.task('default', ['watch']);
  gulp.task('release', ['js', 'css']);
}

// Init
// -----------------------------------------------------------------------------

if (moduleName.name) {
  let taskPath = moduleName.name + '/views/themes/' + moduleName.name + '/assets/',
    runModuleName = 'Running "' + moduleName.name + '" module task in "' + taskPath + '"...';

  if (moduleName.subName) {
    if (moduleName.subName == 'admin') {
      taskPath = moduleName.name + '/views/assets/';
      runModuleName = 'Running "' + moduleName.name + '" module task in "' + taskPath + '"...';
    } else if (moduleName.subName == 'enterprise') {
      taskPath = '../../../enterprise.getqor.com/' + moduleName.name + '/views/themes/' + moduleName.name + '/assets/';
      runModuleName = 'Running "' + moduleName.name + '" module task in "' + taskPath + '"...';
    } else if (moduleName.useSubName) {
      if (moduleName.useSubName == 'admin') {
        taskPath = moduleName.name + '/' + moduleName.subName + '/views/assets/';
      } else {
        taskPath = moduleName.name + '/' + moduleName.subName + '/views/themes/' + moduleName.subName + '/assets/';
      }

      runModuleName = 'Running "' + moduleName.name + ' > ' + moduleName.subName + '" module task in "' + taskPath + '"...';
    } else {
      taskPath = moduleName.name + '/' + moduleName.subName + '/views/themes/' + moduleName.name + '/assets/';
      runModuleName = 'Running "' + moduleName.name + ' > ' + moduleName.subName + '" module task in "' + taskPath + '"...';
    }
  }
  console.log(runModuleName);
  moduleTasks(moduleName);
} else {
  console.log('Running "admin" module task in "admin/views/assets/"...');
  adminTasks();
}

// Task for compress js and css vendor assets
gulp.task('combineJavaScriptVendor', function() {
  return gulp
    .src(['!../admin/views/assets/javascripts/vendors/jquery.min.js', '../admin/views/assets/javascripts/vendors/*.js'])
    .pipe(plugins.concat('vendors.js'))
    .pipe(gulp.dest('../admin/views/assets/javascripts'));
});

gulp.task('compressCSSVendor', function() {
  return gulp
    .src('../admin/views/assets/stylesheets/vendors/*.css')
    .pipe(plugins.concat('vendors.css'))
    .pipe(gulp.dest('../admin/views/assets/stylesheets'));
});
