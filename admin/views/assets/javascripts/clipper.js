/*
 * jQuery image clipper
 * Copyright (c) 2015 Lancee (xrhy.me)
 * Dual licensed under the MIT and GPL licenses
 */

!(function() {
  (function($, Export) {
    "use strict";

    $.clipper = function(fileInput, options) {
      if (!fileInput || fileInput.nodeName !== "INPUT") {
        throw new Error('this is not a input');
      }

      var Clipper = function() {
        this.init();
      }

      Clipper.prototype = {
        constructor: Clipper,

        init: function() {
          var me = this,
              $image = build.call(this),
              blobURL = '';

          $(fileInput).on('change', function(e) {
            if (!$(this).hasClass('cropper-clipper')) {
              $image = build.call(me, e);
            }

            if ($image.parent().find('.cropper-container').length !== 0) {
              $image.cropper('destroy');
              $image.parent().find('.crop-confirm-wrapper').hide();
              $image.parent().find('.clipper-btn').show();
            }

            $('#cropper-data-' + me.id).val('');

            var files = this.files, file = files[0];

            if (file && isImg(file.type)) {
              if (blobURL) {
                blobURL = URL.revokeObjectURL(blobURL);
              }

              blobURL = URL.createObjectURL(file);

              $image.attr('src', blobURL).data('origin', blobURL);

            }
          });

          me.options = options;
        },

        reset: function() {
          var $image = build();
          $image.cropper('reset');
        },

        options: $.clipper.defaults
      } //Clipper.prototype

      function build(event) {
        options = $.extend({}, $.clipper.defaults, options);

        var $image = $(options.imageSelector);

        if (!event && $image.length === 0 || !window.URL) {
          return;
        }

        if (event && $image.length !== 0) {
          $('#cropper-data-' + this.id).val('');
          return $image;
        }

        var $input = $(fileInput).data('clipper', this),
            filePath = $input.val(),
            me = this;

        this.id = uuid();

        if ($image.length === 0) {
          $image = createImg();
        }

        me.$el = $input.addClass('clipper'); // can be $image?

        var $cropperDataHolder = $(options.cropperDataHolderSelector);

        if (!$cropperDataHolder.length) {
          $cropperDataHolder = $(options.cropperDataHolderTemplate);
          $input.before($cropperDataHolder);
        }

        $cropperDataHolder.attr('id', 'cropper-data-' + this.id);

        $image.data('origin', $image[0].src).wrap('<figure class="figure clipper-image-wrapper"></figure>');

        var $cropBtn = $('<a href="javascript:;" class="clipper-btn">Crop</a>'),
            $cropConfirm = $(options.cropperConfirmTemplate);

        $image.before($cropBtn).before($cropConfirm);

        $cropBtn.on('click', function() {
          $(this).hide();
          $cropConfirm.show();

          $image.attr('src', $image.data('origin').replace(/(jpg|jpeg|png|gif|bmp)$/, 'original.$1'));

          var data = JSON.parse($cropperDataHolder.val() || '{}').CropOption;

          for (var k in data) {
            data[k.toLowerCase()] = data[k];
          }

          $image.cropper({
            built: function() {
              if (!data) {
                return;
              }

              var imageData = $(this).cropper('getImageData', true),
                  cropData = {};

              var zoomLevel = imageData.width / imageData.naturalWidth;

              cropData.left = imageData.left + zoomLevel * data.x;
              cropData.top = imageData.top + zoomLevel * data.y;
              cropData.width = data.width * zoomLevel;
              cropData.height = data.height * zoomLevel;

              $(this).addClass('cropper-clipper').cropper('setCropBoxData', cropData);
            },
            crop: function(data) {},
            multiple: true,
            zoomable: false
          });
          
        });

        $cropConfirm.on('click', '.act', function(e) {
          var act = $(e.target).data('act'),
              data = $image.cropper('getData', true);

          data = !!(act*1) ? JSON.stringify({CropOption: data, Crop: !!(act*1)}) : '';

          $cropperDataHolder.val(data);

          var dataURL = act*1 ? $image.cropper('getDataURL') : $image.data('origin');

          $image.cropper('destroy');
          $image.attr('src', dataURL);

          $cropConfirm.hide();

          $cropBtn.show();
        });

        return $image;
      }

      function isImg(suffix) {
        return suffix.search(/jpg|jpeg|png|gif|bmp/i) !== -1;
      }

      function createImg(src) {
        var img = new Image();
        img.src = src;

        $(fileInput).after(img);

        return $(img);
      }

      // https://gist.github.com/LeverOne/1308368
      function uuid(a,b){for(b=a='';a++< 36;b+=a * 51&52?(a^15?8^Math.random()*(a^20?16:4):4).toString(16):'-');return b}

      return new Clipper();

    }

    $.clipper.defaults = {
      imageSelector: '.image-cropper',
      imageClass: 'clipper-uploaded-image',
      cropperDataHolderSelector: '.image-cropper-crop-option',
      cropperDataHolderTemplate: '<textarea name="QorResource.File" style="display:none">',
      cropperConfirmTemplate: '<div class="crop-confirm-wrapper" style="display: none;">\
                                 <a href="javascript:;" data-act="0" class="btn act cancel">Cancel</a>\
                                 <a href="javascript:;" data-act="1" class="btn act save">Save</a>\
                               </div>'
    };

    $.fn.clipper = function(options, callback) {
      var clipper = $(this).data('clipper');

      if ($.isFunction(options)) {
        callback = options;
        options = null;
      } else {
        options = options || {}; 
      }

      if(typeof(options) === 'object') {
        return this.each(function(i) {
          if(!clipper) {
            clipper = $.clipper(this, options);
            if(callback)
              callback.call(clipper);
          } else {
            if(callback)
              callback.call(clipper);
          }
        });
      } else {
        throw new Error('arguments[0] is not a instance of Object');
      }
    }

  })(jQuery, window);

}).call(this);