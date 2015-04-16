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
          data;

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
      this.data = data || {};
      this.load($image.data('originalUrl') || $image.prop('src'));
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
    $('.qor-file-input').each(function () {
      var $this = $(this);

      if (!$this.data('qor.cropper')) {
        $this.data('qor.cropper', new QorCropper(this, {
          target: '.qor-file-image',
          output: '.qor-file-options',
          parent: '.form-group',
          key: 'CropOption',
          data: {
            Crop: true
          }
        }));
      }
    });
  });

  return QorCropper;

});

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

  var NAMESPACE = '.qor.redactor',
      EVENT_CLICK = 'click' + NAMESPACE,
      EVENT_FOCUS = 'focus' + NAMESPACE,
      EVENT_BLUR = 'blur' + NAMESPACE,
      EVENT_IMAGE_UPLOAD = 'imageupload' + NAMESPACE,
      EVENT_IMAGE_DELETE = 'imagedelete' + NAMESPACE,
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
        x: nums[0],
        y: nums[1],
        width: nums[2],
        height: nums[3]
      };
    }

    return data;
  }

  QorRedactor.prototype = {
    constructor: QorRedactor,

    init: function () {
      var _this = this,
          $this = this.$element,
          options = this.options,
          $parent = $this.closest(options.parent),
          click = $.proxy(this.click, this);

      this.$button = $(QorRedactor.BUTTON);

      $this.on(EVENT_IMAGE_UPLOAD, function (e, image) {
        $(image).on(EVENT_CLICK, click);
      }).on(EVENT_IMAGE_DELETE, function (e, image) {
        $(image).off(EVENT_CLICK, click);
      }).on(EVENT_FOCUS, function (e) {
        console.log(e.type);
        $parent.find('img').off(EVENT_CLICK, click).on(EVENT_CLICK, click);
      }).on(EVENT_BLUR, function (e) {
        console.log(e.type);
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
    toggle: false,
    parent: false,
    replace: null,
    complete: null
  };

  QorRedactor.BUTTON = '<span class="redactor-image-cropper">Crop</span>';

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
          }
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
    });
  });

  return QorRedactor;

});

(function (factory) {
  if (typeof define === 'function' && define.amd) {
    // AMD. Register as anonymous module.
    define('selector', ['jquery'], factory);
  } else if (typeof exports === 'object') {
    // Node / CommonJS
    factory(require('jquery'));
  } else {
    // Browser globals.
    factory(jQuery);
  }
})(function ($) {

  'use strict';

  $(function () {
    $('.qor-select').chosen({
      allow_single_deselect: true
    });
  });

});

(function (factory) {
  if (typeof define === 'function' && define.amd) {
    // AMD. Register as anonymous module.
    define('qor', ['jquery'], factory);
  } else if (typeof exports === 'object') {
    // Node / CommonJS
    factory(require('jquery'));
  } else {
    // Browser globals.
    factory(jQuery);
  }
})(function ($) {

  'use strict';

  var $window = $(window),

      Qor = function () {
        this.init();
      };

  Qor.prototype = {
    constructor: Qor,

    init: function () {
      this.initNavbar();
      this.initFooter();
      this.initConfirm();
    },

    initNavbar: function () {
      var $navbar = $('.navbar');

      $navbar.find('.dropdown').on({
        mouseover: function () {
          $(this).addClass('open');
        },
        mouseout: function () {
          $(this).removeClass('open');
        }
      });
    },

    initFooter: function () {
      var $footer = $('.footer'),
          $body = $('body');

      $window.on('resize', function () {
        var minHeight = $window.innerHeight();

        if ($body.height() >= minHeight) {
          $footer.addClass('static');
        } else {
          $footer.removeClass('static');
        }
      }).triggerHandler('resize');
    },

    initConfirm: function () {
      $('[data-confirm]').click(function (e) {
        var message = $(this).data('confirm');

        if (message && !window.confirm(message)) {
          e.preventDefault();
        }
      });
    }
  };

  $(function () {
    $('.main').data('qor', new Qor());
  });

});

//# sourceMappingURL=main.js.map