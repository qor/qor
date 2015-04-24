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

  var $window = $(window),
      $document = $(document),

      Datepicker = function (element, options) {
        this.$element = $(element);
        this.options = $.extend({}, Datepicker.DEFAULTS, $.isPlainObject(options) && options);
        this.visible = false;
        this.isInput = false;
        this.isInline = false;
        this.init();
      };

  function isNumber(n) {
    return typeof n === 'number';
  }

  function isUndefined(n) {
    return typeof n === 'undefined';
  }

  function toArray(obj, offset) {
    var args = [];

    if (isNumber(offset)) { // It's necessary for IE8
      args.push(offset);
    }

    return args.slice.apply(obj, args);
  }

  function isLeapYear (year) {
    return (year % 4 === 0 && year % 100 !== 0) || year % 400 === 0;
  }

  function getDaysInMonth (year, month) {
    return [31, (isLeapYear(year) ? 29 : 28), 31, 30, 31, 30, 31, 31, 30, 31, 30, 31][month];
  }

  function parseFormat (format) {
    var separator = format.match(/[.\/\-\s].*?/) || '/',
        parts = format.split(/\W+/),
        length,
        i;

    if (!parts || parts.length === 0) {
      throw new Error('Invalid date format.');
    }

    format = {
      separator: separator[0],
      parts: parts
    };

    for (i = 0, length = parts.length; i < length; i++) {
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
    var parts,
        length,
        year,
        day,
        month,
        val,
        i;

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
        },
        parts = [],
        length = format.parts.length,
        i;

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
      var $this = this.$element,
          options = this.options,
          $picker;

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
      var $this = this.$element,
          options = this.options;

      this.$picker.on('click', $.proxy(this.click, this));

      if (!this.isInline) {
        if (this.isInput) {
          $this.on('keyup', $.proxy(this.update, this));

          if (!options.trigger) {
            $this.on('focus', $.proxy(this.show, this));
          }
        }

        this.$trigger.on('click', $.proxy(this.show, this));
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
      var $trigger = this.$trigger,
          offset = $trigger.offset();

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
        $window.on('resize', $.proxy(this.place, this));
        $document.on('click', $.proxy(function (e) {
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
        $window.off('resize', this.place);
        $document.off('click', this.hide);
      }
    },

    update: function () {
      var $this = this.$element,
          date = $this.data('date') || (this.isInput ? $this.prop('value') : $this.text());

      this.date = date = parseDate(date, this.format);
      this.viewDate = new Date(date.getFullYear(), date.getMonth(), date.getDate());
      this.fillAll();
    },

    change: function () {
      var $this = this.$element,
          date = formatDate(this.date, this.format);

      if (this.isInput) {
        $this.prop('value', date);
      } else if (!this.isInline) {
        $this.text(date);
      }

      $this.data('date', date).trigger('change');
    },

    getMonthByNumber: function (month, short) {
      var options = this.options,
          months = short ? options.monthsShort : options.months;

      return months[isNumber(month) ? month : this.date.getMonth()];
    },

    getDayByNumber: function (day, short, min) {
      var options = this.options,
          days = min ? options.daysMin : short ? options.daysShort : options.days;

      return days[isNumber(day) ? day : this.date.getDay()];
    },

    getDate: function (format) {
      return format ? formatDate(this.date, this.format) : new Date(this.date);
    },

    template: function (data) {
      var options = this.options,
          defaults = {
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
      var title = '',
          items = [],
          suffix = this.options.yearSuffix || '',
          year = this.date.getFullYear(),
          viewYear = this.viewDate.getFullYear(),
          isCurrent,
          i;

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
      var title = '',
          items = [],
          options = this.options.monthsShort,
          year = this.date.getFullYear(),
          month = this.date.getMonth(),
          viewYear = this.viewDate.getFullYear(),
          isCurrent,
          i;

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
      var options = this.options,
          items = [],
          days = options.daysMin,
          weekStart = parseInt(options.weekStart, 10) % 7,
          i;

      days = $.merge(days.slice(weekStart), days.slice(0, weekStart));

      for (i = 0; i < 7; i++) {
        items.push(this.template({
          text: days[i]
        }));
      }

      this.$picker.find('[data-type="week"]').html(items.join(''));
    },

    fillDays: function () {
      var title = '',
          items = [],
          prevItems = [],
          currentItems = [],
          nextItems = [],
          options = this.options.monthsShort,
          suffix = this.options.yearSuffix || '',
          year = this.date.getFullYear(),
          month = this.date.getMonth(),
          day = this.date.getDate(),
          viewYear = this.viewDate.getFullYear(),
          viewMonth = this.viewDate.getMonth(),
          weekStart = parseInt(this.options.weekStart, 10) % 7,
          isCurrent,
          isDisabled,
          length,
          date,
          i,
          n;

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
      var $target = $(e.target),
          yearRegex = /^\d{2,4}$/,
          isYear = false,
          viewYear,
          viewMonth,
          viewDay,
          year,
          type;

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
    $.extend(Datepicker.DEFAULTS, options);
  };

  // Save the other datepicker
  Datepicker.other = $.fn.datepicker;

  // Register as jQuery plugin
  $.fn.datepicker = function (options) {
    var args = toArray(arguments, 1),
        result;

    this.each(function () {
      var $this = $(this),
          data = $this.data('datepicker'),
          fn;

      if (!data) {
        $this.data('datepicker', (data = new Datepicker(this, options)));
      }

      if (typeof options === 'string' && $.isFunction((fn = data[options]))) {
        result = fn.apply(data, args);
      }
    });

    return isUndefined(result) ? this : result;
  };

  $.fn.datepicker.Constructor = Datepicker;
  $.fn.datepicker.setDefaults = Datepicker.setDefaults;

  // No conflict
  $.fn.datepicker.noConflict = function () {
    $.fn.datepicker = Datepicker.other;
    return this;
  };

});

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
    show: false
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

  if (!$.fn.modal) {
    return;
  }

  $(document).on('click.qor.comparator', '[data-toggle="qor.comparator"]', function (e) {
    var $this = $(this),
        data = $this.data('qor.comparator');

    e.preventDefault();

    if (!data) {
      $this.data('qor.comparator', (data = new QorComparator(this, $this.data())));
    }

    data.show();
  });

});

