/*
 * jQuery image clipper
 * Copyright (c) 2015 Lancee (xrhy.me)
 * Dual licensed under the MIT and GPL licenses
 */

!(function() {
  (function($, Export) {
    "use strict";

    $.clipper = function(fileInput, options) {
      if (!fileInput || fileInput.nodeName.toLowerCase() !== "input") {
        throw new Error('this is not a input');
      }

      var Clipper = function() {
        this.init();
      }

      Clipper.prototype = {
        constructor: Clipper,

        init: function() {
          options = $.extend({}, $.clipper.defaults, options);

          var $input = $(fileInput).data('clipper', this),
              src = $input.val(),
              suffix = src.split('.').reverse()[0].toLowerCase(),
              me = this;

          var $image = $(options.imageSelector);

          me.$el = $input.addClass('clipper');

          if (isImg(suffix) && $image.length !== 0) {
            $image = createImg(src);
          }

          if (!window.URL) {
            return;
          }

          var $cropperDataHolder = $(options.cropperDataHolderSelector);

          if (!$cropperDataHolder.length) {
            $cropperDataHolder = $(options.cropperDataHolderTemplate);
            $input.before($cropperDataHolder);
          }

          $image.cropper({
            done: function(data) {
              $cropperDataHolder.val(JSON.stringify({CropOption: $image.cropper('getData', true), Crop: true}));
            },
            multiple: true,
            zoomable: false
          });

          var blobURL = '';

          $input.on('change', function(e) {
            var files = this.files, file = files[0];

            if (file && isImg(file.type)) {
              if (!($image.length && $image[0].nodeName === 'IMG')) {
                $image = createImg();
              }

              if (blobURL) {
                blobURL = URL.revokeObjectURL(blobURL)
              }

              blobURL = URL.createObjectURL(file);

              $image.cropper("reset", true).cropper("replace", blobURL);

            }
          });

          me.options = options;
        },

        options: $.clipper.defaults
      }

      function isImg(suffix) {
        return suffix.search(/jpg|jpeg|png|gif|bmp/) !== -1; 
      }

      function createImg(src) {
        var img = new Image();
        img.src = src;

        $(fileInput).after(img);

        return $(img);
      }

      return new Clipper();

    }

    $.clipper.defaults = {
      imageSelector: '.image-cropper',
      imageClass: 'clipper-uploaded-image',
      cropperDataHolderSelector: '.image-cropper-crop-option',
      cropperDataHolderTemplate: '<textarea name="QorResource.File" style="display:none">'
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