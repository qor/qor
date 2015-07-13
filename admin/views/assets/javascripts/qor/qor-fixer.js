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
      EVENT_SCROLL = 'scroll.' + NAMESPACE;

  function QorFixer(element, options) {
    this.$element = $(element);
    this.options = $.extend({}, QorFixer.DEFAULTS, $.isPlainObject(options) && options);
    this.init();
  }

  QorFixer.prototype = {
    constructor: QorFixer,

    init: function () {
      var $this = this.$element,
          $container = $this.parent();

      if ($this.is(':hidden') || $this.find('tbody:visible > tr:visible').length <= 1) {
        return;
      }

      if ($container.css('position') === 'static') {
        $container.css('position', 'relative');
      }

      this.maxTop = ($this.outerHeight() - $this.find('thead').height() - $this.find('tbody:last > tr:last').height() - $this.find('tfoot').height());

      this.clone();
      this.place();
      this.bind();
    },

    clone: function () {
      var $this = this.$element,
          $clone = $this.clone(),
          $ths = $clone.find('thead > tr > th');

      $this.find('thead > tr > th').each(function (i) {
        // $ths.eq(i).width($(this).outerWidth());
        $ths.eq(i).prepend($('<div>').css({
          height: 0,
          width: $(this).width(),
          overflow: 'hidden'
        }));
      });

      $clone.find('tbody, tfoot').remove();

      $clone.css({
        position: 'absolute',
        top: 0,
        left: 0,
        width: $this.outerWidth()
      });

      this.$clone = $clone.insertAfter($this);
    },

    bind: function () {
      var _this = this;

      $window.on(EVENT_SCROLL, (this._place = function () {
        _this.place();
      }));
    },

    unbind: function () {
      $window.off(EVENT_SCROLL, this._place);
    },

    place: function () {
      var top = $window.scrollTop() - this.$element.offset().top,
          maxTop = this.maxTop;

      top = top < 0 ? 0 : (top > maxTop ? maxTop : top);
      this.$clone.css('top', top);
    },

    destroy: function () {
      this.unbind();
      this.$clone.remove();
      this.$element.removeData(NAMESPACE);
    }
  };

  QorFixer.DEFAULTS = {
  };

  QorFixer.plugin = function (options) {
    return this.each(function () {
      var $this = $(this),
          data = $this.data(NAMESPACE),
          fn;

      if (!data) {
        $this.data(NAMESPACE, (data = new QorFixer(this, options)));
      }

      if (typeof options === 'string' && $.isFunction(fn = data[options])) {
        fn.call(data, options);
      }
    });
  };

  $(function () {
    QorFixer.plugin.call($('.qor-list'));
  });

});