(function (factory) {
  if (typeof define === 'function' && define.amd) {
    // AMD. Register as anonymous module.
    define('qor-confirmer', ['jquery'], factory);
  } else if (typeof exports === 'object') {
    // Node / CommonJS
    factory(require('jquery'));
  } else {
    // Browser globals.
    factory(jQuery);
  }
})(function ($) {

  'use strict';

  $(document).on('click.qor.confirmer', '[data-toggle="qor.confirmer"]', function (e) {
    var message = $(this).data('message');

    if (message && !window.confirm(message)) {
      e.preventDefault();
    }
  });

});

(function (factory) {
  if (typeof define === 'function' && define.amd) {
    // AMD. Register as anonymous module.
    define('qor-cropper', ['jquery'], factory);
  } else if (typeof exports === 'object') {
    // Node / CommonJS
    factory(require('jquery'));
  } else {
    // Browser globals.
    factory(jQuery);
  }
})(function ($) {

  'use strict';

  var URL = window.URL || window.webkitURL,

      QorCropper = function (element, options) {
        this.$element = $(element);
        this.options = $.extend(true, {}, QorCropper.DEFAULTS, options);
        this.built = false;
        this.url = null;
        this.init();
      };

  QorCropper.prototype = {
    constructor: QorCropper,

    init: function () {
      var $this = this.$element,
          options = this.options,
          $parent,
          $image,
          data,
          url;

      if (options.parent) {
        $parent = $this.closest(options.parent);
      }

      if (!$parent.length) {
        $parent = $this.parent();
      }

      if (options.target) {
        $image = $parent.find(options.target);
      }

      if (!$image.length) {
        $image = $('<img>');
      }

      if (options.output) {
        this.$output = $parent.find(options.output);

        try {
          data = JSON.parse(this.$output.val());
        } catch (e) {
          console.log(e.message);
        }
      }

      this.$parent = $parent;
      this.$image = $image;
      $this.on('change', $.proxy(this.read, this));

      this.data = data || {};
      url = $image.data('originalUrl');

      if (!url) {
        url = $image.prop('src');

        if (url && $.isFunction(options.replace)) {
          url = options.replace(url);
        }
      }

      this.load(url);
      $this.on('change', $.proxy(this.read, this));
    },

    read: function () {
      var files = this.$element.prop('files'),
          file;

      if (files) {
        file = files[0];

        if (/^image\/\w+$/.test(file.type) && URL) {
          this.load(URL.createObjectURL(file), true);
        }
      }
    },

    load: function (url, replaced) {
      if (!url) {
        return;
      }

      if (!this.built) {
        this.build();
      }

      if (/^blob:\w+/.test(this.url) && URL) {
        URL.revokeObjectURL(this.url); // Revoke the old one
      }

      this.url = url;

      if (replaced) {
        this.data[this.options.key] = null;
        this.$image.attr('src', url);
      }
    },

    build: function () {
      if (this.built) {
        return;
      }

      this.built = true;

      this.$cropper = $(QorCropper.TEMPLATE).prepend(this.$image).appendTo(this.$parent);
      this.$cropper.find('.modal').on({
        'shown.bs.modal': $.proxy(this.start, this),
        'hidden.bs.modal': $.proxy(this.stop, this)
      });
    },

    start: function () {
      var $modal = this.$cropper.find('.modal'),
          $clone = $('<img>').attr('src', this.url),
          data = this.data,
          key = this.options.key,
          _this = this;

      $modal.find('.modal-body').html($clone);
      $clone.cropper({
        background: false,
        zoomable: false,
        rotatable: false,

        built: function () {
          var previous = data[key],
              scaled = {},
              scaledRatio,
              imageData,
              canvasData;

          if ($.isPlainObject(previous)) {
            imageData = $clone.cropper('getImageData');
            canvasData = $clone.cropper('getCanvasData');
            scaledRatio = imageData.width / imageData.naturalWidth;

            $.each(previous, function (key, val) {
              scaled[String(key).toLowerCase()] = val * scaledRatio;
            });

            $clone.cropper('setCropBoxData', {
              left: scaled.x + canvasData.left,
              top: scaled.y + canvasData.top,
              width: scaled.width,
              height: scaled.height
            });
          }

          $modal.find('.qor-cropper-save').one('click', function () {
            var cropData = $clone.cropper('getData');

            data[key] = {
              x: Math.round(cropData.x),
              y: Math.round(cropData.y),
              width: Math.round(cropData.width),
              height: Math.round(cropData.height)
            };

            _this.output($clone.cropper('getCroppedCanvas').toDataURL());
            $modal.modal('hide');
          });
        }
      });
    },

    stop: function () {
      this.$cropper.find('.modal-body > img').cropper('destroy').remove();
    },

    output: function (url) {
      var data = $.extend({}, this.data, this.options.data);

      this.$image.attr('src', url);
      this.$output.val(JSON.stringify(data));
    },

    destroy: function () {
      this.$element.off('change');
      this.$cropper.find('.modal').off('shown.bs.modal hidden.bs.modal');
    }
  };

  QorCropper.DEFAULTS = {
    target: '',
    output: '',
    parent: '',
    key: 'qorCropper',
    data: null
  };

  QorCropper.TEMPLATE = (
    '<div class="qor-cropper">' +
      '<a class="qor-cropper-toggle" data-toggle="modal" href="#qorCropperModal" title="Crop the image"><span class="sr-only">Toggle Cropper<span></a>' +
      '<div class="modal fade qor-cropper-modal" id="qorCropperModal" tabindex="-1" role="dialog" aria-labelledby="qorCropperModalLabel" aria-hidden="true">' +
        '<div class="modal-dialog">' +
          '<div class="modal-content">' +
            '<div class="modal-header">' +
              '<h5 class="modal-title" id="qorCropperModalLabel">Crop the image</h5>' +
            '</div>' +
            '<div class="modal-body"></div>' +
            '<div class="modal-footer">' +
              '<button type="button" class="btn btn-link" data-dismiss="modal">Cancel</button>' +
              '<button type="button" class="btn btn-link qor-cropper-save">OK</button>' +
            '</div>' +
          '</div>' +
        '</div>' +
      '</div>' +
    '</div>'
  );

  $(function () {
    if (!$.fn.cropper) {
      return;
    }

    $('input[data-toggle="qor.cropper"]').each(function () {
      var $this = $(this);

      if (!$this.data('qor.cropper')) {
        $this.data('qor.cropper', new QorCropper(this, {
          target: 'img',
          output: 'textarea',
          parent: '.form-group',
          key: 'CropOption',
          data: {
            Crop: true
          },
          replace: function (url) {
            return url.replace(/\.\w+$/, function (extension) {
              return '.original' + extension;
            });
          }
        }));
      }
    });
  });

  return QorCropper;

});

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

