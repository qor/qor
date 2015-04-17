(function (factory) {
  if (typeof define === 'function' && define.amd) {
    // AMD. Register as anonymous module.
    define('qor-datepicker', ['jquery'], factory);
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
    var $daterange = $('.qor-daterange');

    $daterange.datepicker({
      autoclose: true,
      inputs: $daterange.find('input').toArray()
    });

    $('.qor-date').datepicker({
      autoclose: true
    });
  });

});
