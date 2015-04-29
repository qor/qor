(function (factory) {
  if (typeof define === 'function' && define.amd) {
    // AMD. Register as anonymous module.
    define('qor-widgets', ['jquery'], factory);
  } else if (typeof exports === 'object') {
    // Node / CommonJS
    factory(require('jquery'));
  } else {
    // Browser globals.
    factory(jQuery);
  }
})(function ($) {

  'use strict';

  var Widgets = {};

  Widgets.init = function () {
    this.confirm();
    this.checkAll();
    this.tooltip();
  };

  Widgets.confirm = function () {
    $(document).on('click.qor.confirmer', '[data-confirm]', function (e) {
      var message = $(this).data('confirm');

      if (message && !window.confirm(message)) {
        e.preventDefault();
      }
    });
  };

  Widgets.checkAll = function () {
    $('.qor-check-all').each(function () {
      var $this = $(this);

      $this.attr('title', 'Check all').tooltip().on('click', function () {
        if (this.disabled) {
          return;
        }

        $(this).attr('data-original-title', this.checked ? 'Uncheck all' : 'Check all').closest('table').find(':checkbox:not(.qor-check-all)').prop('checked', this.checked);
      });

      if (this.checked) {
        $this.triggerHandler('click');
      }
    });
  };

  Widgets.tooltip = function () {
    $('[data-toggle="tooltip"]').tooltip();
  };

  $(function () {
    Widgets.init();
  });

  return Widgets;

});
