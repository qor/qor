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

  var NAMESPACE = 'qor.redactor',
      EVENT_CLICK = 'click.' + NAMESPACE,
      EVENT_FOCUS = 'focus.' + NAMESPACE,
      EVENT_BLUR = 'blur.' + NAMESPACE,
      EVENT_IMAGE_UPLOAD = 'imageupload.' + NAMESPACE,
      EVENT_IMAGE_DELETE = 'imagedelete.' + NAMESPACE,
      REGEXP_OPTIONS = /x|y|width|height/,

      QorRedactor = function (element, options) {
        this.$element = $(element);
        this.options = $.extend(true, {}, QorRedactor.DEFAULTS, options);
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
        x: Number(nums[0]),
        y: Number(nums[1]),
        width: Number(nums[2]),
        height: Number(nums[3])
      };
    }

    return data;
  }

  function capitalize (str) {
    if (typeof str === 'string') {
      str = str.charAt(0).toUpperCase() + str.substr(1);
    }

    return str;
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

  QorRedactor.prototype = {
    constructor: QorRedactor,

    init: function () {
      var _this = this,
          $this = this.$element,
          options = this.options,
          $parent = $this.closest(options.parent),
          click = $.proxy(this.click, this);

      if (!$parent.length) {
        $parent = $this.parent();
      }

      this.$button = $(QorRedactor.BUTTON);
      this.$modal = $parent.find(options.modal);

      $this.on(EVENT_IMAGE_UPLOAD, function (e, image) {
        $(image).on(EVENT_CLICK, click);
      }).on(EVENT_IMAGE_DELETE, function (e, image) {
        $(image).off(EVENT_CLICK, click);
      }).on(EVENT_FOCUS, function (e) {
        // console.log(e.type);
        $parent.find('img').off(EVENT_CLICK, click).on(EVENT_CLICK, click);
      }).on(EVENT_BLUR, function (e) {
        // console.log(e.type);
        $parent.find('img').off(EVENT_CLICK, click);
      });

      $('body').on(EVENT_CLICK, function () {
        _this.$button.off(EVENT_CLICK).detach();
      });
    },

    click: function (e) {
      var _this = this,
          $image = $(e.target);

      e.stopPropagation();

      setTimeout(function () {
        _this.$button.insertBefore($image).off(EVENT_CLICK).one(EVENT_CLICK, function () {
          _this.crop($image);
        });
      }, 1);
    },

    crop: function ($image) {
      var options = this.options,
          url = $image.attr('src'),
          originalUrl = url,
          $clone = $('<img>'),
          $modal = this.$modal;

      if ($.isFunction(options.replace)) {
        originalUrl = options.replace(originalUrl);
      }

      $clone.attr('src', originalUrl);
      $modal.one('shown.bs.modal', function () {
        $clone.cropper({
          data: decodeCropData($image.attr('data-crop-options')),
          background: false,
          zoomable: false,
          rotatable: false,
          checkImageOrigin: false,

          built: function () {
            $modal.find(options.save).one('click', function () {
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
                  CropOptions: {
                    original: getCapitalizeKeyObject(cropData)
                  },
                  Crop: true
                }),
                dataType: 'json',

                success: function (response) {
                  if ($.isPlainObject(response) && response.url) {
                    $image.attr('src', response.url).attr('data-crop-options', encodeCropData(cropData)).removeAttr('style').removeAttr('rel');

                    if ($.isFunction(options.complete)) {
                      options.complete();
                    }

                    $modal.modal('hide');
                  }
                }
              });
            });
          }
        });
      }).one('hidden.bs.modal', function () {
        $clone.cropper('destroy').remove();
      }).modal('show').find('.modal-body').append($clone);
    }
  };

  QorRedactor.DEFAULTS = {
    remote: false,
    toggle: false,
    parent: false,
    modal: '.qor-cropper-modal',
    save: '.qor-cropper-save',
    replace: null,
    complete: null
  };

  QorRedactor.BUTTON = '<span class="redactor-image-cropper">Crop</span>';

  QorRedactor.plugin = function (/* options */) {
    return this.each(function () {
      var $this = $(this),
          data;

      if (!$this.data(NAMESPACE)) {
        if (!$.fn.redactor) {
          return;
        }

        $this.data(NAMESPACE, true);
        data = $this.data();

        $this.redactor({
          imageUpload: data.uploadUrl,
          fileUpload: data.uploadUrl,

          initCallback: function () {
            $this.data(NAMESPACE, new QorRedactor($this, {
              remote: data.cropUrl,
              toggle: '.redactor-image-cropper',
              parent: '.form-group',
              replace: function (url) {
                return url.replace(/\.\w+$/, function (extension) {
                  return '.original' + extension;
                });
              },
              complete: $.proxy(function () {
                this.code.sync();
              }, this)
            }));
          },

          focusCallback: function (/*e*/) {
            $this.triggerHandler(EVENT_FOCUS);
          },

          blurCallback: function (/*e*/) {
            $this.triggerHandler(EVENT_BLUR);
          },

          imageUploadCallback: function (/*image, json*/) {
            $this.triggerHandler(EVENT_IMAGE_UPLOAD, arguments[0]);
          },

          imageDeleteCallback: function (/*url, image*/) {
            $this.triggerHandler(EVENT_IMAGE_DELETE, arguments[1]);
          }
        });
      }
    });
  };

  $(function () {
    $(document)
      .on('renew.qor.initiator', function (e) {
        var $element = $('.qor-textarea', e.target);

        if ($element.length) {
          QorRedactor.plugin.call($element);
        }
      })
      .triggerHandler('renew.qor.initiator');
  });

  return QorRedactor;

});
