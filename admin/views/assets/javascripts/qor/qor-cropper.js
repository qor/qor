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
  var EVENT_SHOWN = 'shown.bs.modal';
  var EVENT_HIDDEN = 'hidden.bs.modal';

  // Classes
  var CLASS_TOGGLE = '.qor-cropper-toggle';
  var CLASS_CANVAS = '.qor-cropper-canvas';
  var CLASS_CONTAINER = '.qor-cropper-container';
  var CLASS_OPTIONBOX = '.qor-cropper-optionbox';
  var CLASS_SAVE = '.qor-cropper-save';

  // RegExps
  var REGEXP_OPTIONS = /x|y|width|height/;

  function capitalize (str) {
    if (typeof str === 'string') {
      str = str.charAt(0).toUpperCase() + str.substr(1);
    }

    return str;
  }

  function getLowerCaseKeyObject (obj) {
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

  /*function getCapitalizeKeyObject (obj) {
    var newObj = {};
    var key;

    if ($.isPlainObject(obj)) {
      for (key in obj) {
        if (obj.hasOwnProperty(key)) {
          newObj[capitalize(key)] = obj[key];
        }
      }
    }

    return newObj;
  }*/

  function getValueByNoCaseKey (obj, key) {
    var originalKey = String(key);
    var lowerCaseKey = originalKey.toLowerCase();
    var upperCaseKey = originalKey.toUpperCase();
    var capitalizeKey = capitalize(originalKey);

    if ($.isPlainObject(obj)) {
      return (obj[lowerCaseKey] || obj[capitalizeKey] || obj[upperCaseKey]);
    }
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

      this.data = $.extend(data || {}, options.data);
      this.build();
      this.bind();
    },

    build: function () {
      this.wrap();
      this.$modal = $(QorCropper.MODAL).appendTo('body');
    },

    unbuild: function () {
      this.unwrap();
      this.$modal.remove();
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
      this.$modal.modal('show');
    },

    read: function (e) {
      var files = e.target.files;
      var file;

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
      var $list = this.$list;
      var $ul = $list.find('ul');
      var $img;

      if (!$ul.length) {
        $ul  = $(QorCropper.LIST);
        $list.html($ul);
        this.wrap();
      }

      if ($ul.is(':hidden')) {
        $ul.show();
      }

      $img = $list.find('img');
      $img.attr('src', url).data('originalUrl', url);
      this.center($img, true);
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

      $modal.find(CLASS_CONTAINER).html($clone);

      list = this.getList(sizeAspectRatio);

      if (list) {
        $modal.find(CLASS_OPTIONBOX).show().find('ul').html(list);
      }

      $clone.cropper({
        aspectRatio: sizeAspectRatio,
        data: getLowerCaseKeyObject(data[options.key][sizeName]),
        background: false,
        movable: false,
        zoomable: false,
        rotatable: false,
        checkImageOrigin: false,

        built: function () {
          $modal.find(CLASS_SAVE).one('click', function () {
            var cropData = {};
            var syncData = [];
            var url;

            $.each($clone.cropper('getData'), function (i, n) {
              if (REGEXP_OPTIONS.test(i)) {
                cropData[i] = Math.round(n);
              }
            });

            data[options.key][sizeName] = cropData;
            _this.imageData = $clone.cropper('getImageData');
            _this.cropData = cropData;

            try {
              url = $clone.cropper('getCroppedCanvas').toDataURL();
            } catch (e) {}

            $modal.find(CLASS_OPTIONBOX + ' input').each(function () {
              var $this = $(this);

              if ($this.prop('checked')) {
                syncData.push($this.attr('name'));
              }
            });

            _this.output(url, syncData);
            $modal.modal('hide');
          });
        },
      });
    },

    stop: function () {
      this.$modal.
        find(CLASS_CONTAINER + ' > img').
          cropper('destroy').
          remove().
          end().
        find(CLASS_OPTIONBOX).
          hide().
          find('ul').
            empty();
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

      return list.length ? ('<li>' + list.join('</li><li>') + '</li>') : '';
    },

    output: function (url, data) {
      var $target = this.$target;

      if (url) {
        this.center($target.attr('src', url));
      } else {
        this.preview($target);
      }

      if ($.isArray(data) && data.length) {
        this.autoCrop(url, data);
      }

      this.$output.val(JSON.stringify(this.data));
    },

    preview: function ($target) {
      var $canvas = $target.parent();
      var $container = $canvas.parent();

      // minContainerWidth: 160, minContainerHeight: 160
      var containerWidth = Math.max($container.width(), 160);
      var containerHeight = Math.max($container.height(), 160);
      var imageData = this.imageData;

      // Clone one to avoid changing it
      var cropData = $.extend({}, this.cropData);
      var cropAspectRatio = cropData.width / cropData.height;
      var newWidth = containerWidth;
      var newHeight = containerHeight;
      var newRatio;

      if (containerHeight * cropAspectRatio > containerWidth) {
        newHeight = newWidth / cropAspectRatio;
      } else {
        newWidth = newHeight * cropAspectRatio;
      }

      newRatio = cropData.width / newWidth;

      $.each(cropData, function (i, n) {
        cropData[i] = n / newRatio;
      });

      $canvas.css({
        width: cropData.width,
        height: cropData.height,
      });

      $target.css({
        width: imageData.naturalWidth / newRatio,
        height: imageData.naturalHeight / newRatio,
        maxWidth: 'none',
        maxHeight: 'none',
        marginLeft: -cropData.x,
        marginTop: -cropData.y,
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
          $canvas.removeAttr('style');
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
            _this.center($this.attr('src', url));
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
  };

  QorCropper.TOGGLE = '<div class="qor-cropper-toggle"></div>';
  QorCropper.CANVAS = '<div class="qor-cropper-canvas"></div>';
  QorCropper.LIST = '<ul><li><img></li></ul>';
  QorCropper.MODAL = (
    '<div class="modal fade qor-cropper-modal" tabindex="-1" role="dialog" aria-hidden="true">' +
      '<div class="modal-dialog">' +
        '<div class="modal-content">' +
          '<div class="modal-header">' +
            '<h5 class="modal-title">Crop the image</h5>' +
          '</div>' +
          '<div class="modal-body">' +
            '<div class="qor-cropper-container"></div>' +
            '<div class="qor-cropper-optionbox">' +
              '<h5>Sync cropping result to:</h5>' +
              '<ul></ul>' +
            '</div>' +
          '</div>' +
          '<div class="modal-footer">' +
            '<button type="button" class="btn btn-link" data-dismiss="modal">Cancel</button>' +
            '<button type="button" class="btn btn-link qor-cropper-save">OK</button>' +
          '</div>' +
        '</div>' +
      '</div>' +
    '</div>'
  );

  QorCropper.plugin = function (options) {
    return this.each(function () {
      var $this = $(this);
      var data = $this.data(NAMESPACE);
      var fn;

      if (!data) {
        if (!$.fn.cropper) {
          return;
        }

        if (/destroy/.test(options)) {
          return;
        }

        $this.data(NAMESPACE, (data = new QorCropper(this, options)));
      }

      if (typeof options === 'string' && $.isFunction(fn = data[options])) {
        fn.apply(data);
      }
    });
  };

  $(function () {
    var selector = '.qor-file-input';
    var options = {
          parent: '.form-group',
          output: '.qor-file-options',
          list: '.qor-file-list',
          key: 'CropOptions',
          data: {
            Crop: true,
          },
        };

    $(document)
      .on(EVENT_CLICK, selector, function () {
        QorCropper.plugin.call($(this), options);
      })
      .on(EVENT_DISABLE, function (e) {
        QorCropper.plugin.call($(selector, e.target), 'destroy');
      })
      .on(EVENT_ENABLE, function (e) {
        QorCropper.plugin.call($(selector, e.target), options);
      })
      .triggerHandler(EVENT_ENABLE);
  });

  return QorCropper;

});
