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

  var location = window.location;

  var NAMESPACE = 'qor.i18n.inline';
  var EVENT_CLICK = 'click.' + NAMESPACE;
  var EVENT_CHANGE = 'change.' + NAMESPACE;

  // For Qor Autoheight plugin
  var EVENT_INPUT = 'input';

  function I18nInlineEdit(element, options) {
    this.$element = $(element);
    this.options = $.extend({}, I18nInlineEdit.DEFAULTS, $.isPlainObject(options) && options);
    this.multiple = false;
    this.init();
  }

  function encodeSearch(data) {
    var params = [];

    if ($.isPlainObject(data)) {
      $.each(data, function (name, value) {
        params.push([name, value].join('='));
      });
    }

    return params.join('&');
  }

  function decodeSearch(search) {
    var data = {};

    if (search) {
      search = search.replace('?', '').split('&');

      $.each(search, function (i, param) {
        param = param.split('=');
        i = param[0];
        data[i] = param[1];
      });
    }

    return data;
  }

  I18nInlineEdit.prototype = {
    contructor: I18nInlineEdit,

    init: function () {
      var $this = this.$element;
      this.makeInputEditable();
      this.bind();
    },

    bind: function () {
      this.$element.
        on(EVENT_CLICK, $.proxy(this.click, this)).
        on(EVENT_CHANGE, $.proxy(this.change, this));
    },

    unbind: function () {
      this.$element.
        off(EVENT_CLICK, this.click).
        off(EVENT_CHANGE, this.change);
    },

    makeInputEditable : function() {
      $.fn.editable.defaults.mode = 'popup';
      $.fn.editable.defaults.ajaxOptions = {type: "POST"};
      $(".qor-i18n-inline").editable({
        pk: 1,
        params: function(params) {
          params["Value"] = params.value;
          params["Locale"] = $(this).data("locale");
          params["Key"] = $(this).data("key");
          return params;
        },
        url: '/admin/translations'
      });
    }
  };

  I18nInlineEdit.DEFAULTS = {};

  I18nInlineEdit.plugin = function (options) {
    return this.each(function () {
      var $this = $(this);
      var data = $this.data(NAMESPACE);
      var fn;

      if (!data) {
        $this.data(NAMESPACE, (data = new I18nInlineEdit(this, options)));
      }

      if (typeof options === 'string' && $.isFunction((fn = data[options]))) {
        fn.apply(data);
      }
    });
  };

  $(function () {
    I18nInlineEdit.plugin.call($('.qor-i18n-inline'));
  });

});
