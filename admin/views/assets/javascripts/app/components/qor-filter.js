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

  function encodeSearch(data, detched) {
    var search = location.search,
        params;

    if ($.isArray(data)) {
      params = decodeSearch(search);

      $.each(data, function (i, param) {
        i = $.inArray(param, params);

        if (i === -1) {
          params.push(param);
        } else if (detched) {
          params.splice(i, 1);
        }
      });

      search = '?' + params.join('&');
    }

    return search;
  }

  function decodeSearch(search) {
    var data = [];

    if (search && search.indexOf('?') > -1) {
      search = search.split('?')[1];

      if (search && search.indexOf('#') > -1) {
        search = search.split('#')[0];
      }

      if (search) {
        // search = search.toLowerCase();
        data = search.split('&');
      }
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
          params = decodeSearch(location.search);

      this.$toggle.each(function () {
        var $this = $(this);

        $.each(decodeSearch($this.attr('href')), function (i, param) {
          var matched = $.inArray(param, params) > -1;

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
