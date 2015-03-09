/*
 * jQuery Fugue Plugin
 * For Ajax Form submit
 * Copyright (c) 2014 Lancee (xrhy.me)
 * Dual licensed under the MIT and GPL licenses
 */

!(function() {
  (function($, Export) {
    "use strict";

    $.fugue = function(ajaxform, options) {
      if (!ajaxform || ajaxform.nodeName.toLowerCase()!=="form") {
        throw new Error('this is not a form');
      }

      $.support.formdata = Export.FormData !== undefined;

      var Fugue = function() {
        this.init();
      };

      Fugue.prototype = {
        constructor: Fugue,

        init: function() {
          var $form = $(ajaxform).data('fugue', this),
              me = this;

          me.$el = $form.addClass('fugue');

          options = $.extend({}, $.fugue.defaults, options);
          options.type = $form.attr('method') || options.type;
          options.url = $form.attr('action') || options.url;

          $form.on('submit', function(e, func) {
            e = e.originalEvent || e;
            (e.preventDefault) ? e.preventDefault() : e.returnValue = false;

            if (!navigator.onLine) {
              alert('Connection Dead... Please check out the network ;)');
              return;
            }
            me.submit(e, func);
          });

          me.options = options;
        },

        serialize: function() {
          if (options.beforeSerialize && $.isFunction(options.beforeSerialize)) {
            options.beforeSerialize.call(this);
          }
          if ($.support.formdata) {
            var dataArray = this.$el.serializeArray(),
                data = {};

            $.each(dataArray, function(i, field) {
              if (field.value === 'true' || field.value === 'false') {
                var isTrue = (field.value === 'true');
                data[field.name] = isTrue;
              } else {
                data[field.name] ? data[field.name] = [field.value].concat(data[field.name]) : data[field.name] = field.value;
              }
            });
          }
          options.data = data;
          this.options = options;
        },

        submit: function(e, func) {
          var me = this;
          this.serialize();
          if (options.beforeSubmit && $.isFunction(options.beforeSubmit)) {
            options.beforeSubmit.call(this);
          }
          if (this.cancel) {
            return
          }
          this.defered = $.ajax(this.options).done(function(data, textStatus, jqXHR) {
            if (options.done && $.isFunction(options.done)) {
              options.done.call(me, data, textStatus, jqXHR);
            }

            if (func && $.isFunction(func)) {
              func.call(me, data, textStatus, jqXHR);
            }

            if (options.reset && $.isFunction(options.reset)) {
              options.reset.call(me, data, textStatus, jqXHR);
            }
          }).fail(function(jqXHR, textStatus, errorThrown) {
            if (options.fail && $.isFunction(options.fail)) {
              options.fail.call(me, jqXHR, textStatus, errorThrown);
            }
          });
        },
        options: $.fugue.defaults
      }

      return new Fugue();

    }

    $.fugue.defaults = {
      beforeSerialize: function() {},
      beforeSubmit: function() {},
      done: function() {},
      fail: function() {},
      reset: function() {},
      type: 'POST',
      contentType: 'application/json; charset=UTF-8',
      dataType: 'json',
      cache: false,
      timeout: 7777
    };

    $.fn.fugue = function(options, callback) {
      var fugue = $(this).data('fugue');

      if ($.isFunction(options)) {
        callback = options;
        options = null;
      } else {
        options = options || {};
      }

      if(typeof(options) === 'object') {
        return this.each(function(i) {
          if(!fugue) {
            fugue = $.fugue(this, options);
            if(callback)
              callback.call(fugue);
          } else {
            if(callback)
              callback.call(fugue);
          }
        });
      } else {
        throw new Error('arguments[0] is not a instance of Object');
      }
    }

  })(jQuery, window);

}).call(this);