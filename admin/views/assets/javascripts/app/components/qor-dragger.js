(function (factory) {
  if (typeof define === 'function' && define.amd) {
    // AMD. Register as anonymous module.
    define('qor-dragger', ['jquery'], factory);
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
    $('.qor-drag').prop('draggable', true).on({
      dragstart: function (e) {
        var dataTransfer = e.originalEvent.dataTransfer;

        dataTransfer.effectAllowed = 'copy';
        dataTransfer.setData('text/html', $(this).find('.qor-drag-data').html());
      }
    });

    $('.qor-drop').on({
      dragenter: function () {
        $(this).addClass('hover');
      },
      dragover: function (e) {
        e.preventDefault();
      },
      dragleave: function () {
        $(this).removeClass('hover');
      },
      drop: function (e) {
        $(this).removeClass('hover').append(e.originalEvent.dataTransfer.getData('text/html'));
      }
    });
  });

});
