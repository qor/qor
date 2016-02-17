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
  var EVENT_CLICK = 'click.' + NAMESPACE;
  var EVENT_RESIZE = 'resize.' + NAMESPACE;
  var EVENT_SCROLL = 'scroll.' + NAMESPACE;
  var CLASS_IS_HIDDEN = 'is-hidden';
  var CLASS_IS_FIXED = 'is-fixed';

  function QorFixer(element, options) {
    this.$element = $(element);
    this.options = $.extend({}, QorFixer.DEFAULTS, $.isPlainObject(options) && options);
    this.$clone = null;
    this.init();
  }

  QorFixer.prototype = {
    constructor: QorFixer,

    init: function () {
      var options = this.options;
      var $this = this.$element;

      // disable fixer if have multiple tables
      if ($('.qor-page__body .qor-js-table').size() > 1) {
        return;
      }

      if ($this.is(':hidden') || $this.find('tbody > tr:visible').length <= 1) {
        return;
      }

      this.$thead = $this.find('thead:first');
      this.$tbody = $this.find('tbody:first');
      this.$header = $(options.header);
      this.$subHeader = $(options.subHeader);
      this.$content = $(options.content);
      this.marginBottomPX = parseInt(this.$subHeader.css('marginBottom'));
      this.paddingHeight = options.paddingHeight;

      this.resize();
      this.bind();
    },

    bind: function () {
      this.$element.on(EVENT_CLICK, $.proxy(this.check, this));

      this.$content.
        on(EVENT_SCROLL, $.proxy(this.toggle, this)).
        on(EVENT_RESIZE, $.proxy(this.resize, this));
    },

    unbind: function () {
      this.$element.off(EVENT_CLICK, this.check);

      this.$content.
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
        addClass([CLASS_IS_FIXED, CLASS_IS_HIDDEN].join(' ')).
        find('> tr').
          children().
            each(function (i) {
              $(this).width($items.eq(i).width());
            });
    },

    unbuild: function () {
      this.$clone.remove();
    },

    check: function (e) {
      var $target = $(e.target);
      var checked;

      if ($target.is('.qor-js-check-all')) {
        checked = $target.prop('checked');

        $target.
          closest('thead').
          siblings('thead').
            find('.qor-js-check-all').prop('checked', checked).
            closest('.mdl-checkbox').toggleClass('is-checked', checked);
      }
    },

    toggle: function () {
      var $this = this.$element;
      var $clone = this.$clone;
      var theadHeight = this.$thead.outerHeight();
      var tbodyLastRowHeight = this.$tbody.find('tr:last').outerHeight();
      var scrollTop = this.$content.scrollTop();
      var minTop = 0;
      var maxTop = $this.outerHeight() - theadHeight - tbodyLastRowHeight;
      var offsetTop = this.$subHeader.outerHeight() + this.paddingHeight + this.marginBottomPX;
      var headerHeight = $('.qor-page__header').outerHeight();
      var showTop = Math.min(scrollTop - offsetTop, maxTop) + headerHeight;

      if (scrollTop > offsetTop - headerHeight) {
        $clone.css('top', showTop).removeClass(CLASS_IS_HIDDEN);
      } else {
        $clone.css('top', minTop).addClass(CLASS_IS_HIDDEN);
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
    header: false,
    content: false,
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
    var selector = '.qor-js-table';
    var options = {
          header: '.mdl-layout__header',
          subHeader: '.qor-page__header',
          content: '.mdl-layout__content',
          paddingHeight: 2, // Fix sub header height bug
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
