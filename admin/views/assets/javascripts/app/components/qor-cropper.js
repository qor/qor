(function (factory) {
  if (typeof define === 'function' && define.amd) {
    // AMD. Register as anonymous module.
    define('qor-cropper', ['jquery'], factory);
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

      QorCropper = function (element, options) {
        this.$element = $(element);
        this.options = $.extend(true, {}, QorCropper.DEFAULTS, options);
        this.built = false;
        this.url = null;
        this.init();
      };

  QorCropper.prototype = {
    constructor: QorCropper,

    init: function () {
      var $this = this.$element,
          options = this.options,
          $parent,
          $image,
          data,
          url;

      if (options.parent) {
        $parent = $this.closest(options.parent);
      }

      if (!$parent.length) {
        $parent = $this.parent();
      }

      if (options.target) {
        $image = $parent.find(options.target);
      }

      if (!$image.length) {
        $image = $('<img>');
      }

      if (options.output) {
        this.$output = $parent.find(options.output);

        try {
          data = JSON.parse(this.$output.val());
        } catch (e) {
          console.log(e.message);
        }
      }

      this.$parent = $parent;
      this.$image = $image;
      $this.on('change', $.proxy(this.read, this));

      this.data = data || {};
      url = $image.data('originalUrl');

      if (!url) {
        url = $image.prop('src');

        if (url && $.isFunction(options.replace)) {
          url = options.replace(url);
        }
      }

      this.load(url);
      $this.on('change', $.proxy(this.read, this));
    },

    read: function () {
      var files = this.$element.prop('files'),
          file;

      if (files) {
        file = files[0];

        if (/^image\/\w+$/.test(file.type) && URL) {
          this.load(URL.createObjectURL(file), true);
        }
      }
    },

    load: function (url, replaced) {
      if (!url) {
        return;
      }

      if (!this.built) {
        this.build();
      }

      if (/^blob:\w+/.test(this.url) && URL) {
        URL.revokeObjectURL(this.url); // Revoke the old one
      }

      this.url = url;

      if (replaced) {
        this.data[this.options.key] = null;
        this.$image.attr('src', url);
      }
    },

    build: function () {
      if (this.built) {
        return;
      }

      this.built = true;

      this.$cropper = $(QorCropper.TEMPLATE).prepend(this.$image).appendTo(this.$parent);
      this.$cropper.find('.modal').on({
        'shown.bs.modal': $.proxy(this.start, this),
        'hidden.bs.modal': $.proxy(this.stop, this)
      });
    },

    start: function () {
      var $modal = this.$cropper.find('.modal'),
          $clone = $('<img>').attr('src', this.url),
          data = this.data,
          key = this.options.key,
          _this = this;

      $modal.find('.modal-body').html($clone);
      $clone.cropper({
        background: false,
        zoomable: false,
        rotatable: false,

        built: function () {
          var previous = data[key],
              scaled = {},
              scaledRatio,
              imageData,
              canvasData;

          if ($.isPlainObject(previous)) {
            imageData = $clone.cropper('getImageData');
            canvasData = $clone.cropper('getCanvasData');
            scaledRatio = imageData.width / imageData.naturalWidth;

            $.each(previous, function (key, val) {
              scaled[String(key).toLowerCase()] = val * scaledRatio;
            });

            $clone.cropper('setCropBoxData', {
              left: scaled.x + canvasData.left,
              top: scaled.y + canvasData.top,
              width: scaled.width,
              height: scaled.height
            });
          }

          $modal.find('.qor-cropper-save').one('click', function () {
            var cropData = $clone.cropper('getData');

            data[key] = {
              x: Math.round(cropData.x),
              y: Math.round(cropData.y),
              width: Math.round(cropData.width),
              height: Math.round(cropData.height)
            };

            _this.output($clone.cropper('getCroppedCanvas').toDataURL());
            $modal.modal('hide');
          });
        }
      });
    },

    stop: function () {
      this.$cropper.find('.modal-body > img').cropper('destroy').remove();
    },

    output: function (url) {
      var data = $.extend({}, this.data, this.options.data);

      this.$image.attr('src', url);
      this.$output.val(JSON.stringify(data));
    },

    destroy: function () {
      this.$element.off('change');
      this.$cropper.find('.modal').off('shown.bs.modal hidden.bs.modal');
    }
  };

  QorCropper.DEFAULTS = {
    target: '',
    output: '',
    parent: '',
    key: 'qorCropper',
    data: null
  };

  QorCropper.TEMPLATE = (
    '<div class="qor-cropper">' +
      '<a class="qor-cropper-toggle" data-toggle="modal" href="#qorCropperModal" title="Crop the image"><span class="sr-only">Toggle Cropper<span></a>' +
      '<div class="modal fade qor-cropper-modal" id="qorCropperModal" tabindex="-1" role="dialog" aria-labelledby="qorCropperModalLabel" aria-hidden="true">' +
        '<div class="modal-dialog">' +
          '<div class="modal-content">' +
            '<div class="modal-header">' +
              '<h5 class="modal-title" id="qorCropperModalLabel">Crop the image</h5>' +
            '</div>' +
            '<div class="modal-body"></div>' +
            '<div class="modal-footer">' +
              '<button type="button" class="btn btn-link" data-dismiss="modal">Cancel</button>' +
              '<button type="button" class="btn btn-link qor-cropper-save">OK</button>' +
            '</div>' +
          '</div>' +
        '</div>' +
      '</div>' +
    '</div>'
  );

  $(function () {
    if (!$.fn.cropper) {
      return;
    }

    $('input[data-toggle="qor.cropper"]').each(function () {
      var $this = $(this);

      if (!$this.data('qor.cropper')) {
        $this.data('qor.cropper', new QorCropper(this, {
          target: 'img',
          output: 'textarea',
          parent: '.form-group',
          key: 'CropOption',
          data: {
            Crop: true
          },
          replace: function (url) {
            return url.replace(/\.\w+$/, function (extension) {
              return '.original' + extension;
            });
          }
        }));
      }
    });
  });

  return QorCropper;

});
