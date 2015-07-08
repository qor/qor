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

  var FormData = window.FormData,
      NAMESPACE = 'qor.order',
      EVENT_CHANGE = 'change.' + NAMESPACE;

  function QorOrder(element, options) {
    this.$element = $(element);
    this.options = $.extend({}, QorOrder.DEFAULTS, $.isPlainObject(options) && options);
    this.submitting = false;
    this.init();
  }

  QorOrder.prototype = {
    constructor: QorOrder,

    init: function () {
      this.$form = this.$element.find('form');
      this.bind();
    },

    bind: function () {
      this.$element.on(EVENT_CHANGE, $.proxy(this.change, this));
    },

    unbind: function () {
      this.$element.off(EVENT_CHANGE, this.change);
    },

    change: function (e) {
      var $this = this.$element,
          $target = $(e.target),
          $form = this.$form,
          $relatedTarget,
          formData;

      if ($target.is(':checkbox')) {
        switch ($target.data('name')) {
          case 'toggle.fee':
            $relatedTarget = $form.find('[data-name="adjust.fee"]');
            break;

          case 'toggle.price':
            $relatedTarget = $form.find('[data-name="adjust.price"]');
            break;
        }

        if ($target.prop('checked')) {
          $relatedTarget.prop('disabled', false);
        } else {
          $relatedTarget.val(0).trigger(EVENT_CHANGE).prop('disabled', true);
        }

        return;
      }

      if (!FormData) {
        return;
      }

      formData = new FormData($form[0]);
      formData.append('NoSave', true);

      if (this.submitting) {
        clearTimeout(this.submitting);
      }

      this.submitting = setTimeout(function () {
        $.ajax($form.prop('action'), {
          method: $form.prop('method'),
          data: formData,
          dataType: 'json',
          processData: false,
          contentType: false,
          success: function (data) {
            var $field;

            if ($.isPlainObject(data)) {
              $field = $this.find('.detail-product-list');
              $field.find('[data-name="fee"]').text(data.ShippingFee);
              $field.find('[data-name="total"]').text(data.Total);
            }
          },
          complete: function () {
            this.submitting = false;
          }
        });
      }, 500);
    },

    destroy: function () {
      this.unbind();
      this.$element.removeData(NAMESPACE);
    }
  };

  QorOrder.DEFAULTS = {
  };

  QorOrder.plugin = function (options) {
    return this.each(function () {
      var $this = $(this),
          data = $this.data(NAMESPACE),
          fn;

      if (!data) {
        if (!$.fn.modal) {
          return;
        }

        $this.data(NAMESPACE, (data = new QorOrder(this, options)));
      }

      if (typeof options === 'string' && $.isFunction((fn = data[options]))) {
        fn.apply(data);
      }
    });
  };

  $(function () {
    $(document)
      .on('renew.qor.initiator', function (e) {
        var $element = $('[data-toggle="qor.order"]', e.target);

        if ($element.length) {
          QorOrder.plugin.call($element);
        }
      })
      .triggerHandler('renew.qor.initiator');
  });

});
