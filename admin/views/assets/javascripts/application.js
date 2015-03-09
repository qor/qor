$(function() {
  'use strict';

  var $editors = $('.redactor-editor');
  $editors.each(function() {
    $(this).redactor({
      imageUpload: $(this).data("upload-url"),
      fileUpload: $(this).data("upload-url")
    });
  });

  $('.image-cropper-upload').clipper();

});
