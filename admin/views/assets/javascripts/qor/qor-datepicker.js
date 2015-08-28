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

  var NAMESPACE = 'qor.datepicker';
  var EVENT_ENABLE = 'enable.' + NAMESPACE;
  var EVENT_DISABLE = 'disable.' + NAMESPACE;
  var EVENT_CHANGE = 'change.' + NAMESPACE;
  var EVENT_CLICK = 'click.' + NAMESPACE;

  var CLASS_EMBEDDED = '.qor-datepicker__embedded';
  var CLASS_SAVE = '.qor-datepicker__save';

  function replaceText(str, data) {
    if (typeof str === 'string') {
      if (typeof data === 'object') {
        $.each(data, function (key, val) {
          str = str.replace('${' + String(key).toLowerCase() + '}', val);
        });
      }
    }

    return str;
  }

  function QorDatepicker(element, options) {
    this.$element = $(element);
    this.options = $.extend(true, {}, QorDatepicker.DEFAULTS, $.isPlainObject(options) && options);
    this.date = null;
    this.formatDate = null;
    this.built = false;
    this.init();
  }

  QorDatepicker.prototype = {
    init: function () {
      this.bind();
    },

    bind: function () {
      this.$element.on(EVENT_CLICK, $.proxy(this.show, this));
    },

    unbind: function () {
      this.$element.off(EVENT_CLICK, this.show);
    },

    build: function () {
      var $modal;

      if (this.built) {
        return;
      }

      this.$modal = $modal = $(replaceText(QorDatepicker.TEMPLATE, this.options.text)).appendTo('body');

      $modal.
        find(CLASS_EMBEDDED).
          on(EVENT_CHANGE, $.proxy(this.change, this)).
          datepicker({
            date: this.$element.val(),
            dateFormat: 'yyyy-mm-dd',
            inline: true,
          }).
          triggerHandler(EVENT_CHANGE);

      $modal.
        find(CLASS_SAVE).
          on(EVENT_CLICK, $.proxy(this.pick, this));

      this.built = true;
    },

    unbuild: function () {
      if (!this.built) {
        return;
      }

      this.$modal.
        find(CLASS_EMBEDDED).
          off(EVENT_CHANGE, this.change).
          datepicker('destroy').
          end().
        find(CLASS_SAVE).
          off(EVENT_CLICK, this.pick).
          end().
        remove();
    },

    change: function (e) {
      var $modal = this.$modal;
      var $target = $(e.target);
      var date;

      this.date = date = $target.datepicker('getDate');
      this.formatDate = $target.datepicker('getDate', true);

      $modal.find('.qor-datepicker__picked-year').text(date.getFullYear());
      $modal.find('.qor-datepicker__picked-date').text([
        $target.datepicker('getDayByNumber', date.getDay(), true) + ',',
        String($target.datepicker('getMonthByNumber', date.getMonth(), true)),
        date.getDate()
      ].join(' '));
    },

    show: function () {
      if (!this.built) {
        this.build();
      }

      this.$modal.qorModal('show');
    },

    pick: function () {
      this.$element.val(this.formatDate).closest('.mdl-js-textfield').trigger('update.qor.material');
      this.$modal.qorModal('hide');
    },

    destroy: function () {
      this.unbind();
      this.unbuild();
      this.$element.removeData(NAMESPACE);
    },
  };

  QorDatepicker.DEFAULTS = {
    text: {
      title: 'Pick a date',
      ok: 'OK',
      cancel: 'Cancel',
    }
  };

  QorDatepicker.TEMPLATE = (
     '<div class="qor-modal fade qor-datepicker" tabindex="-1" role="dialog" aria-hidden="true">' +
      '<div class="mdl-card mdl-shadow--2dp" role="document">' +
        '<div class="mdl-card__title">' +
          '<h2 class="mdl-card__title-text">${title}</h2>' +
        '</div>' +
        '<div class="mdl-card__supporting-text">' +
          '<div class="qor-datepicker__picked">' +
            '<div class="qor-datepicker__picked-year"></div>' +
            '<div class="qor-datepicker__picked-date"></div>' +
          '</div>' +
          '<div class="qor-datepicker__embedded"></div>' +
        '</div>' +
        '<div class="mdl-card__actions">' +
          '<a class="mdl-button mdl-button--colored mdl-js-button mdl-js-ripple-effect qor-datepicker__save">${ok}</a>' +
          '<a class="mdl-button mdl-button--colored mdl-js-button mdl-js-ripple-effect" data-dismiss="modal">${cancel}</a>' +
        '</div>' +
      '</div>' +
    '</div>'
  );

  QorDatepicker.plugin = function (option) {
    return this.each(function () {
      var $this = $(this);
      var data = $this.data(NAMESPACE);
      var options;
      var fn;

      if (!data) {
        if (!$.fn.datepicker) {
          return;
        }

        if (/destroy/.test(option)) {
          return;
        }

        options = $.extend(true, {}, $this.data(), typeof option === 'object' && option);
        $this.data(NAMESPACE, (data = new QorDatepicker(this, options)));
      }

      if (typeof option === 'string' && $.isFunction(fn = data[option])) {
        fn.apply(data);
      }
    });
  };

  $(function () {
    var selector = '[data-toggle="qor.datepicker"]';

    $(document).
      on(EVENT_DISABLE, function (e) {
        QorDatepicker.plugin.call($(selector, e.target), 'destroy');
      }).
      on(EVENT_ENABLE, function (e) {
        QorDatepicker.plugin.call($(selector, e.target));
      }).
      triggerHandler(EVENT_ENABLE);
  });

  return QorDatepicker;

});
