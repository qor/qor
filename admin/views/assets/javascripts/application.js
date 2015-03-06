$(function() {
  'use strict';

  var $editors = $('.redactor-editor');
  $editors.each(function() {
    $(this).redactor({
      imageUpload: $(this).data("upload-url"),
      fileUpload: $(this).data("upload-url")
    });
  });

  // crop images
  var $image = $(".image-cropper");

  var $optionInput = $(".image-cropper-crop-option");
  
  if (!$optionInput.length) {
    $optionInput = $('<textarea name="QorResource.File" style="display:none">');
    $('#QorResourceFile').after($optionInput);
  }

  $image.cropper({
    done: function(data) {
      $optionInput.val(JSON.stringify({CropOption: $image.cropper('getData', true)}));
    },
    multiple: true,
    zoomable: false
  });

  // if (window.URL) {
  //   var $inputImage = $("input.image-cropper-upload"),
  //       blobURL;
  //
  //   $inputImage.on('change', function () {
  //     var files = this.files, file;
  //
  //     if (files && files.length) {
  //       file = files[0];
  //
  //       if (/^image\/\w+$/.test(file.type)) {
  //         if (blobURL) { // also can be done with FileReader
  //           URL.revokeObjectURL(blobURL); // Revoke the old one
  //         }
  //
  //         blobURL = URL.createObjectURL(file);
  //         $image.cropper("reset", true).cropper("replace", blobURL);
  //         $inputImage.val("");
  //       }
  //     }
  //   });
  // }

});
