(function (factory) {
  if (typeof define === 'function' && define.amd) {
    // AMD. Register as anonymous module.
    define(['jquery'], factory);
  } else if (typeof exports === 'object') {
    // Node / CommonJS
    factory(require('jquery'));
  } else {
    // Browser globals.
    factory(jQuery);
  }
})(function ($) {

  'use strict';

  var URL = window.URL || window.webkitURL;
  var NAMESPACE = 'qor.cropper';

  // Events
  var EVENT_ENABLE = 'enable.' + NAMESPACE;
  var EVENT_DISABLE = 'disable.' + NAMESPACE;
  var EVENT_CHANGE = 'change.' + NAMESPACE;
  var EVENT_CLICK = 'click.' + NAMESPACE;
  var EVENT_SHOWN = 'shown.qor.modal';
  var EVENT_HIDDEN = 'hidden.qor.modal';

  // Classes
  var CLASS_TOGGLE = '.qor-cropper__toggle';
  var CLASS_CANVAS = '.qor-cropper__canvas';
  var CLASS_WRAPPER = '.qor-cropper__wrapper';
  var CLASS_OPTIONS = '.qor-cropper__options';
  var CLASS_SAVE = '.qor-cropper__save';

  function capitalize(str) {
    if (typeof str === 'string') {
      str = str.charAt(0).toUpperCase() + str.substr(1);
    }

    return str;
  }

  function getLowerCaseKeyObject(obj) {
    var newObj = {};
    var key;

    if ($.isPlainObject(obj)) {
      for (key in obj) {
        if (obj.hasOwnProperty(key)) {
          newObj[String(key).toLowerCase()] = obj[key];
        }
      }
    }

    return newObj;
  }

  function getValueByNoCaseKey(obj, key) {
    var originalKey = String(key);
    var lowerCaseKey = originalKey.toLowerCase();
    var upperCaseKey = originalKey.toUpperCase();
    var capitalizeKey = capitalize(originalKey);

    if ($.isPlainObject(obj)) {
      return (obj[lowerCaseKey] || obj[capitalizeKey] || obj[upperCaseKey]);
    }
  }

  function replaceText(str, data) {
    if (typeof str === 'string') {
      if (typeof data === 'object') {
        $.each(data, function (key, val) {
          str = str.replace('${' + String(key).toLowerCase() + '}', val);
        });
      }
    }

    return str;
  }

  function QorCropper(element, options) {
    this.$element = $(element);
    this.options = $.extend(true, {}, QorCropper.DEFAULTS, $.isPlainObject(options) && options);
    this.data = null;
    this.init();
  }

  QorCropper.prototype = {
    constructor: QorCropper,

    init: function () {
      var options = this.options;
      var $this = this.$element;
      var $parent = $this.closest(options.parent);
      var $list;
      var data;

      if (!$parent.length) {
        $parent = $this.parent();
      }

      this.$parent = $parent;
      this.$output = $parent.find(options.output);
      this.$list = $list = $parent.find(options.list);

      if (!$list.find('img').attr('src')) {
        $list.find('ul').hide();
      }

      try {
        data = JSON.parse($.trim(this.$output.val()));
      } catch (e) {}

      this.data = data || {};
      this.build();
      this.bind();
    },

    build: function () {
      this.wrap();
      this.$modal = $(replaceText(QorCropper.MODAL, this.options.text)).appendTo('body');
    },

    unbuild: function () {
      this.$modal.remove();
      this.unwrap();
    },

    wrap: function () {
      var $list = this.$list;
      var $img;

      $list.find('li').append(QorCropper.TOGGLE);
      $img = $list.find('img');
      $img.wrap(QorCropper.CANVAS);
      this.center($img);
    },

    unwrap: function () {
      var $list = this.$list;

      $list.find(CLASS_TOGGLE).remove();
      $list.find(CLASS_CANVAS).each(function () {
        var $this = $(this);

        $this.before($this.html()).remove();
      });
    },

    bind: function () {
      this.$element.
        on(EVENT_CHANGE, $.proxy(this.read, this));

      this.$list.
        on(EVENT_CLICK, $.proxy(this.click, this));

      this.$modal.
        on(EVENT_SHOWN, $.proxy(this.start, this)).
        on(EVENT_HIDDEN, $.proxy(this.stop, this));
    },

    unbind: function () {
      this.$element.
        off(EVENT_CHANGE, this.read);

      this.$list.
        off(EVENT_CLICK, this.click);

      this.$modal.
        off(EVENT_SHOWN, this.start).
        off(EVENT_HIDDEN, this.stop);
    },

    click: function (e) {
      var target = e.target;
      var $target;

      if (target === this.$list[0]) {
        return;
      }

      $target = $(target);

      if (!$target.is('img')) {
        $target = $target.closest('li').find('img');
      }

      this.$target = $target;
      this.$modal.qorModal('show');
    },

    read: function (e) {
      var files = e.target.files;
      var file;

      if (files && files.length) {
        file = files[0];

        if (/^image\/\w+$/.test(file.type) && URL) {
          this.load(URL.createObjectURL(file));
        } else {
          this.$list.empty().text(file.name);
        }
      }
    },

    load: function (url) {
      var options = this.options;
      var _this = this;
      var $list = this.$list;
      var $ul = $list.find('ul');
      var data = this.data;
      var $image;

      if (!$ul.length) {
        $ul  = $(QorCropper.LIST);
        $list.html($ul);
        this.wrap();
      }

      $ul.show(); // show ul when it is hidden

      $image = $list.find('img');
      $image.one('load', function () {
        var $this = $(this);
        var naturalWidth = this.naturalWidth;
        var naturalHeight = this.naturalHeight;
        var sizeData = $this.data();
        var sizeResolution = sizeData.sizeResolution;
        var sizeName = sizeData.sizeName;
        var emulateImageData = {};
        var emulateCropData = {};
        var aspectRatio;
        var width;
        var height;

        if (sizeResolution) {
          width = getValueByNoCaseKey(sizeResolution, 'width');
          height = getValueByNoCaseKey(sizeResolution, 'height');
          aspectRatio = width / height;

          if (naturalHeight * aspectRatio > naturalWidth) {
            width = naturalWidth;
            height = width / aspectRatio;
          } else {
            height = naturalHeight;
            width = height * aspectRatio;
          }

          width *= 0.8;
          height *= 0.8;

          emulateImageData = {
            naturalWidth: naturalWidth,
            naturalHeight: naturalHeight,
          };

          emulateCropData = {
            x: Math.round((naturalWidth - width) / 2),
            y: Math.round((naturalHeight - height) / 2),
            width: Math.round(width),
            height: Math.round(height),
          };

          _this.preview($this, emulateImageData, emulateCropData);

          if (sizeName) {
            data.crop = true;

            if (!data[options.key]) {
              data[options.key] = {};
            }

            data[options.key][sizeName] = emulateCropData;
          }
        } else {
          _this.center($this);
        }

        _this.$output.val(JSON.stringify(data));
      }).attr('src', url).data('originalUrl', url);
    },

    start: function () {
      var options = this.options;
      var $modal = this.$modal;
      var $target = this.$target;
      var sizeData = $target.data();
      var sizeName = sizeData.sizeName || 'original';
      var sizeResolution = sizeData.sizeResolution;
      var $clone = $('<img>').attr('src', sizeData.originalUrl);
      var data = this.data;
      var _this = this;
      var sizeAspectRatio = NaN;
      var sizeWidth;
      var sizeHeight;
      var list;

      if (sizeResolution) {
        sizeWidth = getValueByNoCaseKey(sizeResolution, 'width');
        sizeHeight = getValueByNoCaseKey(sizeResolution, 'height');
        sizeAspectRatio = sizeWidth / sizeHeight;
      }

      if (!data[options.key]) {
        data[options.key] = {};
      }

      $modal.trigger('enable.qor.material').find(CLASS_WRAPPER).html($clone);

      list = this.getList(sizeAspectRatio);

      if (list) {
        $modal.find(CLASS_OPTIONS).show().append(list);
      }

      $clone.cropper({
        aspectRatio: sizeAspectRatio,
        data: getLowerCaseKeyObject(data[options.key][sizeName]),
        background: false,
        movable: false,
        zoomable: false,
        scalable: false,
        rotatable: false,
        checkImageOrigin: false,

        built: function () {
          $modal.find(CLASS_SAVE).one(EVENT_CLICK, function () {
            var cropData = $clone.cropper('getData', true);
            var syncData = [];
            var url;

            data.crop = true;
            data[options.key][sizeName] = cropData;
            _this.imageData = $clone.cropper('getImageData');
            _this.cropData = cropData;

            try {
              url = $clone.cropper('getCroppedCanvas').toDataURL();
            } catch (e) {}

            $modal.find(CLASS_OPTIONS + ' input').each(function () {
              var $this = $(this);

              if ($this.prop('checked')) {
                syncData.push($this.attr('name'));
              }
            });

            _this.output(url, syncData);
            $modal.qorModal('hide');
          });
        },
      });
    },

    stop: function () {
      this.$modal.
        trigger('disable.qor.material').
        find(CLASS_WRAPPER + ' > img').
          cropper('destroy').
          remove().
          end().
        find(CLASS_OPTIONS).
          hide().
          find('ul').
            remove();
    },

    getList: function (aspectRatio) {
      var list = [];

      this.$list.find('img').not(this.$target).each(function () {
        var data = $(this).data();
        var resolution = data.sizeResolution;
        var name = data.sizeName;
        var width;
        var height;

        if (resolution) {
          width = getValueByNoCaseKey(resolution, 'width');
          height = getValueByNoCaseKey(resolution, 'height');

          if (width / height === aspectRatio) {
            list.push(
              '<label>' +
                '<input type="checkbox" name="' + name + '" checked> ' +
                '<span>' + name +
                  '<small>(' + width + '&times;' + height + ' px)</small>' +
                '</span>' +
              '</label>'
            );
          }
        }
      });

      return list.length ? ('<ul><li>' + list.join('</li><li>') + '</li></ul>') : '';
    },

    output: function (url, data) {
      var $target = this.$target;

      if (url) {
        this.center($target.attr('src', url), true);
      } else {
        this.preview($target);
      }

      if ($.isArray(data) && data.length) {
        this.autoCrop(url, data);
      }

      this.$output.val(JSON.stringify(this.data));
    },

    preview: function ($target, emulateImageData, emulateCropData) {
      var $canvas = $target.parent();
      var $container = $canvas.parent();
      var containerWidth = $container.width();
      var containerHeight = $container.height();
      var imageData = emulateImageData || this.imageData;
      var cropData = $.extend({}, emulateCropData || this.cropData); // Clone one to avoid changing it
      var aspectRatio = cropData.width / cropData.height;
      var canvasWidth = containerWidth;
      var canvasHeight = containerHeight;
      var scaledRatio;

      if (containerHeight * aspectRatio > containerWidth) {
        canvasHeight = containerWidth / aspectRatio;
      } else {
        canvasWidth = containerHeight * aspectRatio;
      }

      scaledRatio = cropData.width / canvasWidth;

      $canvas.css({
        width: canvasWidth,
        height: canvasHeight,
      });

      $target.css({
        maxWidth: 'none',
        maxHeight: 'none',
        width: imageData.naturalWidth / scaledRatio,
        height: imageData.naturalHeight / scaledRatio,
        marginLeft: -cropData.x / scaledRatio,
        marginTop: -cropData.y / scaledRatio,
      });

      this.center($target);
    },

    center: function ($target, reset) {
      $target.each(function () {
        var $this = $(this);
        var $canvas = $this.parent();
        var $container = $canvas.parent();

        function center() {
          var containerHeight = $container.height();
          var canvasHeight = $canvas.height();
          var marginTop = 'auto';

          if (canvasHeight < containerHeight) {
            marginTop = (containerHeight - canvasHeight) / 2;
          }

          $canvas.css('margin-top', marginTop);
        }

        if (reset) {
          $canvas.add($this).removeAttr('style');
        }

        if (this.complete) {
          center.call(this);
        } else {
          this.onload = center;
        }
      });
    },

    autoCrop: function (url, data) {
      var cropData = this.cropData;
      var cropOptions = this.data[this.options.key];
      var _this = this;

      this.$list.find('img').not(this.$target).each(function () {
        var $this = $(this);
        var sizeName = $this.data('sizeName');

        if ($.inArray(sizeName, data) > -1) {
          cropOptions[sizeName] = $.extend({}, cropData);

          if (url) {
            _this.center($this.attr('src', url), true);
          } else {
            _this.preview($this);
          }
        }
      });
    },

    destroy: function () {
      this.unbind();
      this.unbuild();
      this.$element.removeData(NAMESPACE);
    },
  };

  QorCropper.DEFAULTS = {
    parent: false,
    output: false,
    list: false,
    key: 'data',
    data: null,
    text: {
      title: 'Crop the image',
      ok: 'OK',
      cancel: 'Cancel',
    },
  };

  QorCropper.TOGGLE = '<div class="qor-cropper__toggle"><i class="material-icons">crop</i></div>';
  QorCropper.CANVAS = '<div class="qor-cropper__canvas"></div>';
  QorCropper.LIST = '<ul><li><img></li></ul>';
  QorCropper.MODAL = (
    '<div class="qor-modal fade" tabindex="-1" role="dialog" aria-hidden="true">' +
      '<div class="mdl-card mdl-shadow--2dp" role="document">' +
        '<div class="mdl-card__title">' +
          '<h2 class="mdl-card__title-text">${title}</h2>' +
        '</div>' +
        '<div class="mdl-card__supporting-text">' +
          '<div class="qor-cropper__wrapper"></div>' +
          '<div class="qor-cropper__options">' +
            '<p>Sync cropping result to:</p>' +
          '</div>' +
        '</div>' +
        '<div class="mdl-card__actions mdl-card--border">' +
          '<a class="mdl-button mdl-button--colored mdl-js-button mdl-js-ripple-effect qor-cropper__save">${ok}</a>' +
          '<a class="mdl-button mdl-button--colored mdl-js-button mdl-js-ripple-effect" data-dismiss="modal">${cancel}</a>' +
        '</div>' +
        '<div class="mdl-card__menu">' +
          '<button class="mdl-button mdl-button--icon mdl-js-button mdl-js-ripple-effect" data-dismiss="modal" aria-label="close">' +
            '<i class="material-icons">close</i>' +
          '</button>' +
        '</div>' +
      '</div>' +
    '</div>'
  );

  QorCropper.plugin = function (option) {
    return this.each(function () {
      var $this = $(this);
      var data = $this.data(NAMESPACE);
      var options;
      var fn;

      if (!data) {
        if (!$.fn.cropper) {
          return;
        }

        if (/destroy/.test(option)) {
          return;
        }

        options = $.extend(true, {}, $this.data(), typeof option === 'object' && option);
        $this.data(NAMESPACE, (data = new QorCropper(this, options)));
      }

      if (typeof option === 'string' && $.isFunction(fn = data[option])) {
        fn.apply(data);
      }
    });
  };

  $(function () {
    var selector = '.qor-file__input';
    var options = {
          parent: '.qor-file',
          output: '.qor-file__options',
          list: '.qor-file__list',
          key: 'CropOptions',
        };

    $(document).
      on(EVENT_ENABLE, function (e) {
        QorCropper.plugin.call($(selector, e.target), options);
      }).
      on(EVENT_DISABLE, function (e) {
        QorCropper.plugin.call($(selector, e.target), 'destroy');
      }).
      triggerHandler(EVENT_ENABLE);
  });

  return QorCropper;

});