(function (factory) {
  if (typeof define === 'function' && define.amd) {
    // AMD. Register as anonymous module.
    define('qor-filter', ['jquery'], factory);
  } else if (typeof exports === 'object') {
    // Node / CommonJS
    factory(require('jquery'));
  } else {
    // Browser globals.
    factory(jQuery);
  }
})(function ($) {

  'use strict';

  var location = window.location,

      QorFilter = function (element, options) {
        this.$element = $(element);
        this.options = $.extend({}, QorFilter.DEFAULTS, options);
        this.init();
      };

  // Array: Extend a with b
  function extend(a, b) {
    $.each(b, function (i, n) {
      if ($.inArray(n, a) === -1) {
        a.push(n);
      }
    });

    return a;
  }

  // Array: detach b from a
  function detach(a, b) {
    $.each(b, function (i, n) {
      i = $.inArray(n, a);

      if (i > -1) {
        a.splice(i, 1);
      }
    });

    return a;
  }

  function encodeSearch(data, detched) {
    var search = location.search,
        params = [],
        source;

    if ($.isPlainObject(data)) {
      source = decodeSearch(search);

      $.each(data, function (name, values) {
        if ($.isArray(source[name])) {
          if (!detched) {
            source[name] = extend(source[name], values);
          } else {
            source[name] = detach(source[name], values);
          }

        } else {
          if (!detched) {
            source[name] = values;
          }
        }
      });

      $.each(source, function (name, values) {
        params = params.concat($.map(values, function (value) {
          return value === '' ? name : [name, value].join('=');
        }));
      });

      search = '?' + params.join('&');
    }

    return search;
  }

  function decodeSearch(search) {
    var data = {},
        params = [];

    if (search && search.indexOf('?') > -1) {
      search = search.split('?')[1];

      if (search && search.indexOf('#') > -1) {
        search = search.split('#')[0];
      }

      if (search) {
        params = search.split('&');
      }

      $.each(params, function (i, n) {
        var param = n.split('='),
            name = param[0].toLowerCase(),
            value = param[1] || '';

        if ($.isArray(data[name])) {
          data[name].push(value);
        } else {
          data[name] = [value];
        }
      });
    }

    return data;
  }

  QorFilter.prototype = {
    constructor: QorFilter,

    init: function () {
      var $this = this.$element,
          options = this.options,
          $toggle = $this.find(options.toggle);

      if (!$toggle.length) {
        return;
      }

      this.$toggle = $toggle;
      this.parse();
      this.bind();
    },

    bind: function () {
      this.$element.on('click', this.options.toggle, $.proxy(this.toggle, this));
    },

    parse: function () {
      var options = this.options,
          data = decodeSearch(location.search);

      this.$toggle.each(function () {
        var $this = $(this),
            params = decodeSearch($this.attr('href'));

        $.each(data, function (name, value) {
          var matched = false;

          $.each(params, function (key, val) {
            var equal = false;

            if (key === name) {
              $.each(val, function (i, n) {
                if ($.inArray(n, value) > -1) {
                  equal = true;
                  return false;
                }
              });
            }

            if (equal) {
              matched = true;
              return false;
            }
          });

          $this.toggleClass(options.activeClass, matched);

          if (matched) {
            return false;
          }
        });
      });
    },

    toggle: function (e) {
      var $target = $(e.target),
          data = decodeSearch(e.target.href),
          search;

      e.preventDefault();

      if ($target.hasClass(this.options.activeClass)) {
        search = encodeSearch(data, true); // set `true` to detch
      } else {
        search = encodeSearch(data);
      }

      location.search = search;
    }
  };

  QorFilter.DEFAULTS = {
    toggle: '',
    activeClass: 'active'
  };

  $(function () {
    $('.qor-label-group').each(function () {
      var $this = $(this);

      if (!$this.data('qor.filter')) {
        $this.data('qor.filter', new QorFilter(this, {
          toggle: '.label',
          activeClass: 'label-primary'
        }));
      }
    });
  });

});

