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

  function QorSelector() {
    var $this = $(this);

    if (!$this.prop('multiple') && !$this.find('option[selected]').length) {
      $this.prepend('<option value="" selected></option>');
    }

    $this.chosen();
  }

  $(function () {
    if (!$.fn.chosen) {
      return;
    }

    $(document)
      .on('renew.qor.initiator', function (e) {
        var $element = $('select[data-toggle="qor.selector"]', e.target);

        if ($element.length) {
          QorSelector.call($element);
        }
      })
      .triggerHandler('renew.qor.initiator');
  });

});
