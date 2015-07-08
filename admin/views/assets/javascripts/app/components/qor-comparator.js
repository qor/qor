(function (factory) {
  if (typeof define === 'function' && define.amd) {
    // AMD. Register as anonymous module.
    define('qor-comparator', ['jquery'], factory);
  } else if (typeof exports === 'object') {
    // Node / CommonJS
    factory(require('jquery'));
  } else {
    // Browser globals.
    factory(jQuery);
  }
})(function ($) {

  'use strict';

  var QorComparator = function (element, options) {
        this.$element = $(element);
        this.options = $.extend({}, QorComparator.DEFAULTS, options);
        this.init();
      };

  QorComparator.prototype = {
    constructor: QorComparator,

    init: function () {
      if (!$.fn.modal) {
        return;
      }

      this.$modal = $(QorComparator.TEMPLATE.replace(/\{\{key\}\}/g, Date.now())).appendTo('body');
      this.$modal.modal(this.options);
    },

    show: function () {
      this.$modal.modal('show');
    }
  };

  QorComparator.DEFAULTS = {
    keyboard: true,
    backdrop: true,
    remote: false,
    show: true
  };

  QorComparator.TEMPLATE = (
    '<div class="modal fade qor-comparator-modal" id="qorComparatorModal{{key}}" aria-labelledby="qorComparatorModalLabel{{key}}" tabindex="-1" role="dialog" aria-hidden="true">' +
      '<div class="modal-dialog">' +
        '<div class="modal-content">' +
          '<div class="modal-header">' +
            '<h5 class="modal-title" id="qorComparatorModalLabel{{key}}">Diff</h5>' +
          '</div>' +
          '<div class="modal-body"></div>' +
        '</div>' +
      '</div>' +
    '</div>'
  );

  QorComparator.plugin = function (options) {
    var args = [].slice.call(arguments, 1),
        result;

    this.each(function () {
      var $this = $(this),
          data = $this.data('qor.comparator'),
          fn;

      if (!data) {
        $this.data('qor.comparator', (data = new QorComparator(this, options)));
      } else {
        options = 'show';
      }

      if (typeof options === 'string' && $.isFunction((fn = data[options]))) {
        result = fn.apply(data, args);
      }
    });

    return typeof result === 'undefined' ? this : result;
  };

  $(function () {
    if (!$.fn.modal) {
      return;
    }

    $(document)
      .on('click.qor.comparator.initiator', '[data-toggle="qor.comparator"]', function (e) {
        var $this = $(this);

        e.preventDefault();
        QorComparator.plugin.call($this, $this.data());
      });
  });

  return QorComparator;

});
