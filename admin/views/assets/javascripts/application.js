$(function() {
  'use strict';

  var $editors = $('.redactor-editor');
  $editors.each(function(i, me) {
    var cropURL = $(this).data('crop-url'), 
        uploadURL = $(this).data("upload-url");

    $(me).redactor({
      plugins: ['clipper'],
      imageUpload: uploadURL,
      fileUpload: uploadURL,
      initCallback: function() {
        $(me).after('<div id="crop-data-wrapper"></div>');
      },
      imageUploadCallback: function($image) {
        var src = $image[0].src;
        var $cropDataHolder = $('<textarea id="redactor-crop-data'+ src +'" style="display:none">');

        $('#crop-data-wrapper').append($cropDataHolder);

      },
      modalOpenedCallback: function(name) {

      }
    });
  });

  $('.image-cropper-upload').clipper();

});
