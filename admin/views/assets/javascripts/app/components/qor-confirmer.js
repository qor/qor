(function (factory) {
  if (typeof define === 'function' && define.amd) {
    // AMD. Register as anonymous module.
    define('qor-confirmer', ['jquery'], factory);
  } else if (typeof exports === 'object') {
    // Node / CommonJS
    factory(require('jquery'));
  } else {
    // Browser globals.
    factory(jQuery);
  }
})(function ($) {

  'use strict';

  $(document).on('click.qor.confirmer', '[data-confirm]', function (e) {
    var message = $(this).data('confirm');

    if (message && !window.confirm(message)) {
      e.preventDefault();
    }
  });

});
