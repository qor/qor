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

  var URL = window.URL || window.webkitURL,
      NAMESPACE = 'qor.cropper',
      EVENT_CHANGE = 'change.' + NAMESPACE,
      EVENT_CLICK = 'click.' + NAMESPACE,
      EVENT_SHOWN = 'shown.bs.modal',
      EVENT_HIDDEN = 'hidden.bs.modal',
      REGEXP_OPTIONS = /x|y|width|height/,

      QorCropper = function (element, options) {
        this.$element = $(element);
        this.options = $.extend(true, {}, QorCropper.DEFAULTS, options);
        this.data = null;
        this.init();
      };

  function capitalize (str) {
    if (typeof str === 'string') {
      str = str.charAt(0).toUpperCase() + str.substr(1);
    }

    return str;
  }

  function getLowerCaseKeyObject (obj) {
    var newObj = {},
        key;

    if ($.isPlainObject(obj)) {
      for (key in obj) {
        if (obj.hasOwnProperty(key)) {
          newObj[String(key).toLowerCase()] = obj[key];
        }
      }
    }

    return newObj;
  }

  function getCapitalizeKeyObject (obj) {
    var newObj = {},
        key;

    if ($.isPlainObject(obj)) {
      for (key in obj) {
        if (obj.hasOwnProperty(key)) {
          newObj[capitalize(key)] = obj[key];
        }
      }
    }

    return newObj;
  }

  function getValueByNoCaseKey (obj, key) {
    var originalKey = String(key),
        lowerCaseKey = originalKey.toLowerCase(),
        upperCaseKey = originalKey.toUpperCase(),
        capitalizeKey = capitalize(originalKey);

    if ($.isPlainObject(obj)) {
      return (obj[lowerCaseKey] || obj[capitalizeKey] || obj[upperCaseKey]);
    }
  }

  QorCropper.prototype = {
    constructor: QorCropper,

    init: function () {
      var $this = this.$element,
          options = this.options,
          $parent = $this.closest(options.parent),
          $output,
          data;

      if (!$parent.length) {
        $parent = $this.parent();
      }

      this.$parent = $parent;
      this.$output = $output = $parent.find(options.output);
      this.$list = $parent.find(options.list);
      this.$modal = $parent.find(options.modal);

      try {
        data = JSON.parse($.trim($output.val()));
      } catch (e) {}

      this.data = $.extend(data || {}, options.data);
      this.build();
      this.bind();
    },

    build: function () {
      var $list = this.$list,
          $img;

      $list.find('li').append(QorCropper.TOGGLE);
      $img = $list.find('img');
      $img.wrap(QorCropper.CANVAS);
      this.center($img);
    },

    bind: function () {
      this.$element.on(EVENT_CHANGE, $.proxy(this.read, this));
      this.$list.on(EVENT_CLICK, $.proxy(this.click, this));
      this.$modal.on(EVENT_SHOWN, $.proxy(this.start, this)).on(EVENT_HIDDEN, $.proxy(this.stop, this));
    },

    unbind: function () {
      this.$element.off(EVENT_CHANGE, this.read);
      this.$list.off(EVENT_CLICK, this.click);
      this.$modal.off(EVENT_SHOWN, this.start).on(EVENT_HIDDEN, this.stop);
    },

    click: function (e) {
      var target = e.target,
          $target;

      if (e.target === this.$list[0]) {
        return;
      }

      $target = $(target);

      if (!$target.is('img')) {
        $target = $target.closest('li').find('img');
      }

      this.$target = $target;
      this.$modal.modal('show');
    },

    read: function () {
      var files = this.$element.prop('files'),
          file,
          url;

      if (files && files.length) {
        file = files[0];

        this.data[this.options.key] = {};
        this.$output.val(JSON.stringify(this.data));

        if (/^image\/\w+$/.test(file.type) && URL) {
          this.load(URL.createObjectURL(file));
        } else {
          this.$list.empty().text(file.name);
        }
      }
    },

    load: function (url) {
      var $list = this.$list,
          $img;

      if (!$list.find('ul').length) {
        $list.html(QorCropper.TEMPLATE);
        this.build();
      }

      $img = $list.find('img');
      $img.attr('src' , url).data('originalUrl', url);
      this.center($img);
    },

    start: function () {
      var options = this.options,
          $modal = this.$modal,
          $target = this.$target,
          targetData = $target.data(),
          sizeName = targetData.sizeName || 'original',
          sizeResolution = targetData.sizeResolution,
          $clone = $('<img>').attr('src', targetData.originalUrl),
          aspectRatio = sizeResolution ? getValueByNoCaseKey(sizeResolution, 'width') / getValueByNoCaseKey(sizeResolution, 'height') : NaN,
          data = this.data,
          _this = this;

      if (!data[options.key]) {
        data[options.key] = {};
      }

      $modal.find('.modal-body').html($clone);
      $clone.cropper({
        aspectRatio: aspectRatio,
        data: getLowerCaseKeyObject(data[options.key][sizeName]),
        background: false,
        zoomable: false,
        rotatable: false,
        checkImageOrigin: false,

        built: function () {
          $modal.find(options.save).one('click', function () {
            var cropData = {},
                url;

            $.each($clone.cropper('getData'), function (i, n) {
              if (REGEXP_OPTIONS.test(i)) {
                cropData[i] = Math.round(n);
              }
            });

            data[options.key][sizeName] = getCapitalizeKeyObject(cropData);
            _this.imageData = $clone.cropper('getImageData');
            _this.cropData = cropData;

            try {
              url = $clone.cropper('getCroppedCanvas').toDataURL();
            } catch (e) {}

            _this.output(url);
            $modal.modal('hide');
          });
        }
      });
    },

    stop: function () {
      this.$modal.find('.modal-body > img').cropper('destroy').remove();
    },

    output: function (url) {
      if (url) {
        this.center(this.$target.attr('src', url));
      } else {
        this.preview();
      }

      this.$output.val(JSON.stringify(this.data));
    },

    preview: function () {
      var $target = this.$target,
          $canvas = $target.parent(),
          $container = $canvas.parent(),
          containerWidth = Math.max($container.width(), 160), // minContainerWidth: 160
          containerHeight = Math.max($container.height(), 160), // minContainerHeight: 160
          imageData = this.imageData,
          cropData = this.cropData,
          newAspectRatio = cropData.width / cropData.height,
          newWidth = containerWidth,
          newHeight = containerHeight,
          newRatio;

      if (containerHeight * newAspectRatio > containerWidth) {
        newHeight = newWidth / newAspectRatio;
      } else {
        newWidth = newHeight * newAspectRatio;
      }

      newRatio = cropData.width / newWidth;

      $.each(cropData, function (i, n) {
        cropData[i] = n / newRatio;
      });

      $canvas.css({
        width: cropData.width,
        height: cropData.height
      });

      $target.css({
        width: imageData.naturalWidth / newRatio,
        height: imageData.naturalHeight / newRatio,
        maxWidth: 'none',
        maxHeight: 'none',
        marginLeft: -cropData.x,
        marginTop: -cropData.y
      });

      this.center($target);
    },

    center: function ($target) {
      $target.each(function () {
        var $this = $(this),
            $canvas = $this.parent(),
            $container = $canvas.parent(),
            center = function () {
              var containerHeight = $container.height(),
                  canvasHeight = $canvas.height(),
                  marginTop = 'auto';

              if (canvasHeight < containerHeight) {
                marginTop = (containerHeight - canvasHeight) / 2;
              }

              $canvas.css('margin-top', marginTop);
            };

        if (this.complete) {
          center.call(this);
        } else {
          this.onload = center;
        }
      });
    },

    destroy: function () {
      this.unbind();
      this.$element.removeData(NAMESPACE);
    }
  };

  QorCropper.DEFAULTS = {
    parent: false,
    output: false,
    list: false,
    modal: '.qor-cropper-modal',
    save: '.qor-cropper-save',
    key: 'data',
    data: null
  };

  QorCropper.TOGGLE = ('<div class="qor-cropper-toggle"></div>');
  QorCropper.CANVAS = ('<div class="qor-cropper-canvas"></div>');
  QorCropper.TEMPLATE = ('<ul><li><img></li></ul>');

  QorCropper.plugin = function (options) {
    return this.each(function () {
      var $this = $(this),
          data = $this.data(NAMESPACE),
          fn;

      if (!data) {
        if (!$.fn.cropper) {
          return;
        }

        $this.data(NAMESPACE, (data = new QorCropper(this, options)));
      }

      if (typeof options === 'string' && $.isFunction((fn = data[options]))) {
        fn.apply(data);
      }
    });
  };

  $(function () {
    var selector = '.qor-file-input',
        options = {
          parent: '.form-group',
          output: '.qor-file-options',
          list: '.qor-file-list',
          key: 'CropOptions',
          data: {
            Crop: true
          }
        };

    $(document)
      .on('click.qor.cropper.initiator', selector, function () {
        QorCropper.plugin.call($(this), options);
      })
      .on('renew.qor.initiator', function (e) {
        var $element = $(selector, e.target);

        if ($element.length) {
          QorCropper.plugin.call($element, options);
        }
      })
      .triggerHandler('renew.qor.initiator');
  });

  return QorCropper;

});
