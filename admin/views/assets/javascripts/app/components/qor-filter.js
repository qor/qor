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
