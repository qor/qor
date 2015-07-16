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

  var NAMESPACE = 'qor.textviewer',
      EVENT_CLICK = 'click.' + NAMESPACE;

  function TextViewer(element, options) {
    this.$element = $(element);
    this.options = $.extend({}, TextViewer.DEFAULTS, $.isPlainObject(options) && options);
    this.$modal = null;
    this.built = false;
    this.init();
  }

  TextViewer.prototype = {
    constructor: TextViewer,

    init: function () {
      this.$element.find(this.options.toggle).each(function () {
        var $this = $(this);

        // 8 for correction as `scrollHeight` will large than `offsetHeight` most of the time.
        if (this.scrollHeight > this.offsetHeight + 8) {
          $this.addClass('active').wrapInner(TextViewer.INNER);
        }
      });
      this.bind();
    },

    build: function () {
      if (this.built) {
        return;
      }

      this.built = true;
      this.$modal = $(TextViewer.TEMPLATE).modal({
        show: false
      }).appendTo('body');
    },

    bind: function () {
      this.$element.on(EVENT_CLICK, this.options.toggle, $.proxy(this.click, this));
    },

    unbind: function () {
      this.$element.off(EVENT_CLICK, this.click);
    },

    click: function (e) {
      var target = e.currentTarget,
          $target = $(target),
          $modal;

      if (!this.built) {
        this.build();
      }

      if ($target.hasClass('active')) {
        $modal = this.$modal;
        $modal.find('.modal-title').text($target.closest('td').attr('title'));
        $modal.find('.modal-body').html($target.find('.text-inner').html());
        $modal.modal('show');
      }
    },

    destroy: function () {
      this.unbind();
      this.$element.removeData(NAMESPACE);
    }
  };

  TextViewer.DEFAULTS = {
    toggle: '.qor-list-text'
  };

  TextViewer.INNER = ('<div class="text-inner"></div>');

  TextViewer.TEMPLATE = (
    '<div class="modal fade qor-list-modal" id="qorListModal" tabindex="-1" role="dialog" aria-labelledby="qorListModalLabel" aria-hidden="true">' +
      '<div class="modal-dialog">' +
        '<div class="modal-content">' +
          '<div class="modal-header">' +
            '<button type="button" class="close" data-dismiss="modal" aria-label="Close"><span aria-hidden="true">&times;</span></button>' +
            '<h4 class="modal-title" id="qorPublishModalLabel"></h4>' +
          '</div>' +
          '<div class="modal-body"></div>' +
        '</div>' +
      '</div>' +
    '</div>'
  );

  TextViewer.plugin = function (options) {
    return this.each(function () {
      var $this = $(this),
          data = $this.data(NAMESPACE),
          fn;

      if (!data) {
        if (!$.fn.modal) {
          return;
        }

        $this.data(NAMESPACE, (data = new TextViewer(this, options)));
      }

      if (typeof options === 'string' && $.isFunction((fn = data[options]))) {
        fn.apply(data);
      }
    });
  };

  $(function () {
    TextViewer.plugin.call($('.qor-list'));
  });

});
