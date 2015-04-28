(function (factory) {
  if (typeof define === 'function' && define.amd) {
    // AMD. Register as anonymous module.
    define('qor-selector', ['jquery'], factory);
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
    if (!$.fn.chosen) {
      return;
    }

    $('select[data-toggle="qor.selector"]').each(function () {
      var $this = $(this);

      if (!$this.prop('multiple') && !$this.find('option[selected]').length) {
        $this.prepend('<option value="" selected></option>');
      }

      $this.chosen();
    });
  });

});
