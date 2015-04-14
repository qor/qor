(function (factory) {
  if (typeof define === 'function' && define.amd) {
    // AMD. Register as anonymous module.
    define('qor-redactor', ['jquery'], factory);
  } else if (typeof exports === 'object') {
    // Node / CommonJS
    factory(require('jquery'));
  } else {
    // Browser globals.
    factory(jQuery);
  }
})(function ($) {

  'use strict';

  var REGEXP_OPTIONS = /x|y|width|height/,

      QorRedactor = function (element, options) {
        this.$element = $(element);
        this.options = $.extend(true, {}, QorRedactor.DEFAULTS, options);
        this.built = false;
        this.init();
      };

  function encodeCropData(data) {
    var nums = [];

    if ($.isPlainObject(data)) {
      $.each(data, function () {
        nums.push(arguments[1]);
      });
    }

    return nums.join();
  }

  function decodeCropData(data) {
    var nums = data && data.split(',');

    data = null;

    if (nums && nums.length === 4) {
      data = {
        x: nums[0],
        y: nums[1],
        width: nums[2],
        height: nums[3]
      }
    }

    return data;
  }

  QorRedactor.prototype = {
    constructor: QorRedactor,

    init: function () {
      var _this = this,
          $this = this.$element,
          options = this.options,
          $parent,
          $button;

      if (options.parent) {
        $parent = $this.closest(options.parent);
      }

      if (!$parent.length) {
        $parent = $this.parent();
      }

      $button = $(QorRedactor.BUTTON);

      $parent.on('click', 'img', function (e) {
        var $this = $(this),
            originalEvent = e.originalEvent;

        if (originalEvent) {
          originalEvent.qorCropperReady = true;
        }

        $button.insertBefore($this).one('click', function () {
          _this.crop($this);
        });
      });

      $('body').on('click', function (e) {
        var originalEvent = e.originalEvent;

        if (originalEvent && !originalEvent.qorCropperReady) {
          $button.detach();
        }
      });
    },

    crop: function ($image) {
      var options = this.options,
          url = $image.attr('src'),
          originalUrl = url,
          $clone = $('<img>'),
          $modal = $(QorRedactor.TEMPLATE);

      if ($.isFunction(options.replace)) {
        originalUrl = options.replace(originalUrl);
      }

      $clone.attr('src', originalUrl);
      $modal.appendTo('body').modal('show').find('.modal-body').append($clone);

      $modal.one('shown.bs.modal', function () {
        $clone.cropper({
          background: false,
          zoomable: false,
          rotatable: false,

          built: function () {
            var data = decodeCropData($image.attr('data-crop-option')),
                canvasData,
                imageData;

            if ($.isPlainObject(data)) {
              imageData = $clone.cropper('getImageData');
              canvasData = $clone.cropper('getCanvasData');
              imageData.ratio = imageData.width / imageData.naturalWidth;

              $clone.cropper('setCropBoxData', {
                left: data.x * imageData.ratio + canvasData.left,
                top: data.y * imageData.ratio + canvasData.top,
                width: data.width * imageData.ratio,
                height: data.height * imageData.ratio
              });
            }

            $modal.find('.qor-cropper-save').one('click', function () {
              var cropData = {};

              $.each($clone.cropper('getData'), function (i, n) {
                if (REGEXP_OPTIONS.test(i)) {
                  cropData[i] = Math.round(n);
                }
              });

              $.ajax(options.remote, {
                type: 'POST',
                contentType: 'application/json',
                data: JSON.stringify({
                  Url: url,
                  CropOption: cropData,
                  Crop: true
                }),
                dataType: 'json',

                success: function (response) {
                  if ($.isPlainObject(response) && response.url) {
                    $image.attr('src', response.url).attr('data-crop-option', encodeCropData(cropData)).removeAttr('style').removeAttr('rel');

                    if ($.isFunction(options.complete)) {
                      options.complete();
                    }

                    $modal.modal('hide');
                  }
                },

                error: function () {
                  console.log(arguments);
                }
              });
            });
          }
        });
      }).one('hidden.bs.modal', function () {
        $clone.cropper('destroy').remove();
        $modal.remove();
      });
    }
  };

  QorRedactor.DEFAULTS = {
    remote: false,
    parent: false,
    replace: null,
    complete: null
  };

  QorRedactor.BUTTON = '<span id="redactor-image-cropper">Crop</span>';

  QorRedactor.TEMPLATE = (
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
    '</div>'
  );

  $(function () {
    $('.qor-text-editor').each(function () {
      var $this = $(this),
          data = $this.data();

      $this.redactor({
        imageUpload: data.uploadUrl,
        fileUpload: data.uploadUrl,

        initCallback: function () {
          if (!$this.data('qor.redactor')) {
            $this.data('qor.redactor', new QorRedactor($this, {
              remote: data.cropUrl,
              parent: '.redactor-editor',
              replace: function (url) {
                return url.replace(/\.\w+$/, function (extension) {
                  return '.original' + extension;
                });
              },
              complete: $.proxy(function () {
                this.code.sync();
              }, this)
            }));
          }
        }
      });
    });
  });

  return QorRedactor;

});
