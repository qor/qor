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

  var console = window.console || (window.console = { log: $.noop }),
      $window = $(window),
      $body = $(document.body),

      Qor = function (element, options) {
        var $element = $(element);

        this.$element = $element;
        this.options = $.extend({}, Qor.DEFAULTS, $.isPlainObject(options) && options);
        this.namespace = $element.data('namespace');
        this.dependency = $element.data('dependency');
        this.init();
      };

  Qor.DEFAULTS = {

  };

  Qor.prototype = {
    constructor: Qor,

    init: function () {
      var dependency = this.dependency;

      this.initNavbar();
      this.initFooter();

      if (!dependency) {
        console.log('No dependency.');
        return;
      }

      console.log(dependency + ' is loading...');

      require([
        dependency,
        'modules/utilities'
      ], function (Controller) {
        if ($.isFunction(Controller)) {
          console.log(dependency + ' is running...');
          return new Controller();
        }
      });
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
      var $footer = $('.footer');

      $window.on('resize', function () {
        $footer.toggleClass('static', $body.height() > $window.height());
      }).triggerHandler('resize');

      $footer.removeClass('invisible');
    }
  };

  $(function () {
    var $main = $('.main');

    $main.data('qor', new Qor($main, window.options));
  });

  return Qor;

});
