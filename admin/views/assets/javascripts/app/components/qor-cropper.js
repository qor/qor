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
      var $cropper,
          $toggle,
          $modal;

      if (this.built) {
        return;
      }

      this.built = true;

      this.$cropper = $cropper = $(QorCropper.TEMPLATE).appendTo(this.$parent);
      this.$canvas = $cropper.find('.qor-cropper-canvas').html(this.$image);
      this.$toggle = $toggle = $cropper.find('.qor-cropper-toggle');
      this.$modal = $modal = $cropper.find('.qor-cropper-modal');

      $modal.on({
        'shown.bs.modal': $.proxy(this.start, this),
        'hidden.bs.modal': $.proxy(this.stop, this)
      });

      $toggle.on('click', function () {
        $modal.modal();
      });
    },

    start: function () {
      var $modal = this.$modal,
          $clone = $('<img>').attr('src', this.url),
          data = this.data,
          key = this.options.key,
          _this = this;

      $modal.find('.modal-body').html($clone);
      $clone.cropper({
        data: data[key],
        background: false,
        zoomable: false,
        rotatable: false,
        checkImageOrigin: false,

        built: function () {
          $modal.find('.qor-cropper-save').one('click', function () {
            var cropData = $clone.cropper('getData'),
                url;

            data[key] = {
              x: Math.round(cropData.x),
              y: Math.round(cropData.y),
              width: Math.round(cropData.width),
              height: Math.round(cropData.height)
            };

            _this.imageData = $clone.cropper('getImageData');
            _this.cropData = cropData;

            try {
              url = $clone.cropper('getCroppedCanvas').toDataURL();
            } catch (e) {
              console.log(e.message);
            }

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
      var outputData = $.extend({}, this.data, this.options.data);

      if (url) {
        this.$image.attr('src', url);
      } else {
        this.preview();
      }

      this.$output.val(JSON.stringify(outputData));
    },

    preview: function () {
      var $cropper = this.$cropper,
          containerWidth = Math.max($cropper.width(), 320), // minContainerWidth: 320
          containerHeight = Math.max($cropper.height(), 180), // minContainerHeight: 180
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

      this.$canvas.css({
        width: cropData.width,
        height: cropData.height
      });

      this.$image.css({
        width: imageData.naturalWidth / newRatio,
        height: imageData.naturalHeight / newRatio,
        maxWidth: 'none',
        maxHeight: 'none',
        marginLeft: -cropData.x,
        marginTop: -cropData.y
      });
    },

    destroy: function () {
      this.$element.off('change');

      if (this.built) {
        this.$toggle.off('click');
        this.$modal.off('shown.bs.modal hidden.bs.modal');
      }
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
      '<div class="qor-cropper-canvas"></div>' +
      '<a class="qor-cropper-toggle" title="Crop the image"><span class="sr-only">Toggle Cropper</span></a>' +
      '<div class="modal fade qor-cropper-modal" tabindex="-1" role="dialog" aria-hidden="true">' +
        '<div class="modal-dialog">' +
          '<div class="modal-content">' +
            '<div class="modal-header">' +
              '<h5 class="modal-title">Crop the image</h5>' +
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

  QorCropper.plugin = function (options) {
    var args = [].slice.call(arguments, 1),
        result;

    this.each(function () {
      var $this = $(this),
          data = $this.data('qor.cropper'),
          fn;

      if (!data) {
        $this.data('qor.cropper', (data = new QorCropper(this, options)));
      }

      if (typeof options === 'string' && $.isFunction((fn = data[options]))) {
        result = fn.apply(data, args);
      }
    });

    return typeof result === 'undefined' ? this : result;
  };

  $(function () {
    var selector = '.qor-fileinput',
        options = {
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
