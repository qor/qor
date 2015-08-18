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

  var NAMESPACE = 'qor.autoheight';
  var EVENT_ENABLE = 'enable.' + NAMESPACE;
  var EVENT_DISABLE = 'disable.' + NAMESPACE;
  var EVENT_INPUT = 'input';

  function QorAutoheight(element, options) {
    this.$element = $(element);
    this.options = $.extend({}, QorAutoheight.DEFAULTS, $.isPlainObject(options) && options);
    this.init();
  }

  QorAutoheight.prototype = {
    constructor: QorAutoheight,

    init: function () {
      var $this = this.$element;

      this.overflow = $this.css('overflow');
      this.paddingTop = parseInt($this.css('padding-top'), 10);
      this.paddingBottom = parseInt($this.css('padding-bottom'), 10);
      $this.css('overflow', 'hidden');
      this.resize();
      this.bind();
    },

    bind: function () {
      this.$element.on(EVENT_INPUT, $.proxy(this.resize, this));
    },

    unbind: function () {
      this.$element.off(EVENT_INPUT, this.resize);
    },

    resize: function () {
      var $this = this.$element;

      if ($this.is(':hidden')) {
        return;
      }

      $this.height('auto').height($this.prop('scrollHeight') - this.paddingTop - this.paddingBottom);
    },

    destroy: function () {
      this.unbind();
      this.$element.css('overflow', this.overflow).removeData(NAMESPACE);
    },
  };

  QorAutoheight.DEFAULTS = {};

  QorAutoheight.plugin = function (options) {
    return this.each(function () {
      var $this = $(this);
      var data = $this.data(NAMESPACE);
      var fn;

      if (!data) {
        if (/destroy/.test(options)) {
          return;
        }

        $this.data(NAMESPACE, (data = new QorAutoheight(this, options)));
      }

      if (typeof options === 'string' && $.isFunction(fn = data[options])) {
        fn.apply(data);
      }
    });
  };

  $(function () {
    var selector = 'textarea.qor-js-autoheight';

    $(document).
      on(EVENT_DISABLE, function (e) {
        QorAutoheight.plugin.call($(selector, e.target), 'destroy');
      }).
      on(EVENT_ENABLE, function (e) {
        QorAutoheight.plugin.call($(selector, e.target));
      }).
      triggerHandler(EVENT_ENABLE);
  });

  return QorAutoheight;

});
