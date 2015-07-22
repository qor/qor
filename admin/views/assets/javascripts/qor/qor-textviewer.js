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
      EVENT_ENABLE = 'enable.' + NAMESPACE,
      EVENT_DISABLE = 'disable.' + NAMESPACE,
      EVENT_CLICK = 'click.' + NAMESPACE;

  function QorTextviewer(element, options) {
    this.$element = $(element);
    this.options = $.extend({}, QorTextviewer.DEFAULTS, $.isPlainObject(options) && options);
    this.$modal = null;
    this.built = false;
    this.init();
  }

  QorTextviewer.prototype = {
    constructor: QorTextviewer,

    init: function () {
      this.$element.find(this.options.toggle).each(function () {
        var $this = $(this);

        // 8 for correction as `scrollHeight` will large than `offsetHeight` most of the time.
        if (this.scrollHeight > this.offsetHeight + 8) {
          $this.addClass('active').wrapInner(QorTextviewer.INNER);
        }
      });

      this.bind();
    },

    bind: function () {
      this.$element.on(EVENT_CLICK, this.options.toggle, $.proxy(this.click, this));
    },

    unbind: function () {
      this.$element.off(EVENT_CLICK, this.click);
    },

    build: function () {
      if (this.built) {
        return;
      }

      this.built = true;
      this.$modal = $(QorTextviewer.TEMPLATE).modal({
        show: false
      }).appendTo('body');
    },

    unbuild: function () {
      if (!this.built) {
        return;
      }

      this.built = false;
      this.$modal.remove();
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
        $modal.find('.modal-title').text($target.closest('td').data('heading'));
        $modal.find('.modal-body').html($target.find('.text-inner').html());
        $modal.modal('show');
      }
    },

    destroy: function () {
      this.$element.find(this.options.toggle).find('.text-inner').each(function () {
        var $this = $(this);

        $this.before($this.html()).remove();
      });

      this.unbind();
      this.unbuild();
      this.$element.removeData(NAMESPACE);
    }
  };

  QorTextviewer.DEFAULTS = {
    toggle: false
  };

  QorTextviewer.INNER = ('<div class="text-inner"></div>');

  QorTextviewer.TEMPLATE = (
    '<div class="modal fade qor-list-modal" id="qorListModal" tabindex="-1" role="dialog" aria-labelledby="qorListModalLabel" aria-hidden="true">' +
      '<div class="modal-dialog">' +
        '<div class="modal-content">' +
          '<div class="modal-header">' +
            '<button type="button" class="close" data-dismiss="modal" aria-label="Close"><span aria-hidden="true">&times;</span></button>' +
            '<h4 class="modal-title" id="qorListModalLabel">&nbsp;</h4>' +
          '</div>' +
          '<div class="modal-body"></div>' +
        '</div>' +
      '</div>' +
    '</div>'
  );

  QorTextviewer.plugin = function (options) {
    return this.each(function () {
      var $this = $(this),
          data = $this.data(NAMESPACE),
          fn;

      if (!data) {
        if (!$.fn.modal) {
          return;
        }

        if (/destroy/.test(options)) {
          return;
        }

        $this.data(NAMESPACE, (data = new QorTextviewer(this, options)));
      }

      if (typeof options === 'string' && $.isFunction(fn = data[options])) {
        fn.apply(data);
      }
    });
  };

  $(function () {
    var selector = '.qor-list',
        options = {
          toggle: '.qor-list-text'
        };

    $(document)
      .on(EVENT_DISABLE, function (e) {
        QorTextviewer.plugin.call($(selector, e.target), 'destroy');
      })
      .on(EVENT_ENABLE, function (e) {
        QorTextviewer.plugin.call($(selector, e.target), options);
      })
      .triggerHandler(EVENT_ENABLE);
  });

});
