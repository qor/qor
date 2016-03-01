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
  var NAMESPACE = 'qor.globalSearch';
  var EVENT_ENABLE = 'enable.' + NAMESPACE;
  var EVENT_DISABLE = 'disable.' + NAMESPACE;
  var EVENT_CLICK = 'click.' + NAMESPACE;

  function QorSearchCenter(element, options) {
    this.$element = $(element);
    this.options = $.extend({}, QorSearchCenter.DEFAULTS, $.isPlainObject(options) && options);
    this.init();
  }

  QorSearchCenter.prototype = {
    constructor: QorSearchCenter,

    init: function () {
      this.bind();
    },

    bind: function () {
      this.$element.on(EVENT_CLICK, $.proxy(this.click, this));
    },

    unbind: function () {
      this.$element.off(EVENT_CLICK, this.check);
    },

    click : function (e) {
      var $target = $(e.target);
    },

    showGlobalSearch: function (e) {
      var $target = $(e.target);
      var searchData = $target.data();

      console.log($target)
      console.log(searchData)

      $body.append(this.renderTmpl(searchData));

      return false;
    },

    submit: function () {
      var self = this;
      var $parent;


      $.ajax(properties.url, {
        method: properties.method,
        data: ajaxForm.formData,
        dataType: properties.datatype,

        success: function (data) {



        },
        error: function (xhr, textStatus, errorThrown) {
          self.$element.find(ACTION_BUTTON).attr('disabled', false);
          window.alert([textStatus, errorThrown].join(': '));
        }
      });
    },

    destroy: function () {
      this.unbind();
      this.$element.removeData(NAMESPACE);
    }

  };

  QorSearchCenter.DEFAULTS = {
  };

  QorSearchCenter.plugin = function (options) {
    return this.each(function () {
      var $this = $(this);
      var data = $this.data(NAMESPACE);
      var fn;

      if (!data) {
        $this.data(NAMESPACE, (data = new QorSearchCenter(this, options)));
      }

      if (typeof options === 'string' && $.isFunction(fn = data[options])) {
        fn.call(data);
      }
    });
  };

  $(function () {
    var selector = '[data-toggle="qor.global.search"]';
    var options = {};

    $(document).
      on(EVENT_DISABLE, function (e) {
        QorSearchCenter.plugin.call($(selector, e.target), 'destroy');
      }).
      on(EVENT_ENABLE, function (e) {
        QorSearchCenter.plugin.call($(selector, e.target), options);
      }).
      triggerHandler(EVENT_ENABLE);
  });

  return QorSearchCenter;

});
