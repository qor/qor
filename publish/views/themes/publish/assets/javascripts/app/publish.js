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

  var NAMESPACE = 'qor.publish',
      EVENT_CLICK = 'click.' + NAMESPACE;

  function Publish(element, options) {
    this.$element = $(element);
    this.options = $.extend({}, Publish.DEFAULTS, $.isPlainObject(options) && options);
    this.loading = false;
    this.init();
  }

  Publish.prototype = {
    constructor: Publish,

    init: function () {
      var options = this.options,
          $this = this.$element;

      this.$modal = $this.find(options.modal);

      if ($.fn.tooltip) {
        $this.find(options.toggleCheck).tooltip();
      }

      this.bind();
    },

    bind: function () {
      this.$element.on(EVENT_CLICK, $.proxy(this.click, this));
    },

    unbind: function () {
      this.$element.off(EVENT_CLICK, this.click);
    },

    click: function (e) {
      var options = this.options,
          $target = $(e.target);

      if ($target.is(options.toggleView)) {
        e.preventDefault();

        if (this.loading) {
          return;
        }

        this.loading = true;
        this.$modal.find('.modal-body').empty().load($target.data('url'), $.proxy(this.show, this));
      } else if ($target.is(options.toggleCheck)) {
        if (!$target.prop('disabled')) {
          $target.closest('table').find(':checkbox').not($target).prop('checked', $target.prop('checked'));
        }
      }
    },

    show: function () {
      this.loading = false;
      this.$modal.modal('show');
    },

    destroy: function () {
      this.unbind();
      this.$element.removeData(NAMESPACE);
    }
  };

  Publish.DEFAULTS = {
    modal: '.qor-publish-modal',
    toggleView: '.qor-action-diff',
    toggleCheck: '.qor-check-all'
  };

  Publish.plugin = function (options) {
    return this.each(function () {
      var $this = $(this),
          data = $this.data(NAMESPACE),
          fn;

      if (!data) {
        if (!$.fn.modal) {
          return;
        }

        $this.data(NAMESPACE, (data = new Publish(this, options)));
      }

      if (typeof options === 'string' && $.isFunction((fn = data[options]))) {
        fn.apply(data);
      }
    });
  };

  $(function () {
    Publish.plugin.call($('.qor-publish'));
  });

});
