(function (factory) {
  if (typeof define === 'function' && define.amd) {
    // AMD. Register as anonymous module.
    define('qor-datepicker', ['jquery'], factory);
  } else if (typeof exports === 'object') {
    // Node / CommonJS
    factory(require('jquery'));
  } else {
    // Browser globals.
    factory(jQuery);
  }
})(function ($) {

  'use strict';

  var QorDatepicker = function (element, options) {
        this.$element = $(element);
        this.options = $.extend({}, QorDatepicker.DEFAULTS, options);
        this.date = null;
        this.formatDate = null;
        this.built = false;
        this.init();
      };

  QorDatepicker.prototype = {
    init: function () {
      this.$element.on('click', $.proxy(this.show, this));

      if (this.options.show) {
        this.show();
      }
    },

    build: function () {
      var _this = this,
          $modal,
          $year,
          $month,
          $week,
          $day;

      if (this.built) {
        return;
      }

      this.$modal = $modal = $(QorDatepicker.TEMPLATE).appendTo('body');

      $year = $modal.find('.qor-datepicker-year');
      $month = $modal.find('.qor-datepicker-month');
      $week = $modal.find('.qor-datepicker-week');
      $day = $modal.find('.qor-datepicker-day');

      $modal.find('.qor-datepicker-embedded').on('change', function () {
        var $this = $(this),
            date;

        _this.date = date = $this.datepicker('getDate');
        _this.formatDate = $this.datepicker('getDate', true);
        $year.text(date.getFullYear());
        $month.text(String($this.datepicker('getMonthByNumber', date.getMonth(), true)).toUpperCase());
        $week.text($this.datepicker('getDayByNumber', date.getDay()));
        $day.text(date.getDate());
      }).datepicker({
        date: this.$element.val(),
        dateFormat: 'yyyy-mm-dd',
        inline: true
      }).triggerHandler('change');

      $modal.find('.qor-datepicker-save').on('click', $.proxy(this.pick, this));

      this.built = true;
    },

    show: function () {
      if (!this.built) {
        this.build();
      }

      this.$modal.modal('show');
    },

    pick: function () {
      this.$element.val(this.formatDate);
      this.$modal.modal('hide');
    }
  };

  QorDatepicker.DEFAULTS = {
    show: true
  };

  QorDatepicker.TEMPLATE = (
    '<div class="modal fade qor-datepicker-modal" id="qorDatepickerModal" tabindex="-1" role="dialog" aria-labelledby="qorDatepickerModalLabel" aria-hidden="true">' +
      '<div class="modal-dialog qor-datepicker">' +
        '<div class="modal-content">' +
          '<div class="modal-header sr-only">' +
            '<h5 class="modal-title" id="qorDatepickerModalLabel">Pick a date</h5>' +
          '</div>' +
          '<div class="modal-body">' +
            '<div class="qor-datepicker-picked">' +
              '<div class="qor-datepicker-week"></div>' +
              '<div class="qor-datepicker-month"></div>' +
              '<div class="qor-datepicker-day"></div>' +
              '<div class="qor-datepicker-year"></div>' +
            '</div>' +
            '<div class="qor-datepicker-embedded"></div>' +
          '</div>' +
          '<div class="modal-footer">' +
            '<button type="button" class="btn btn-link" data-dismiss="modal">Cancel</button>' +
            '<button type="button" class="btn btn-link qor-datepicker-save">OK</button>' +
          '</div>' +
        '</div>' +
      '</div>' +
    '</div>'
  );

  if (!$.fn.datepicker) {
    return;
  }

  $(document).on('click.qor.datepicker', '[data-toggle="qor.datepicker"]', function () {
    var $this = $(this),
        data = $this.data('qor.datepicker');

    if (!data) {
      $this.data('qor.datepicker', (data = new QorDatepicker(this, {
        show: false
      })));
    }

    data.show();
  });

  $(document).on('click.datepicker', '[data-toggle="datepicker"]', function () {
    var $this = $(this);

    if (!$this.data('datepicker')) {
      $this.datepicker({
        autoClose: true
      });
    }

    $this.datepicker('show');
  });

});
