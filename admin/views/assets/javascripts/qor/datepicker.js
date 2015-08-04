(function (factory) {
  if (typeof define === 'function' && define.amd) {
    // AMD. Register as anonymous module.
    define('datepicker', ['jquery'], factory);
  } else if (typeof exports === 'object') {
    // Node / CommonJS
    factory(require('jquery'));
  } else {
    // Browser globals.
    factory(jQuery);
  }
})(function ($) {

  'use strict';

  var $window = $(window);
  var $document = $(document);
  var NAMESPACE = 'datepicker';
  var EVENT_CLICK = 'click.' + NAMESPACE;
  var EVENT_KEYUP = 'keyup.' + NAMESPACE;
  var EVENT_FOCUS = 'focus.' + NAMESPACE;
  var EVENT_RESIZE = 'resize.' + NAMESPACE;

  function Datepicker(element, options) {
    this.$element = $(element);
    this.options = $.extend({}, Datepicker.DEFAULTS, $.isPlainObject(options) && options);
    this.visible = false;
    this.isInput = false;
    this.isInline = false;
    this.init();
  }

  function isNumber(n) {
    return typeof n === 'number';
  }

  function isLeapYear (year) {
    return (year % 4 === 0 && year % 100 !== 0) || year % 400 === 0;
  }

  function getDaysInMonth (year, month) {
    return [31, (isLeapYear(year) ? 29 : 28), 31, 30, 31, 30, 31, 31, 30, 31, 30, 31][month];
  }

  function parseFormat (format) {
    var separator = format.match(/[.\/\-\s].*?/) || '/';
    var parts = format.split(/\W+/);
    var length;
    var i;

    if (!parts || parts.length === 0) {
      throw new Error('Invalid date format.');
    }

    format = {
      separator: separator[0],
      parts: parts
    };

    length = parts.length;

    for (i = 0; i < length; i++) {
      switch (parts[i]) {
        case 'dd':
        case 'd':
          format.day = true;
          break;

        case 'mm':
        case 'm':
          format.month = true;
          break;

        case 'yyyy':
        case 'yy':
          format.year = true;
          break;

        // No default
      }
    }

    return format;
  }

  function parseDate (date, format) {
    var parts;
    var length;
    var year;
    var day;
    var month;
    var val;
    var i;

    parts = typeof date === 'string' && date ? date.split(format.separator) : [];
    length = format.parts.length;

    date = new Date();
    year = date.getFullYear();
    day = date.getDate();
    month = date.getMonth();

    if (parts.length === length) {
      for (i = 0; i < length; i++) {
        val = parseInt(parts[i], 10) || 1;

        switch (format.parts[i]) {
          case 'dd':
          case 'd':
            day = val;
            break;

          case 'mm':
          case 'm':
            month = val - 1;
            break;

          case 'yy':
            year = 2000 + val;
            break;

          case 'yyyy':
            year = val;
            break;

          // No default
        }
      }
    }

    return new Date(year, month, day, 0, 0, 0, 0);
  }

  function formatDate (date, format) {
    var val = {
          d: date.getDate(),
          m: date.getMonth() + 1,
          yy: date.getFullYear().toString().substring(2),
          yyyy: date.getFullYear()
        };
    var parts = [];
    var length = format.parts.length;
    var i;

    val.dd = (val.d < 10 ? '0' : '') + val.d;
    val.mm = (val.m < 10 ? '0' : '') + val.m;

    for (i = 0; i < length; i++) {
      parts.push(val[format.parts[i]]);
    }

    return parts.join(format.separator);
  }

  Datepicker.prototype = {
    constructor: Datepicker,

    init: function () {
      var $this = this.$element;
      var options = this.options;
      var $picker;

      this.$trigger = $(options.trigger || $this);
      this.$picker = $picker = $(options.template);
      this.$years = $picker.find('[data-type="years picker"]');
      this.$months = $picker.find('[data-type="months picker"]');
      this.$days = $picker.find('[data-type="days picker"]');
      this.isInput = $this.is('input') || $this.is('textarea');
      this.isInline = options.inline && (options.container || !this.isInput);

      if (this.isInline) {
        $picker.find('.datepicker-arrow').hide();
        $(options.container || $this).append($picker);
      } else {
        $(options.container || 'body').append($picker);
        this.place();
        $picker.hide();
      }

      if (options.date) {
        $this.data('date', options.date);
      }

      this.format = parseFormat(options.dateFormat);
      this.fillWeek();
      this.bind();
      this.update();

      if (this.isInline) {
        this.show();
      }
    },

    bind: function () {
      var $this = this.$element;
      var options = this.options;

      this.$picker.on(EVENT_CLICK, $.proxy(this.click, this));

      if (!this.isInline) {
        if (this.isInput) {
          $this.on(EVENT_KEYUP, $.proxy(this.update, this));

          if (!options.trigger) {
            $this.on(EVENT_FOCUS, $.proxy(this.show, this));
          }
        }

        this.$trigger.on(EVENT_CLICK, $.proxy(this.show, this));
      }
    },

    unbind: function () {
      var $this = this.$element;
      var options = this.options;

      this.$picker.off(EVENT_CLICK, this.click);

      if (!this.isInline) {
        if (this.isInput) {
          $this.off(EVENT_KEYUP, this.update);

          if (!options.trigger) {
            $this.off(EVENT_FOCUS, this.show);
          }
        }

        this.$trigger.off(EVENT_CLICK, this.show);
      }
    },

    showView: function (type) {
      var format = this.format;

      if (format.year || format.month || format.day) {
        switch (type) {

          case 2:
          case 'years':
            this.$months.hide();
            this.$days.hide();

            if (format.year) {
              this.fillYears();
              this.$years.show();
            } else {
              this.showView(0);
            }

            break;

          case 1:
          case 'months':
            this.$years.hide();
            this.$days.hide();

            if (format.month) {
              this.fillMonths();
              this.$months.show();
            } else {
              this.showView(2);
            }

            break;

          // case 0:
          // case 'days':
          default:
            this.$years.hide();
            this.$months.hide();

            if (format.day) {
              this.fillDays();
              this.$days.show();
            } else {
              this.showView(1);
            }
        }
      }
    },

    hideView: function () {
      if (this.options.autoClose) {
        this.hide();
      }
    },

    place: function () {
      var $trigger = this.$trigger;
      var offset = $trigger.offset();

      this.$picker.css({
        position: 'absolute',
        top: offset.top + $trigger.outerHeight(),
        left: offset.left,
        zIndex: this.options.zIndex
      });
    },

    show: function () {
      if (this.visible) {
        return;
      }

      this.visible = true;
      this.$picker.show();

      if (!this.isInline) {
        $window.on(EVENT_RESIZE, $.proxy(this.place, this));
        $document.on(EVENT_CLICK, $.proxy(function (e) {
          if (e.target !== this.$element[0]) {
            this.hide();
          }
        }, this));
      }

      this.showView(this.options.viewStart);
    },

    hide: function () {
      if (!this.visible) {
        return;
      }

      this.visible = false;
      this.$picker.hide();

      if (!this.isInline) {
        $window.off(EVENT_RESIZE, this.place);
        $document.off(EVENT_CLICK, this.hide);
      }
    },

    update: function () {
      var $this = this.$element;
      var date = $this.data('date') || (this.isInput ? $this.prop('value') : $this.text());

      this.date = date = parseDate(date, this.format);
      this.viewDate = new Date(date.getFullYear(), date.getMonth(), date.getDate());
      this.fillAll();
    },

    change: function () {
      var $this = this.$element;
      var date = formatDate(this.date, this.format);

      if (this.isInput) {
        $this.prop('value', date);
      } else if (!this.isInline) {
        $this.text(date);
      }

      $this.data('date', date).trigger('change');
    },

    getMonthByNumber: function (month, short) {
      var options = this.options;
      var months = short ? options.monthsShort : options.months;

      return months[isNumber(month) ? month : this.date.getMonth()];
    },

    getDayByNumber: function (day, short, min) {
      var options = this.options;
      var days = min ? options.daysMin : short ? options.daysShort : options.days;

      return days[isNumber(day) ? day : this.date.getDay()];
    },

    getDate: function (format) {
      return format ? formatDate(this.date, this.format) : new Date(this.date);
    },

    template: function (data) {
      var options = this.options;
      var defaults = {
            text: '',
            type: '',
            selected: false,
            disabled: false
          };

      $.extend(defaults, data);

      return [
        '<' + options.itemTag + ' ',
        (defaults.selected ? 'class="' + options.selectedClass + '"' :
        defaults.disabled ? 'class="' + options.disabledClass + '"' : ''),
        (defaults.type ? ' data-type="' + defaults.type + '"' : ''),
        '>',
        defaults.text,
        '</' + options.itemTag + '>'
      ].join('');
    },

    fillAll: function () {
      this.fillYears();
      this.fillMonths();
      this.fillDays();
    },

    fillYears: function () {
      var title = '';
      var items = [];
      var suffix = this.options.yearSuffix || '';
      var year = this.date.getFullYear();
      var viewYear = this.viewDate.getFullYear();
      var isCurrent;
      var i;

      title = (viewYear - 5) + suffix + ' - ' + (viewYear + 6) + suffix;

      for (i = -5; i < 7; i++) {
        isCurrent = (viewYear + i) === year;
        items.push(this.template({
          text: viewYear + i,
          type: isCurrent ? 'year selected' : 'year',
          selected: isCurrent,
          disabled: i === -5 || i === 6
        }));
      }

      this.$picker.find('[data-type="years current"]').html(title);
      this.$picker.find('[data-type="years"]').empty().html(items.join(''));
    },

    fillMonths: function () {
      var title = '';
      var items = [];
      var options = this.options.monthsShort;
      var year = this.date.getFullYear();
      var month = this.date.getMonth();
      var viewYear = this.viewDate.getFullYear();
      var isCurrent;
      var i;

      title = viewYear.toString() + this.options.yearSuffix || '';

      for (i = 0; i < 12; i++) {
        isCurrent = viewYear === year && i === month;

        items.push(this.template({
          text: options[i],
          type: isCurrent ? 'month selected' : 'month',
          selected: isCurrent
        }));
      }

      this.$picker.find('[data-type="year current"]').html(title);
      this.$picker.find('[data-type="months"]').empty().html(items.join(''));
    },

    fillWeek: function () {
      var options = this.options;
      var items = [];
      var days = options.daysMin;
      var weekStart = parseInt(options.weekStart, 10) % 7;
      var i;

      days = $.merge(days.slice(weekStart), days.slice(0, weekStart));

      for (i = 0; i < 7; i++) {
        items.push(this.template({
          text: days[i]
        }));
      }

      this.$picker.find('[data-type="week"]').html(items.join(''));
    },

    fillDays: function () {
      var title = '';
      var items = [];
      var prevItems = [];
      var currentItems = [];
      var nextItems = [];
      var options = this.options.monthsShort;
      var suffix = this.options.yearSuffix || '';
      var year = this.date.getFullYear();
      var month = this.date.getMonth();
      var day = this.date.getDate();
      var viewYear = this.viewDate.getFullYear();
      var viewMonth = this.viewDate.getMonth();
      var weekStart = parseInt(this.options.weekStart, 10) % 7;
      var isCurrent;
      var isDisabled;
      var length;
      var date;
      var i;
      var n;

      // Title of current month
      title = this.options.showMonthAfterYear ? (viewYear + suffix + ' ' + options[viewMonth]) : options[viewMonth] + ' ' + viewYear + suffix;

      // Days of prev month
      length = viewMonth === 0 ? getDaysInMonth(viewYear - 1, 11) : getDaysInMonth(viewYear, viewMonth - 1);

      for (i = 1; i <= length; i++) {
        prevItems.push(this.template({
          text: i,
          type: 'day prev',
          disabled: true
        }));
      }

      date = new Date(viewYear, viewMonth, 1, 0, 0, 0, 0); // The first day of current month
      n = (7 + (date.getDay() - weekStart)) % 7;
      n = n > 0 ? n : 7;
      prevItems = prevItems.slice((length - n));

      // Days of prev month next
      length = viewMonth === 11 ? getDaysInMonth(viewYear + 1, 0) : getDaysInMonth(viewYear, viewMonth + 1);

      for (i = 1; i <= length; i++) {
        nextItems.push(this.template({
          text: i,
          type: 'day next',
          disabled: true
        }));
      }

      length = getDaysInMonth(viewYear, viewMonth);
      date = new Date(viewYear, viewMonth, length, 0, 0, 0, 0); // The last day of current month
      n = (7 - (date.getDay() + 1 - weekStart)) % 7;
      n = n >= (7 * 6 - (prevItems.length + length)) ? n : n + 7; // 7 * 6 : 7 columns & 6 rows, 42 items
      nextItems = nextItems.slice(0, n);

      // Days of current month
      for (i = 1; i <= length; i++) {
        isCurrent = viewYear === year && viewMonth === month && i === day;
        isDisabled = this.options.isDisabled(new Date(viewYear, viewMonth, i));

        currentItems.push(this.template({
          text: i,
          type: isDisabled ? 'day disabled' : isCurrent ? 'day selected' : 'day',
          selected: isCurrent,
          disabled: isDisabled
        }));
      }

      // Merge all the days
      $.merge(items, prevItems);
      $.merge(items, currentItems);
      $.merge(items, nextItems);

      this.$picker.find('[data-type="month current"]').html(title);
      this.$picker.find('[data-type="days"]').empty().html(items.join(''));
    },

    click: function (e) {
      var $target = $(e.target);
      var yearRegex = /^\d{2,4}$/;
      var isYear = false;
      var viewYear;
      var viewMonth;
      var viewDay;
      var year;
      var type;

      e.stopPropagation();
      e.preventDefault();

      if ($target.length === 0) {
        return;
      }

      viewYear = this.viewDate.getFullYear();
      viewMonth = this.viewDate.getMonth();
      viewDay = this.viewDate.getDate();
      type = $target.data().type;

      switch (type) {
        case 'years prev':
        case 'years next':
          viewYear = type === 'years prev' ? viewYear - 10 : viewYear + 10;
          year = $target.text();
          isYear = yearRegex.test(year);

          if (isYear) {
            viewYear = parseInt(year, 10);
            this.date = new Date(viewYear, viewMonth, Math.min(viewDay, 28), 0, 0, 0, 0);
          }

          this.viewDate = new Date(viewYear, viewMonth, Math.min(viewDay, 28), 0, 0, 0, 0);
          this.fillYears();

          if (isYear) {
            this.showView(1);
            this.change();
          }

          break;

        case 'year prev':
        case 'year next':
          viewYear = type === 'year prev' ? viewYear - 1 : viewYear + 1;
          this.viewDate = new Date(viewYear, viewMonth, Math.min(viewDay, 28), 0, 0, 0, 0);
          this.fillMonths();
          break;

        case 'year current':

          if (this.format.year) {
            this.showView(2);
          }

          break;

        case 'year selected':

          if (this.format.month) {
            this.showView(1);
          } else {
            this.hideView();
          }

          break;

        case 'year':
          viewYear = parseInt($target.text(), 10);
          this.date = new Date(viewYear, viewMonth, Math.min(viewDay, 28), 0, 0, 0, 0);
          this.viewDate = new Date(viewYear, viewMonth, Math.min(viewDay, 28), 0, 0, 0, 0);

          if (this.format.month) {
            this.showView(1);
          } else {
            this.hideView();
          }

          this.change();
          break;

        case 'month prev':
        case 'month next':
          viewMonth = type === 'month prev' ? viewMonth - 1 : type === 'month next' ? viewMonth + 1 : viewMonth;
          this.viewDate = new Date(viewYear, viewMonth, Math.min(viewDay, 28), 0, 0, 0, 0);
          this.fillDays();
          break;

        case 'month current':

          if (this.format.month) {
            this.showView(1);
          }

          break;

        case 'month selected':

          if (this.format.day) {
            this.showView(0);
          } else {
            this.hideView();
          }

          break;

        case 'month':
          viewMonth = $target.parent().children().index($target);
          this.date = new Date(viewYear, viewMonth, Math.min(viewDay, 28), 0, 0, 0, 0);
          this.viewDate = new Date(viewYear, viewMonth, Math.min(viewDay, 28), 0, 0, 0, 0);

          if (this.format.day) {
            this.showView(0);
          } else {
            this.hideView();
          }

          this.change();
          break;

        case 'day prev':
        case 'day next':
        case 'day':
          viewMonth = type === 'day prev' ? viewMonth - 1 : type === 'day next' ? viewMonth + 1 : viewMonth;
          viewDay = parseInt($target.text(), 10);
          this.date = new Date(viewYear, viewMonth, viewDay, 0, 0, 0, 0);
          this.viewDate = new Date(viewYear, viewMonth, viewDay, 0, 0, 0, 0);
          this.fillDays();

          if (type === 'day') {
            this.hideView();
          }

          this.change();
          break;

        case 'day selected':
          this.hideView();
          this.change();
          break;

        case 'day disabled':
          this.hideView();
          break;

        // No default
      }
    },

    destroy: function () {
      this.unbind();
      this.$picker.remove();
      this.$element.removeData(NAMESPACE);
    }
  };

  Datepicker.DEFAULTS = {
    date: false,
    dateFormat: 'mm/dd/yyyy',
    disabledClass: 'disabled',
    selectedClass: 'selected',
    autoClose: false,
    inline: false,
    trigger: false,
    container: false,
    showMonthAfterYear: false,
    zIndex: 1,
    viewStart: 0, // 0 for 'days', 1 for 'months', 2 for 'years'
    weekStart: 0, // 0 for Sunday, 1 for Monday, 2 for Tuesday, 3 for Wednesday, 4 for Thursday, 5 for Friday, 6 for Saturday
    yearSuffix: '',
    days: ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday', 'Sunday'],
    daysShort: ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'],
    daysMin: ['Su', 'Mo', 'Tu', 'We', 'Th', 'Fr', 'Sa', 'Su'],
    months: ['January', 'February', 'March', 'April', 'May', 'June', 'July', 'August', 'September', 'October', 'November', 'December'],
    monthsShort: ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'],
    itemTag: 'li',
    template: (
      '<div class="datepicker-container" data-type="datepicker">' +
        '<div class="datepicker-arrow"></div>' +
        '<div class="datepicker-content">' +
          '<div class="content-years" data-type="years picker">' +
            '<ul class="datepicker-title">' +
              '<li class="datepicker-prev" data-type="years prev">&lsaquo;</li>' +
              '<li class="col-5" data-type="years current"></li>' +
              '<li class="datepicker-next" data-type="years next">&rsaquo;</li>' +
            '</ul>' +
            '<ul class="datepicker-years" data-type="years"></ul>' +
          '</div>' +
          '<div class="content-months" data-type="months picker">' +
            '<ul class="datepicker-title">' +
              '<li class="datepicker-prev" data-type="year prev">&lsaquo;</li>' +
              '<li class="col-5" data-type="year current"></li>' +
              '<li class="datepicker-next" data-type="year next">&rsaquo;</li>' +
            '</ul>' +
            '<ul class="datepicker-months" data-type="months"></ul>' +
          '</div>' +
          '<div class="content-days" data-type="days picker">' +
            '<ul class="datepicker-title">' +
              '<li class="datepicker-prev" data-type="month prev">&lsaquo;</li>' +
              '<li class="col-5" data-type="month current"></li>' +
              '<li class="datepicker-next" data-type="month next">&rsaquo;</li>' +
            '</ul>' +
            '<ul class="datepicker-week" data-type="week"></ul>' +
            '<ul class="datepicker-days" data-type="days"></ul>' +
          '</div>' +
        '</div>' +
      '</div>'
    ),

    isDisabled: function (/*date*/) {
      return false;
    }
  };

  Datepicker.setDefaults = function (options) {
    $.extend(Datepicker.DEFAULTS, $.isPlainObject(options) && options);
  };

  // Save the other datepicker
  Datepicker.other = $.fn.datepicker;

  // Register as jQuery plugin
  $.fn.datepicker = function (options) {
    var args = [].slice.call(arguments, 1);
    var result;

    this.each(function () {
      var $this = $(this);
      var data = $this.data(NAMESPACE);
      var fn;

      if (!data) {
        if (/destroy/.test(options)) {
          return;
        }

        $this.data(NAMESPACE, (data = new Datepicker(this, options)));
      }

      if (typeof options === 'string' && $.isFunction(fn = data[options])) {
        result = fn.apply(data, args);
      }
    });

    return typeof result === 'undefined' ? this : result;
  };

  $.fn.datepicker.Constructor = Datepicker;
  $.fn.datepicker.setDefaults = Datepicker.setDefaults;

  // No conflict
  $.fn.datepicker.noConflict = function () {
    $.fn.datepicker = Datepicker.other;
    return this;
  };

});
