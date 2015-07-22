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

  $(function () {
    $(document).on('click.qor.alert', '[data-dismiss="alert"]', function () {
      $(this).closest('.qor-alert').remove();
    });

    setTimeout(function () {
      $('.qor-alert[data-dismissible="true"]').remove();
    }, 3000);
  });

});
