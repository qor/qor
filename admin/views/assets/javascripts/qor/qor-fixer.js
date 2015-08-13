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

  var NAMESPACE = 'qor.fixer';
  var EVENT_ENABLE = 'enable.' + NAMESPACE;
  var EVENT_DISABLE = 'disable.' + NAMESPACE;
  var EVENT_RESIZE = 'resize.' + NAMESPACE;
  var EVENT_SCROLL = 'scroll.' + NAMESPACE;

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

      if ($this.is(':hidden') || $this.find('tbody > tr:visible').length <= 1) {
        return;
      }

      this.$thead = $this.find('thead:first');
      this.$tbody = $this.find('tbody:first');
      this.$context = $(this.options.context);

      this.resize();
      this.bind();
    },

    bind: function () {
      this.$context.
        on(EVENT_SCROLL, $.proxy(this.toggle, this)).
        on(EVENT_RESIZE, $.proxy(this.resize, this));
    },

    unbind: function () {
      this.$context.
        off(EVENT_SCROLL, this.toggle).
        off(EVENT_RESIZE, this.resize);
    },

    build: function () {
      var $this = this.$element;
      var $thead = this.$thead;
      var $clone = this.$clone;
      var $items = $thead.find('> tr').children();

      if (!$clone) {
        this.$clone = $clone = $thead.clone().prependTo($this);
      }

      $clone.
        css({
          position: 'fixed',
          top: 64,
          zIndex: 1,
          display: 'none',
          width: $thead.width(),
          backgroundColor: '#fff',
          borderBottom: '1px solid rgba(0,0,0,0.12)',
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
      var $this = this.$element;
      var $clone = this.$clone;
      var offset = $this.offset();
      var top = this.$context.scrollTop() - offset.top;
      var theadHeight = this.$thead.height();
      var tbodyHeight = this.$tbody.find('> tr:last').height();
      var maxTop = $this.outerHeight() - theadHeight - tbodyHeight;

      if (top > 0 && top < maxTop) {
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
    },
  };

  QorFixer.DEFAULTS = {
    context: window
  };

  QorFixer.plugin = function (options) {
    return this.each(function () {
      var $this = $(this);
      var data = $this.data(NAMESPACE);
      var fn;

      if (!data) {
        $this.data(NAMESPACE, (data = new QorFixer(this, options)));
      }

      if (typeof options === 'string' && $.isFunction(fn = data[options])) {
        fn.call(data);
      }
    });
  };

  $(function () {
    var selector = '.qor-table';
    var options = {
          context: '.mdl-layout__content',
        };

    $(document).
      on(EVENT_DISABLE, function (e) {
        QorFixer.plugin.call($(selector, e.target), 'destroy');
      }).
      on(EVENT_ENABLE, function (e) {
        QorFixer.plugin.call($(selector, e.target), options);
      }).
      triggerHandler(EVENT_ENABLE);
  });

  return QorFixer;

});
