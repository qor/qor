
# qor gulp-uglify打包流程
####gulpfile文件
```
 task: 一个task就是一个压缩任务【来源可以单个或者数组】
 src: 压缩文件来源
 plumber: 检查执行的例外情况
 concat: 文件合并，不存在会自动创建
 uglify：压缩
 eslint: es语法检查
 pathto: 指向的路径
 gulp.dest: 压缩之后的文件输出到哪里
 plugins.xxx: 这个是一个加载package.json文件中间插件的工具(gulp-load-plugins)
```

####打包前准备工作
1 . 安装所有的依赖: npm install,根据GitHub上面的源码查看不要修改原有插件的版本号

2 . 进入qor(project),admin(project)这两个项目必须在同一个目录下面，如果出现报错如下
```
Plumber found unhandled error:
 GulpUglifyError: unable to minify JavaScript
Caused by: SyntaxError: Unexpected token: name (remoteDataPrimaryKey)
```
解决方式是在gulpfile.js文件里面的所需要用到的task下面增加es-2015插件，代码如下:

```
在qor项目和admin下面执行: npm install --save-dev babel-preset-es2015
然后在需要用到的task里面增加es2015的兼容
gulp.task('qor', function() {
  return gulp
  .src([scripts.qorInit, scripts.qorCommon, scripts.qor])
  .pipe(
    babel({
      presets: ['es2015']
    })
  )
  .pipe(plumber())
  .pipe(plugins.concat('qor.js'))
  .pipe(plugins.uglify())
  .pipe(gulp.dest(scripts.dest));
});
```


3 . 如果在打包的时候出现env环境不可达的情况，请按照如下操作，在qor和admin里面必须全部添加

```
安装如下插件
npm install --save-dev babel-preset-env
```


####开始打包
1 . ***开始打包的时候一定要主意，因为gulp-uglify打包vendors目录的时候会按照目录顺序取出来进行打包，那么juqery就会排在cropper.min.js的前面被打包，导致所有压缩之后的包在后期使用的时候会出现jquery没有引入的错误，这里解决方式是将jquery.min.js的名称修改为a-jquery.min.js，就会排在第一被打包***

###task解释
1 . 所有的task解释(js)
```
1. gulp qor
解释: 将javascripts/qor/*.js，javascripts/qor/qor-config.js,javascripts/qor/qor-common.js合并压缩为qor.js
2. gulp js
解释: 这个task执行之前必须将任务qor执行完毕，这个是将javascripts/app/*.js压缩成app.js
3. gulp qor+
解释: 和第一步其实是一样的，增加了eslintrc,env环境
4. gulp js+
解释: 同上
5. gulp release_js
解释: 将javascripts/qor.js，javascripts/app.js压缩合并成qor_admin_default.js
6. gulp combineJavaScriptVendor
解释: 打包vendor下面所有的用到的第三方库
第六条注意修改成这个样子(这个task执行之前一定要修改jquery.min.js的名称，使之在目录里面排在第一位)
gulp.task('combineJavaScriptVendor', function() {
  return gulp
    .src(['../admin/views/assets/javascripts/vendors/*.js'])
    .pipe(plugins.concat('vendors.js'))
    .pipe(gulp.dest('../admin/views/assets/javascripts'));
});
```

####打包
1 . 执行task（只打包js）
```
 gulp release
 gulp release_js
 gulp combineJavaScriptVendor(这里一定注意jquery包的顺序问题,按照上面说的必须在第一位)
```