(function (factory) {
  if (typeof define === 'function' && define.amd) {
    // AMD. Register as anonymous module.
    define('qor-redactor', ['jquery'], factory);
  } else if (typeof exports === 'object') {
    // Node / CommonJS
    factory(require('jquery'));
  } else {
    // Browser globals.
    factory(jQuery);
  }
})(function ($) {

  'use strict';

  var NAMESPACE = '.qor.redactor',
      EVENT_CLICK = 'click' + NAMESPACE,
      EVENT_FOCUS = 'focus' + NAMESPACE,
      EVENT_BLUR = 'blur' + NAMESPACE,
      EVENT_IMAGE_UPLOAD = 'imageupload' + NAMESPACE,
      EVENT_IMAGE_DELETE = 'imagedelete' + NAMESPACE,
      REGEXP_OPTIONS = /x|y|width|height/,

      QorRedactor = function (element, options) {
        this.$element = $(element);
        this.options = $.extend(true, {}, QorRedactor.DEFAULTS, options);
        this.init();
      };

  function encodeCropData(data) {
    var nums = [];

    if ($.isPlainObject(data)) {
      $.each(data, function () {
        nums.push(arguments[1]);
      });
    }

    return nums.join();
  }

  function decodeCropData(data) {
    var nums = data && data.split(',');

    data = null;

    if (nums && nums.length === 4) {
      data = {
        x: nums[0],
        y: nums[1],
        width: nums[2],
        height: nums[3]
      };
    }

    return data;
  }

  QorRedactor.prototype = {
    constructor: QorRedactor,

    init: function () {
      var _this = this,
          $this = this.$element,
          options = this.options,
          $parent = $this.closest(options.parent),
          click = $.proxy(this.click, this);

      this.$button = $(QorRedactor.BUTTON);

      $this.on(EVENT_IMAGE_UPLOAD, function (e, image) {
        $(image).on(EVENT_CLICK, click);
      }).on(EVENT_IMAGE_DELETE, function (e, image) {
        $(image).off(EVENT_CLICK, click);
      }).on(EVENT_FOCUS, function (e) {
        console.log(e.type);
        $parent.find('img').off(EVENT_CLICK, click).on(EVENT_CLICK, click);
      }).on(EVENT_BLUR, function (e) {
        console.log(e.type);
        $parent.find('img').off(EVENT_CLICK, click);
      });

      $('body').on(EVENT_CLICK, function () {
        _this.$button.off(EVENT_CLICK).detach();
      });
    },

    click: function (e) {
      var _this = this,
          $image = $(e.target);

      e.stopPropagation();

      setTimeout(function () {
        _this.$button.insertBefore($image).off(EVENT_CLICK).one(EVENT_CLICK, function () {
          _this.crop($image);
        });
      }, 1);
    },

    crop: function ($image) {
      var options = this.options,
          url = $image.attr('src'),
          originalUrl = url,
          $clone = $('<img>'),
          $modal = $(QorRedactor.TEMPLATE);

      if ($.isFunction(options.replace)) {
        originalUrl = options.replace(originalUrl);
      }

      $clone.attr('src', originalUrl);
      $modal.appendTo('body').modal('show').find('.modal-body').append($clone);

      $modal.one('shown.bs.modal', function () {
        $clone.cropper({
          background: false,
          zoomable: false,
          rotatable: false,

          built: function () {
            var data = decodeCropData($image.attr('data-crop-option')),
                canvasData,
                imageData;

            if ($.isPlainObject(data)) {
              imageData = $clone.cropper('getImageData');
              canvasData = $clone.cropper('getCanvasData');
              imageData.ratio = imageData.width / imageData.naturalWidth;

              $clone.cropper('setCropBoxData', {
                left: data.x * imageData.ratio + canvasData.left,
                top: data.y * imageData.ratio + canvasData.top,
                width: data.width * imageData.ratio,
                height: data.height * imageData.ratio
              });
            }

            $modal.find('.qor-cropper-save').one('click', function () {
              var cropData = {};

              $.each($clone.cropper('getData'), function (i, n) {
                if (REGEXP_OPTIONS.test(i)) {
                  cropData[i] = Math.round(n);
                }
              });

              $.ajax(options.remote, {
                type: 'POST',
                contentType: 'application/json',
                data: JSON.stringify({
                  Url: url,
                  CropOption: cropData,
                  Crop: true
                }),
                dataType: 'json',

                success: function (response) {
                  if ($.isPlainObject(response) && response.url) {
                    $image.attr('src', response.url).attr('data-crop-option', encodeCropData(cropData)).removeAttr('style').removeAttr('rel');

                    if ($.isFunction(options.complete)) {
                      options.complete();
                    }

                    $modal.modal('hide');
                  }
                },

                error: function () {
                  console.log(arguments);
                }
              });
            });
          }
        });
      }).one('hidden.bs.modal', function () {
        $clone.cropper('destroy').remove();
        $modal.remove();
      });
    }
  };

  QorRedactor.DEFAULTS = {
    remote: false,
    toggle: false,
    parent: false,
    replace: null,
    complete: null
  };

  QorRedactor.BUTTON = '<span class="redactor-image-cropper">Crop</span>';

  QorRedactor.TEMPLATE = (
    '<div class="modal fade qor-cropper-modal" id="qorCropperModal" tabindex="-1" role="dialog" aria-labelledby="qorCropperModalLabel" aria-hidden="true">' +
      '<div class="modal-dialog">' +
        '<div class="modal-content">' +
          '<div class="modal-header">' +
            '<h5 class="modal-title" id="qorCropperModalLabel">Crop the image</h5>' +
          '</div>' +
          '<div class="modal-body"></div>' +
          '<div class="modal-footer">' +
            '<button type="button" class="btn btn-link" data-dismiss="modal">Cancel</button>' +
            '<button type="button" class="btn btn-link qor-cropper-save">OK</button>' +
          '</div>' +
        '</div>' +
      '</div>' +
    '</div>'
  );

  $(function () {
    if (!$.fn.redactor) {
      return;
    }

    $('textarea[data-toggle="qor.redactor"]').each(function () {
      var $this = $(this),
          data = $this.data();

      $this.redactor({
        imageUpload: data.uploadUrl,
        fileUpload: data.uploadUrl,

        initCallback: function () {
          if (!$this.data('qor.redactor')) {
            $this.data('qor.redactor', new QorRedactor($this, {
              remote: data.cropUrl,
              toggle: '.redactor-image-cropper',
              parent: '.form-group',
              replace: function (url) {
                return url.replace(/\.\w+$/, function (extension) {
                  return '.original' + extension;
                });
              },
              complete: $.proxy(function () {
                this.code.sync();
              }, this)
            }));
          }
        },

        focusCallback: function (/*e*/) {
          $this.triggerHandler(EVENT_FOCUS);
        },

        blurCallback: function (/*e*/) {
          $this.triggerHandler(EVENT_BLUR);
        },

        imageUploadCallback: function (/*image, json*/) {
          $this.triggerHandler(EVENT_IMAGE_UPLOAD, arguments[0]);
        },

        imageDeleteCallback: function (/*url, image*/) {
          $this.triggerHandler(EVENT_IMAGE_DELETE, arguments[1]);
        }
      });
    });
  });

});

