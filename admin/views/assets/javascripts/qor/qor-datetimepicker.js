(function ($, moment) {
  var pluginName = 'materialDatePicker';
  var pluginDataName = 'plugin_' + pluginName;

  moment.locale('en');

  function Plugin(element, options) {
    this.currentView = 0;

    this.minDate = null;
    this.maxDate = null;

    this._attachedEvents = [];

    this.element = element;
    this.$element = $(element);

    this.params = {
      date: true,
      time: true,
      format: 'YYYY-MM-DD',
      minDate: null,
      maxDate: null,
      currentDate: null,
      lang: 'en',
      weekStart: 0,
      shortTime: false,
      cancelText: 'Cancel',
      okText: 'OK'
    };
    this.params = $.fn.extend(this.params, options);

    this.name = 'qor_' + this.setName();
    this.$element.attr('data-dtp', this.name);

    this.init();
  }

  $.fn[pluginName] = function (options, p) {
    this.each(function () {
      if (!$.data(this, pluginDataName)) {
        $.data(this, pluginDataName, new Plugin(this, options));
      } else {
        if (typeof ($.data(this, pluginDataName)[options]) === 'function') {
          $.data(this, pluginDataName)[options](p);
        }
        if (options === 'destroy') {
          delete $.data(this, pluginDataName);
        }
      }
    });
    return this;
  };

  Plugin.prototype = {
    init: function () {
      this.initDays();
      this.initDates();

      this.initTemplate();

      this.initButtons();

      this._attachEvent($(window), 'resize', this._centerBox(this));
      this._attachEvent(this.$dtpElement.find('.qor-datepicker__content'), 'click', this._onElementClick.bind(this));
      this._attachEvent(this.$dtpElement, 'click', this._onBackgroundClick.bind(this));
      this._attachEvent(this.$dtpElement.find('.qor-close > a'), 'click', this._onCloseClick.bind(this));
      this._attachEvent(this.$element, 'focus', this._onFocus.bind(this));
    },
    initDays: function () {
      this.days = [];
      for (var i = this.params.weekStart; this.days.length < 7; i++) {
        if (i > 6) {
          i = 0;
        }
        this.days.push(i.toString());
      }
    },
    initDates: function () {
      if (this.$element.val().length > 0) {
        if (typeof (this.params.format) !== 'undefined' && this.params.format !== null) {
          this.currentDate = moment(this.$element.val(), this.params.format).locale(this.params.lang);
        } else {
          this.currentDate = moment(this.$element.val()).locale(this.params.lang);
        }
      } else {
        if (typeof (this.$element.attr('value')) !== 'undefined' && this.$element.attr('value') !== null && this.$element.attr('value') !== '') {
          if (typeof (this.$element.attr('value')) === 'string') {
            if (typeof (this.params.format) !== 'undefined' && this.params.format !== null) {
              this.currentDate = moment(this.$element.attr('value'), this.params.format).locale(this.params.lang);
            } else {
              this.currentDate = moment(this.$element.attr('value')).locale(this.params.lang);
            }
          }
        } else {
          if (typeof (this.params.currentDate) !== 'undefined' && this.params.currentDate !== null) {
            if (typeof (this.params.currentDate) === 'string') {
              if (typeof (this.params.format) !== 'undefined' && this.params.format !== null) {
                this.currentDate = moment(this.params.currentDate, this.params.format).locale(this.params.lang);
              } else {
                this.currentDate = moment(this.params.currentDate).locale(this.params.lang);
              }
            } else {
              if (typeof (this.params.currentDate.isValid) === 'undefined' || typeof (this.params.currentDate.isValid) !== 'function') {
                var x = this.params.currentDate.getTime();
                this.currentDate = moment(x, 'x').locale(this.params.lang);
              } else {
                this.currentDate = this.params.currentDate;
              }
            }
            this.$element.val(this.currentDate.format(this.params.format));
          } else
            this.currentDate = moment();
        }
      }

      if (typeof (this.params.minDate) !== 'undefined' && this.params.minDate !== null) {
        if (typeof (this.params.minDate) === 'string') {
          if (typeof (this.params.format) !== 'undefined' && this.params.format !== null) {
            this.minDate = moment(this.params.minDate, this.params.format).locale(this.params.lang);
          } else {
            this.minDate = moment(this.params.minDate).locale(this.params.lang);
          }
        } else {
          if (typeof (this.params.minDate.isValid) === 'undefined' || typeof (this.params.minDate.isValid) !== 'function') {
            var x = this.params.minDate.getTime();
            this.minDate = moment(x, 'x').locale(this.params.lang);
          } else {
            this.minDate = this.params.minDate;
          }
        }
      }

      if (typeof (this.params.maxDate) !== 'undefined' && this.params.maxDate !== null) {
        if (typeof (this.params.maxDate) === 'string') {
          if (typeof (this.params.format) !== 'undefined' && this.params.format !== null) {
            this.maxDate = moment(this.params.maxDate, this.params.format).locale(this.params.lang);
          } else {
            this.maxDate = moment(this.params.maxDate).locale(this.params.lang);
          }
        } else {
          if (typeof (this.params.maxDate.isValid) === 'undefined' || typeof (this.params.maxDate.isValid) !== 'function') {
            var x = this.params.maxDate.getTime();
            this.maxDate = moment(x, 'x').locale(this.params.lang);
          } else {
            this.maxDate = this.params.maxDate;
          }
        }
      }

      if (!this.isAfterMinDate(this.currentDate)) {
        this.currentDate = moment(this.minDate);
      }
      if (!this.isBeforeMaxDate(this.currentDate)) {
        this.currentDate = moment(this.maxDate);
      }
    },
    initTemplate: function () {
      this.template = '<div class="qor-datepicker hidden" id="' + this.name + '">' +
        '<div class="qor-datepicker__content">' +
        '<div class="qor-date-view">' +
        '<header class="qor-header">' +
        '<div class="qor-actual-day">Lundi</div>' +
        '<div class="qor-close"><a href="javascript:void(0);"><i class="material-icons">clear</i></</div>' +
        '</header>' +

        '<div class="qor-date hidden">' +
        '<div>' +
        '<div class="left center p10">' +
        '<a href="javascript:void(0);" class="qor-select-year-before"><i class="material-icons">chevron_left</i></a>' +
        '</div>' +
        '<span class="qor-actual-year">2014</span>' +
        '<div class="right center p10">' +
        '<a href="javascript:void(0);" class="qor-select-year-after"><i class="material-icons">chevron_right</i></a>' +
        '</div>' +
        '</div>' +

        '<div>' +
        '<div class="left center p10">' +
        '<a href="javascript:void(0);" class="qor-select-month-before"><i class="material-icons">chevron_left</i></a>' +
        '</div>' +
        '<span><span class="qor-actual-month ">MAR</span> <span class="qor-actual-num">13</span></span>' +
        '<div class="right center p10">' +
        '<a href="javascript:void(0);" class="qor-select-month-after"><i class="material-icons">chevron_right</i></a>' +
        '</div>' +
        '</div>' +

        '</div>' +

        '<div class="qor-time hidden">' +
        '<div class="qor-actual-maxtime">23:55</div>' +
        '</div>' +
        '<div class="qor-picker">' +
        '<div class="qor-picker-calendar"></div>' +
        '<div class="qor-picker-datetime hidden">' +
        '<div class="qor-actual-meridien">' +
        '<div class="left p20">' +
        '<a class="qor-meridien-am" href="javascript:void(0);">AM</a>' +
        '</div>' +
        '<div class="qor-actual-time p60"></div>' +
        '<div class="right p20">' +
        '<a class="qor-meridien-pm" href="javascript:void(0);">PM</a>' +
        '</div>' +
        '</div>' +
        '<div class="qor-picker-clock"></div>' +
        '</div>' +
        '</div>' +
        '</div>' +
        '<div class="qor-datepicker__buttons">' +
        '<button class="qor-datepicker__save mdl-button mdl-button--colored mdl-js-button mdl-js-ripple-effect">' + this.params.okText + '</button>' +
        '<button class="qor-datepicker__cancel mdl-button mdl-button--colored mdl-js-button mdl-js-ripple-effect">' + this.params.cancelText + '</button>' +
        '</div>' +
        '</div>' +
        '</div>';

      if ($('body').find('#' + this.name).length <= 0) {
        $('body').append(this.template);

        this.dtpElement = $('body').find('#' + this.name);
        this.$dtpElement = $(this.dtpElement);
      }
    },
    initButtons: function () {
      this._attachEvent(this.$dtpElement.find('.qor-datepicker__cancel'), 'click', this._onCancelClick.bind(this));
      this._attachEvent(this.$dtpElement.find('.qor-datepicker__save'), 'click', this._onOKClick.bind(this));
      this._attachEvent(this.$dtpElement.find('a.qor-select-month-before'), 'click', this._onMonthBeforeClick.bind(this));
      this._attachEvent(this.$dtpElement.find('a.qor-select-month-after'), 'click', this._onMonthAfterClick.bind(this));
      this._attachEvent(this.$dtpElement.find('a.qor-select-year-before'), 'click', this._onYearBeforeClick.bind(this));
      this._attachEvent(this.$dtpElement.find('a.qor-select-year-after'), 'click', this._onYearAfterClick.bind(this));
    },
    initMeridienButtons: function () {
      this.$dtpElement.find('a.qor-meridien-am').off('click').on('click', this._onSelectAM.bind(this));
      this.$dtpElement.find('a.qor-meridien-pm').off('click').on('click', this._onSelectPM.bind(this));
    },
    initDate: function (d) {
      this.currentView = 0;

      this.$dtpElement.find('.qor-picker-calendar').removeClass('hidden');
      this.$dtpElement.find('.qor-picker-datetime').addClass('hidden');

      var _date = ((typeof (this.currentDate) !== 'undefined' && this.currentDate !== null) ? this.currentDate : null);
      var _calendar = this.generateCalendar(this.currentDate);

      if (typeof (_calendar.week) !== 'undefined' && typeof (_calendar.days) !== 'undefined') {
        var _template = this.constructHTMLCalendar(_date, _calendar);

        this.$dtpElement.find('a.qor-select-day').off('click');
        this.$dtpElement.find('.qor-picker-calendar').html(_template);

        this.$dtpElement.find('a.qor-select-day').on('click', this._onSelectDate.bind(this));

        this.toggleButtons(_date);
      }

      this._centerBox();
      this.showDate(_date);
    },
    initHours: function () {
      this.currentView = 1;

      if (!this.params.date) {
        var w = this.$dtpElement.find('.qor-datepicker__content').width();

        var ml = this.$dtpElement.find('.qor-picker-clock').css('marginLeft').replace('px', '');
        var mr = this.$dtpElement.find('.qor-picker-clock').css('marginRight').replace('px', '');

        var pl = this.$dtpElement.find('.qor-picker').css('paddingLeft').replace('px', '');
        var pr = this.$dtpElement.find('.qor-picker').css('paddingRight').replace('px', '');

        this.$dtpElement.find('.qor-picker-clock').innerWidth(w - (parseInt(ml) + parseInt(mr) + parseInt(pl) + parseInt(pr)));
      }

      this.showTime(this.currentDate);
      this.initMeridienButtons();

      this.$dtpElement.find('.qor-picker-datetime').removeClass('hidden');
      this.$dtpElement.find('.qor-picker-calendar').addClass('hidden');

      if (this.currentDate.hour() < 12) {
        this.$dtpElement.find('a.qor-meridien-am').click();
      } else {
        this.$dtpElement.find('a.qor-meridien-pm').click();
      }

      var hours = [];

      for (var h = 0; h < 12; ++h) {

        var hour = $('<div>', {
          class: 'qor-picker-time qor-picker-time-' + (h + 1)
        });

        var cH = ((this.currentDate.format('h') == 12) ? 0 : this.currentDate.format('h'));
        var hourLink = $('<a>', {
          href: 'javascript:void(0);',
          class: 'qor-select-hour'
        }).data('hour', h).text((h == 0 ? 12 : h));
        if (h == parseInt(cH))
          hourLink.addClass('selected');

        hour.append(hourLink);
        hours.push(hour);
      }

      this.$dtpElement.find('a.qor-select-hour').off('click');

      this.$dtpElement.find('.qor-picker-clock').html(hours);
      this.toggleTime(true);

      this.initHands(true);
    },
    initMinutes: function () {
      this.currentView = 2;

      this.showTime(this.currentDate);

      this.initMeridienButtons();

      if (this.currentDate.hour() < 12) {
        this.$dtpElement.find('a.qor-meridien-am').click();
      } else {
        this.$dtpElement.find('a.qor-meridien-pm').click();
      }

      this.$dtpElement.find('.qor-picker-calendar').addClass('hidden');
      this.$dtpElement.find('.qor-picker-datetime').removeClass('hidden');

      var minutes = [];

      for (var m = 0; m < 60; m += 5) {

        var minute = $('<div>', {
          class: 'qor-picker-time qor-picker-time-' + (m / 5 + 1)
        });

        var minuteLink = $('<a>', {
          href: 'javascript:void(0);',
          class: 'qor-select-minute'
        }).data('minute', m).text(((m.toString().length == 2) ? m : '0' + m));
        if (m == 5 * Math.round(this.currentDate.minute() / 5)) {
          minuteLink.addClass('selected');
          this.currentDate.minute(m);
        }

        minute.append(minuteLink);
        minutes.push(minute);
      }

      this.$dtpElement.find('a.qor-select-minute').off('click');

      this.$dtpElement.find('.qor-picker-clock').html(minutes);
      this.toggleTime(false);

      this.initHands(false);

      this._centerBox();
    },
    initHands: function (t) {
      this.$dtpElement.find('.qor-picker-clock').append(
        '<div class="qor-hand qor-hour-hand"></div>' +
        '<div class="qor-hand qor-minute-hand"></div>' +
        '<div class="qor-clock-center"></div>'
      );

      this.$dtpElement.find('.qor-hour-hand').addClass((t === true) ? 'on' : '');
      this.$dtpElement.find('.qor-minute-hand').addClass((t === false) ? 'on' : '');

      this.animateHands();

      this._centerBox();
    },
    animateHands: function () {
      var h = this.currentDate.hour();
      var m = this.currentDate.minute();

      this.rotateElement(this.$dtpElement.find('.qor-hour-hand'), (360 / 12) * h);
      this.rotateElement(this.$dtpElement.find('.qor-minute-hand'), ((360 / 60) * (5 * Math.round(this.currentDate.minute() / 5))));
    },
    isAfterMinDate: function (date, checkHour, checkMinute) {
      var _return = true;

      if (typeof (this.minDate) !== 'undefined' && this.minDate !== null) {
        var _minDate = moment(this.minDate);
        var _date = moment(date);

        if (!checkHour && !checkMinute) {
          _minDate.hour(0);
          _minDate.minute(0);

          _date.hour(0);
          _date.minute(0);
        }

        _minDate.second(0);
        _date.second(0);
        _minDate.millisecond(0);
        _date.millisecond(0);

        if (!checkMinute) {
          _date.minute(0);
          _minDate.minute(0);

          _return = (parseInt(_date.format('X')) >= parseInt(_minDate.format('X')));
        } else {
          _return = (parseInt(_date.format('X')) >= parseInt(_minDate.format('X')));
        }
      }

      return _return;
    },
    isBeforeMaxDate: function (date, checkTime, checkMinute) {
      var _return = true;

      if (typeof (this.maxDate) !== 'undefined' && this.maxDate !== null) {
        var _maxDate = moment(this.maxDate);
        var _date = moment(date);

        if (!checkTime && !checkMinute) {
          _maxDate.hour(0);
          _maxDate.minute(0);

          _date.hour(0);
          _date.minute(0);
        }

        _maxDate.second(0);
        _date.second(0);
        _maxDate.millisecond(0);
        _date.millisecond(0);

        if (!checkMinute) {
          _date.minute(0);
          _maxDate.minute(0);

          _return = (parseInt(_date.format('X')) <= parseInt(_maxDate.format('X')));
        } else {
          _return = (parseInt(_date.format('X')) <= parseInt(_maxDate.format('X')));
        }
      }

      return _return;
    },
    rotateElement: function (el, deg) {
      $(el).css({
        WebkitTransform: 'rotate(' + deg + 'deg)',
        '-moz-transform': 'rotate(' + deg + 'deg)'
      });
    },
    showDate: function (date) {
      if (date) {
        this.$dtpElement.find('.qor-actual-day').html(date.locale(this.params.lang).format('dddd'));
        this.$dtpElement.find('.qor-actual-month').html(date.locale(this.params.lang).format('MMM').toUpperCase());
        this.$dtpElement.find('.qor-actual-num').html(date.locale(this.params.lang).format('DD'));
        this.$dtpElement.find('.qor-actual-year').html(date.locale(this.params.lang).format('YYYY'));
      }
    },
    showTime: function (date) {
      if (date) {
        var minutes = (5 * Math.round(date.minute() / 5));
        var content = ((this.params.shortTime) ? date.format('hh') : date.format('HH')) + ':' + ((minutes.toString().length == 2) ? minutes : '0' + minutes);

        if (this.params.date)
          this.$dtpElement.find('.qor-actual-time').html(content);
        else {
          if (this.params.shortTime)
            this.$dtpElement.find('.qor-actual-day').html(date.format('A'));
          else
            this.$dtpElement.find('.qor-actual-day').html(' ');

          this.$dtpElement.find('.qor-actual-maxtime').html(content);
        }
      }
    },
    selectDate: function (date) {
      if (date) {
        this.currentDate.date(date);

        this.showDate(this.currentDate);
        this.$element.trigger('dateSelected', this.currentDate);
      }
    },
    generateCalendar: function (date) {
      var _calendar = {};

      if (date !== null) {
        var startOfMonth = moment(date).locale(this.params.lang).startOf('month');
        var endOfMonth = moment(date).locale(this.params.lang).endOf('month');

        var iNumDay = startOfMonth.format('d');

        _calendar.week = this.days;
        _calendar.days = [];

        for (var i = startOfMonth.date(); i <= endOfMonth.date(); i++) {
          if (i === startOfMonth.date()) {
            var iWeek = _calendar.week.indexOf(iNumDay.toString());
            if (iWeek > 0) {
              for (var x = 0; x < iWeek; x++) {
                _calendar.days.push(0);
              }
            }
          }
          _calendar.days.push(moment(startOfMonth).locale(this.params.lang).date(i));
        }
      }

      return _calendar;
    },
    constructHTMLCalendar: function (date, calendar) {
      var _template = '';

      _template += '<div class="qor-picker-month">' + date.locale(this.params.lang).format('MMMM YYYY') + '</div>';
      _template += '<table class="table qor-picker-days"><thead>';
      for (var i = 0; i < calendar.week.length; i++) {
        _template += '<th>' + moment(parseInt(calendar.week[i]), 'd').locale(this.params.lang).format('dd').substring(0, 1) + '</th>';
      }

      _template += '</thead>';
      _template += '<tbody><tr>';

      for (var i = 0; i < calendar.days.length; i++) {
        if (i % 7 == 0)
          _template += '</tr><tr>';
        _template += '<td data-date="' + moment(calendar.days[i]).locale(this.params.lang).format('D') + '">';
        if (calendar.days[i] != 0) {
          if (this.isBeforeMaxDate(moment(calendar.days[i]), false, false) === false || this.isAfterMinDate(moment(calendar.days[i]), false, false) === false) {
            _template += '<span class="qor-select-day">' + moment(calendar.days[i]).locale(this.params.lang).format('DD') + '</span>';
          } else {
            if (moment(calendar.days[i]).locale(this.params.lang).format('DD') === moment(this.currentDate).locale(this.params.lang).format('DD')) {
              _template += '<a href="javascript:void(0);" class="qor-select-day selected">' + moment(calendar.days[i]).locale(this.params.lang).format('DD') + '</a>';
            } else {
              _template += '<a href="javascript:void(0);" class="qor-select-day">' + moment(calendar.days[i]).locale(this.params.lang).format('DD') + '</a>';
            }
          }

          _template += '</td>';
        }
      }
      _template += '</tr></tbody></table>';

      return _template;
    },
    setName: function () {
      var text = '';
      var possible = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';

      for (var i = 0; i < 5; i++) {
        text += possible.charAt(Math.floor(Math.random() * possible.length));
      }

      return text;
    },
    isPM: function () {
      return this.$dtpElement.find('a.qor-meridien-pm').hasClass('selected');
    },
    setElementValue: function () {
      this.$element.trigger('beforeChange', this.currentDate);
      if (typeof ($.material) !== 'undefined') {
        this.$element.removeClass('empty');
      }
      this.$element.val(moment(this.currentDate).locale(this.params.lang).format(this.params.format));
      this.$element.trigger('change', this.currentDate);
    },
    toggleButtons: function (date) {
      if (date && date.isValid()) {
        var startOfMonth = moment(date).locale(this.params.lang).startOf('month');
        var endOfMonth = moment(date).locale(this.params.lang).endOf('month');

        if (!this.isAfterMinDate(startOfMonth, false, false)) {
          this.$dtpElement.find('a.qor-select-month-before').addClass('invisible');
        } else {
          this.$dtpElement.find('a.qor-select-month-before').removeClass('invisible');
        }

        if (!this.isBeforeMaxDate(endOfMonth, false, false)) {
          this.$dtpElement.find('a.qor-select-month-after').addClass('invisible');
        } else {
          this.$dtpElement.find('a.qor-select-month-after').removeClass('invisible');
        }

        var startOfYear = moment(date).locale(this.params.lang).startOf('year');
        var endOfYear = moment(date).locale(this.params.lang).endOf('year');

        if (!this.isAfterMinDate(startOfYear, false, false)) {
          this.$dtpElement.find('a.qor-select-year-before').addClass('invisible');
        } else {
          this.$dtpElement.find('a.qor-select-year-before').removeClass('invisible');
        }

        if (!this.isBeforeMaxDate(endOfYear, false, false)) {
          this.$dtpElement.find('a.qor-select-year-after').addClass('invisible');
        } else {
          this.$dtpElement.find('a.qor-select-year-after').removeClass('invisible');
        }
      }
    },
    toggleTime: function (isHours) {
      if (isHours) {
        this.$dtpElement.find('a.qor-select-hour').removeClass('disabled');
        this.$dtpElement.find('a.qor-select-hour').removeProp('disabled');
        this.$dtpElement.find('a.qor-select-hour').off('click');

        var _self = this;

        this.$dtpElement.find('a.qor-select-hour').each(function () {
          var _hour = $(this).data('hour');

          var _date = moment(_self.currentDate);
          _date.hour(_self.convertHours(_hour)).minute(0).second(0);

          if (_self.isAfterMinDate(_date, true, false) === false || _self.isBeforeMaxDate(_date, true, false) === false) {
            $(this).prop('disabled');
            $(this).addClass('disabled');
          } else {
            $(this).on('click', _self._onSelectHour.bind(_self));
          }
        });
      } else {
        this.$dtpElement.find('a.qor-select-minute').removeClass('disabled');
        this.$dtpElement.find('a.qor-select-minute').removeProp('disabled');
        this.$dtpElement.find('a.qor-select-minute').off('click');

        var _self = this;

        this.$dtpElement.find('a.qor-select-minute').each(function () {
          var _minute = $(this).data('minute');

          var _date = moment(_self.currentDate);
          _date.minute(_minute).second(0);

          if (_self.isAfterMinDate(_date, true, true) === false || _self.isBeforeMaxDate(_date, true, true) === false) {
            $(this).prop('disabled');
            $(this).addClass('disabled');
          } else {
            $(this).on('click', _self._onSelectMinute.bind(_self));
          }
        });
      }
    },
    _attachEvent: function (el, ev, fn) {
      el.on(ev, fn);
      this._attachedEvents.push([el, ev, fn]);
    },
    _detachEvents: function () {
      for (var i = this._attachedEvents.length - 1; i >= 0; i--) {
        this._attachedEvents[i][0].off(this._attachedEvents[i][1], this._attachedEvents[i][2]);
        this._attachedEvents.splice(i, 1);
      }
    },
    _onFocus: function () {
      this.currentView = 0;
      this.$element.blur();

      this.initDates();

      this.show();

      if (this.params.date) {
        this.$dtpElement.find('.qor-date').removeClass('hidden');
        this.initDate();
      } else {
        if (this.params.time) {
          this.$dtpElement.find('.qor-time').removeClass('hidden');
          this.initHours();
        }
      }
    },
    _onBackgroundClick: function (e) {
      e.stopPropagation();
      this.hide();
    },
    _onElementClick: function (e) {
      e.stopPropagation();
    },
    _onCloseClick: function () {
      this.hide();
    },
    _onOKClick: function () {
      switch (this.currentView) {
      case 0:
        if (this.params.time === true) {
          this.initHours();
        } else {
          this.setElementValue();
          this.hide();
        }
        break;
      case 1:
        this.initMinutes();
        break;
      case 2:
        this.setElementValue();
        this.hide();
        break;
      }
    },
    _onCancelClick: function () {
      if (this.params.time) {
        switch (this.currentView) {
        case 0:
          this.hide();
          break;
        case 1:
          if (this.params.date) {
            this.initDate();
          } else {
            this.hide();
          }
          break;
        case 2:
          this.initHours();
          break;
        }
      } else {
        this.hide();
      }
    },
    _onMonthBeforeClick: function () {
      this.currentDate.subtract(1, 'months');
      this.initDate(this.currentDate);
    },
    _onMonthAfterClick: function () {
      this.currentDate.add(1, 'months');
      this.initDate(this.currentDate);
    },
    _onYearBeforeClick: function () {
      this.currentDate.subtract(1, 'years');
      this.initDate(this.currentDate);
    },
    _onYearAfterClick: function () {
      this.currentDate.add(1, 'years');
      this.initDate(this.currentDate);
    },
    _onSelectDate: function (e) {
      this.$dtpElement.find('a.qor-select-day').removeClass('selected');
      $(e.currentTarget).addClass('selected');

      this.selectDate($(e.currentTarget).parent().data('date'));
    },
    _onSelectHour: function (e) {
      this.$dtpElement.find('a.qor-select-hour').removeClass('selected');
      $(e.currentTarget).addClass('selected');

      var dataHour = parseInt($(e.currentTarget).data('hour'));
      if (this.isPM())
        dataHour += 12;

      this.currentDate.hour(dataHour);
      this.showTime(this.currentDate);

      this.animateHands();
    },
    _onSelectMinute: function (e) {
      this.$dtpElement.find('a.qor-select-minute').removeClass('selected');
      $(e.currentTarget).addClass('selected');

      this.currentDate.minute(parseInt($(e.currentTarget).data('minute')));
      this.showTime(this.currentDate);

      this.animateHands();
    },
    _onSelectAM: function (e) {
      $('.qor-actual-meridien').find('a').removeClass('selected');
      $(e.currentTarget).addClass('selected');

      if (this.currentDate.hour() >= 12) {
        if (this.currentDate.subtract(12, 'hours'))
          this.showTime(this.currentDate);
      }
      this.toggleTime((this.currentView === 1));
    },
    _onSelectPM: function (e) {
      $('.qor-actual-meridien').find('a').removeClass('selected');
      $(e.currentTarget).addClass('selected');

      if (this.currentDate.hour() < 12) {
        if (this.currentDate.add(12, 'hours'))
          this.showTime(this.currentDate);
      }
      this.toggleTime((this.currentView === 1));
    },
    convertHours: function (h) {
      var _return = h;

      if ((h < 12) && this.isPM())
        _return += 12;

      return _return;
    },
    setDate: function (date) {
      this.params.currentDate = date;
      this.initDates();
    },
    setMinDate: function (date) {
      this.params.minDate = date;
      this.initDates();
    },
    setMaxDate: function (date) {
      this.params.maxDate = date;
      this.initDates();
    },
    destroy: function () {
      this._detachEvents();
      this.$dtpElement.remove();
    },
    show: function () {
      this.$dtpElement.removeClass('hidden');
      this._centerBox();
    },
    hide: function () {
      this.$dtpElement.addClass('hidden');
    },
    resetDate: function () {

    },
    _centerBox: function () {
      var h = (this.$dtpElement.height() - this.$dtpElement.find('.qor-datepicker__content').height()) / 2;
      this.$dtpElement.find('.qor-datepicker__content').css('marginLeft', -(this.$dtpElement.find('.qor-datepicker__content').width() / 2) + 'px');
      this.$dtpElement.find('.qor-datepicker__content').css('top', h + 'px');
    }
  };
})(jQuery, moment);
