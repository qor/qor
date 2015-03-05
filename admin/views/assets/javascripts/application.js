$(function() {
  var $editors = $('.redactor-editor');
  $editors.each(function() {
    $(this).redactor({
      imageUpload: $(this).data("upload-url"),
      fileUpload: $(this).data("upload-url")
    });
  });

  // crop images
  var $image = $(".image-cropper");

  $image.cropper({
    done: function(data) {
      console.log(data)
    },
    multiple: true,
    zoomable: false,
  });

  if (window.URL) {
    var $inputImage = $("input.image-cropper-upload"),
      blobURL;

    $inputImage.change(function () {
      var files = this.files,
      file;

      if (files && files.length) {
        file = files[0];

        if (/^image\/\w+$/.test(file.type)) {
          if (blobURL) {
            URL.revokeObjectURL(blobURL); // Revoke the old one
          }

          blobURL = URL.createObjectURL(file);
          $image.cropper("reset", true).cropper("replace", blobURL);
          $inputImage.val("");
        }
      }
    });
  }
});
