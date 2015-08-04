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

  var $document = $(document);
  var NAMESPACE = 'qor.dropdown';
  var EVENT_ENABLE = 'enable.' + NAMESPACE;
  var EVENT_DISABLE = 'disable.' + NAMESPACE;
  var EVENT_CLICK = 'click.' + NAMESPACE;

  function QorDropdown(element, options) {
    this.$element = $(element);
    this.options = $.extend({}, QorDropdown.DEFAULTS, $.isPlainObject(options) && options);
    this.shown = false;
    this.init();
  }

  QorDropdown.prototype = {
    constructor: QorDropdown,

    init: function () {
      this.$parent = this.$element.closest(this.options.parent);
      this.bind();
    },

    bind: function () {
      this.$element.on(EVENT_CLICK, $.proxy(this.toggle, this));
      $document.on(EVENT_CLICK, $.proxy(this.click, this));
    },

    unbind: function () {
      this.$element.off(EVENT_CLICK, this.toggle);
      $document.off(EVENT_CLICK, this.click);
    },

    click: function (e) {
      if (!$(e.target).closest(this.options.parent).length) {
        this.hide();
      }
    },

    show: function () {
      this.shown = true;
      this.$parent.addClass(this.options.activeClass);
    },

    hide: function () {
      this.shown = false;
      this.$parent.removeClass(this.options.activeClass);
    },

    toggle: function () {
      if (this.shown) {
        this.hide();
      } else {
        this.show();
      }
    },

    destroy: function () {
      this.unbind();
      this.$element.removeData(NAMESPACE);
    },
  };

  QorDropdown.DEFAULTS = {
    parent: '.qor-dropdown',
    activeClass: 'open',
  };

  QorDropdown.plugin = function (options) {
    return this.each(function () {
      var $this = $(this);
      var data = $this.data(NAMESPACE);
      var fn;

      if (!data) {
        if (/destroy/.test(options)) {
          return;
        }

        $this.data(NAMESPACE, (data = new QorDropdown(this, options)));
      }

      if (typeof options === 'string' && $.isFunction(fn = data[options])) {
        fn.apply(data);
      }
    });
  };

  $(function () {
    var selector = '[data-toggle="qor.dropdown"]';

    $(document)
      .on(EVENT_CLICK, selector, function () {
        QorDropdown.plugin.call($(this));
      })
      .on(EVENT_DISABLE, function (e) {
        QorDropdown.plugin.call($(selector, e.target), 'destroy');
      })
      .on(EVENT_ENABLE, function (e) {
        QorDropdown.plugin.call($(selector, e.target));
      })
      .triggerHandler(EVENT_ENABLE);
  });

  return QorDropdown;

});
