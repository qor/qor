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

  /*var NAMESPACE = 'qor.l10n';

  function L10n(element, options) {
    this.$element = $(element);
    this.options = $.extend({}, L10n.DEFAULTS, $.isPlainObject(options) && options);
    this.init();
  }

  L10n.prototype = {
    contructor: L10n,

    init: function () {
      this.bind();
    },

    bind: function () {

    },

    unbind: function () {

    },

    destroy: function () {
      this.unbind();
      this.$element.removeData(NAMESPACE);
    }
  };

  L10n.DEFAULTS = {

  };

  L10n.plugin = function (options) {
    return this.each(function () {
      var $this = $(this),
          data = $this.data(NAMESPACE),
          fn;

      if (!data) {
        $this.data(NAMESPACE, (data = new L10n(this, options)));
      }

      if (typeof options === 'string' && $.isFunction((fn = data[options]))) {
        fn.apply(data);
      }
    });
  };*/

  $(function () {
    // L10n.plugin.call($('.qor-l10n'));
  });

});
