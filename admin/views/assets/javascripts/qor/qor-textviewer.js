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
    $('.qor-list').find('.qor-list-text').each(function () {
      var $this = $(this);

      // 8 for correction as `scrollHeight` will large than `offsetHeight`.
      if (this.scrollHeight > this.offsetHeight + 8) {
        $this.addClass('scrollable');
      }
    });
  });

});
