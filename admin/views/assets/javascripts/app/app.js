$(function () {

  'use strict';

  var location = window.location;

  $('.qor-lang-selector, .qor-locale-selector').on('change', function () {
    var url = $(this).val();

    if (url) {
      location.assign(url);
    }
  });

});
