(function (factory) {
  if (typeof define === 'function' && define.amd) {
    // AMD. Register as anonymous module.
    define('qor', ['jquery'], factory);
  } else if (typeof exports === 'object') {
    // Node / CommonJS
    factory(require('jquery'));
  } else {
    // Browser globals.
    factory(jQuery);
  }
})(function ($) {

  'use strict';

  var $window = $(window),

      Qor = function () {
        this.init();
      };

  Qor.prototype = {
    constructor: Qor,

    init: function () {
      this.initNavbar();
      this.initFooter();
      this.initConfirm();
    },

    initNavbar: function () {
      var $navbar = $('.navbar');

      $navbar.find('.dropdown').on({
        mouseover: function () {
          $(this).addClass('open');
        },
        mouseout: function () {
          $(this).removeClass('open');
        }
      });
    },

    initFooter: function () {
      var $footer = $('.footer'),
          $body = $('body');

      $window.on('resize', function () {
        var minHeight = $window.innerHeight();

        if ($body.height() >= minHeight) {
          $footer.addClass('static');
        } else {
          $footer.removeClass('static');
        }
      }).triggerHandler('resize');
    },

    initConfirm: function () {
      $('[data-confirm]').click(function (e) {
        var message = $(this).data('confirm');

        if (message && !confirm(message)) {
          e.preventDefault();
        }
      });
    }
  };

  $(function () {
    $('.main').data('qor', new Qor());
  });

});
