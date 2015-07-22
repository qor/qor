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

  var $window = $(window),
      NAMESPACE = 'qor.fixer',
      EVENT_ENABLE = 'enable.' + NAMESPACE,
      EVENT_DISABLE = 'disable.' + NAMESPACE,
      EVENT_RESIZE = 'resize.' + NAMESPACE,
      EVENT_SCROLL = 'scroll.' + NAMESPACE;

  function QorFixer(element, options) {
    this.$element = $(element);
    this.options = $.extend({}, QorFixer.DEFAULTS, $.isPlainObject(options) && options);
    this.$clone = null;
    this.init();
  }

  QorFixer.prototype = {
    constructor: QorFixer,

    init: function () {
      var $this = this.$element;

      if ($this.is(':hidden') || $this.find('tbody:visible > tr:visible').length <= 1) {
        return;
      }

      this.$thead = $this.find('thead:first');
      this.$tbody = $this.find('tbody:first');
      this.$tfoot = $this.find('tfoot:first');

      this.resize();
      this.bind();
    },

    bind: function () {
      $window
        .on(EVENT_SCROLL, $.proxy(this.toggle, this))
        .on(EVENT_RESIZE, $.proxy(this.resize, this));
    },

    unbind: function () {
      $window
        .off(EVENT_SCROLL, this.toggle)
        .off(EVENT_RESIZE, this.resize);
    },

    build: function () {
      var $this = this.$element,
          $thead = this.$thead,
          $tbody = this.$tbody,
          $tfoot = this.$tfoot,
          $clone = this.$clone,
          $items = $thead.find('> tr').children();

      this.offsetTop = $this.offset().top;
      this.maxTop = $this.outerHeight() - $thead.height() - $tbody.find('> tr:last').height() - $tfoot.height();

      if (!$clone) {
        this.$clone = $clone = $thead.clone().prependTo($this);
      }

      $clone.
        css({
          position: 'fixed',
          top: 0,
          zIndex: 1,
          display: 'none',
          width: $thead.width()
        }).
        find('> tr').
          children().
            each(function (i) {
              $(this).width($items.eq(i).width());
            });
    },

    unbuild: function () {
      this.$clone.remove();
    },

    toggle: function () {
      var $clone = this.$clone,
        top = $window.scrollTop() - this.offsetTop;

      if (top > 0 && top < this.maxTop) {
        $clone.show();
      } else {
        $clone.hide();
      }
    },

    resize: function () {
      this.build();
      this.toggle();
    },

    destroy: function () {
      this.unbind();
      this.unbuild();
      this.$element.removeData(NAMESPACE);
    }
  };

  QorFixer.DEFAULTS = {};

  QorFixer.plugin = function (options) {
    return this.each(function () {
      var $this = $(this),
          data = $this.data(NAMESPACE),
          fn;

      if (!data) {
        $this.data(NAMESPACE, (data = new QorFixer(this, options)));
      }

      if (typeof options === 'string' && $.isFunction(fn = data[options])) {
        fn.call(data);
      }
    });
  };

  $(function () {
    var selector = '.qor-list';

    $(document)
      .on(EVENT_DISABLE, function (e) {
        QorFixer.plugin.call($(selector, e.target), 'destroy');
      })
      .on(EVENT_ENABLE, function (e) {
        QorFixer.plugin.call($(selector, e.target));
      })
      .triggerHandler(EVENT_ENABLE);
  });

  return QorFixer;

});