(function (factory) {
  if (typeof define === 'function' && define.amd) {
    // AMD. Register as anonymous module.
    define('qor-replicator', ['jquery'], factory);
  } else if (typeof exports === 'object') {
    // Node / CommonJS
    factory(require('jquery'));
  } else {
    // Browser globals.
    factory(jQuery);
  }
})(function ($) {

  'use strict';

  var QorReplicator = function (element, options) {
        this.$element = $(element);
        this.options = $.extend({}, QorReplicator.DEFAULTS, options);
        this.index = 0;
        this.init();
      };

  QorReplicator.prototype = {
    constructor: QorReplicator,

    init: function () {
      var $this = this.$element,
          options = this.options,
          $all = $this.find(options.itemClass),
          $template;

      if (!$all.length) {
        return;
      }

      $template = $all.filter(options.newClass);

      if (!$template.length) {
        $template = $all.last();
      }

      this.template = $template.prop('outerHTML');
      this.parse();
      this.bind();
    },

    parse: function () {
      var i = 0;

      this.template = this.template.replace(/(\w+)\="(\S*\[\d+\]\S*)"/g, function (attribute, name, value) {
        value = value.replace(/^(\S*)\[(\d+)\]([^\[\]]*)$/, function (input, prefix, index, suffix) {
          if (input === value) {
            if (name === 'name') {
              i = index;
            }

            return (prefix + '[{{index}}]' + suffix);
          }
        });

        return (name + '="' + value + '"');
      });

      this.index = parseFloat(i);
    },

    bind: function () {
      var $this = this.$element,
          options = this.options;

      $this.on('click', options.addClass, $.proxy(this.add, this));
      $this.on('click', options.delClass, $.proxy(this.del, this));
    },

    add: function () {
      this.$element.find(this.options.itemClass).last().after(this.template.replace(/\{\{index\}\}/g, ++this.index));
    },

    del: function (e) {
      var options = this.options,
          $item = $(e.target).closest(options.itemClass),
          $alert;

      if ($item.is(options.newClass)) {
        $item.remove();
      } else {
        $item.children(':visible').addClass('hidden').hide();
        $alert = $(options.alertTemplate.replace('{{name}}', this.parseName($item)));
        $alert.find(options.undoClass).one('click', function () {
          $alert.remove();
          $item.children('.hidden').removeClass('hidden').show();
        });
        $item.append($alert);
      }
    },

    parseName: function ($item) {
      var name = $item.find('input[name]').attr('name');

      if (name) {
        return name.replace(/[^\[\]]+$/, '');
      }
    }
  };

  QorReplicator.DEFAULTS = {
    itemClass: '',
    newClass: '',
    addClass: '',
    delClass: '',
    alertTemplate: ''
  };

  $(function () {
    $('.qor-collection-group').each(function () {
      var $this = $(this);

      if (!$this.data('qor.replicator')) {
        $this.data('qor.replicator', new QorReplicator(this, {
          itemClass: '.qor-collection',
          newClass: '.qor-collection-new',
          addClass: '.qor-collection-add',
          delClass: '.qor-collection-del',
          undoClass: '.qor-collection-undo',
          alertTemplate: '<div class="alert alert-danger"><input type="hidden" name="{{name}}._destroy" value="1"><a href="javascript:void(0);" class="alert-link qor-collection-undo">Undo Delete</a></div>'
        }));
      }
    });
  });

});

(function (factory) {
  if (typeof define === 'function' && define.amd) {
    // AMD. Register as anonymous module.
    define('qor-selector', ['jquery'], factory);
  } else if (typeof exports === 'object') {
    // Node / CommonJS
    factory(require('jquery'));
  } else {
    // Browser globals.
    factory(jQuery);
  }
})(function ($) {

  'use strict';

  $(function () {
    if (!$.fn.chosen) {
      return;
    }

    $('select[data-toggle="qor.selector"]').each(function () {
      var $this = $(this);

      if (!$this.prop('multiple') && !$this.find('option[selected]').length) {
        $this.prepend('<option value="" selected></option>');
      }

      $this.chosen();
    });
  });

});
