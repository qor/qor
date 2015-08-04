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

  var NAMESPACE = 'qor.textviewer';
  var EVENT_ENABLE = 'enable.' + NAMESPACE;
  var EVENT_DISABLE = 'disable.' + NAMESPACE;

  function QorTextviewer(element, options) {
    this.$element = $(element);
    this.options = $.extend({}, QorTextviewer.DEFAULTS, $.isPlainObject(options) && options);
    this.init();
  }

  QorTextviewer.prototype = {
    constructor: QorTextviewer,

    init: function () {
      this.$element.find(this.options.toggle).each(function () {
        var $this = $(this);

        // 8 for correction as `scrollHeight` will large than `offsetHeight`.
        if (this.scrollHeight > this.offsetHeight + 8) {
          $this.after(QorTextviewer.TEMPLATE);
        } else {
          $this.addClass('viewable');
        }
      });
    },

    destroy: function () {
      this.$element.removeData(NAMESPACE);
    },
  };

  QorTextviewer.DEFAULTS = {
    toggle: false,
  };

  QorTextviewer.TEMPLATE = '<p class="qor-list-ellipsis">...</p>';

  QorTextviewer.plugin = function (options) {
    return this.each(function () {
      var $this = $(this);
      var data = $this.data(NAMESPACE);
      var fn;

      if (!data) {
        if (/destroy/.test(options)) {
          return;
        }

        $this.data(NAMESPACE, (data = new QorTextviewer(this, options)));
      }

      if (typeof options === 'string' && $.isFunction(fn = data[options])) {
        fn.apply(data);
      }
    });
  };

  $(function () {
    var selector = '.qor-list';
    var options = {
          toggle: '.qor-list-text',
        };

    $(document)
      .on(EVENT_DISABLE, function (e) {
        QorTextviewer.plugin.call($(selector, e.target), 'destroy');
      })
      .on(EVENT_ENABLE, function (e) {
        QorTextviewer.plugin.call($(selector, e.target), options);
      })
      .triggerHandler(EVENT_ENABLE);
  });

});
