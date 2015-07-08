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

  var $document = $(document),
      FormData = window.FormData,

      NAMESPACE = 'qor.popper',
      EVENT_CLICK = 'click.' + NAMESPACE,
      EVENT_SUBMIT = 'submit.' + NAMESPACE,
      EVENT_SHOW = 'show.' + NAMESPACE,
      EVENT_SHOWN = 'shown.' + NAMESPACE,
      EVENT_HIDE = 'hide.' + NAMESPACE,
      EVENT_HIDDEN = 'hidden.' + NAMESPACE,

      QorPopper = function (element, options) {
        this.$element = $(element);
        this.options = $.extend({}, QorPopper.DEFAULTS, options);
        this.active = false;
        this.disabled = false;
        this.animating = false;
        this.init();
        console.log(this);
      };

  QorPopper.prototype = {
    constructor: QorPopper,

    init: function () {
      var $popper;

      this.$popper = $popper = $(QorPopper.TEMPLATE).appendTo('body');
      this.$title = $popper.find('.popper-title');
      this.$body = $popper.find('.popper-body');
      this.bind();
    },

    bind: function () {
      this.$popper.on(EVENT_SUBMIT, 'form', $.proxy(this.submit, this));
      $document.on(EVENT_CLICK, $.proxy(this.click, this));
    },

    unbind: function () {
      this.$popper.off(EVENT_SUBMIT, this.submit);
      $document.off(EVENT_CLICK, this.click);
    },

    click: function (e) {
      var $this = this.$element,
          popper = this.$popper.get(0),
          target = e.target,
          dismissible,
          $target,
          data;

      while (target !== document) {
        dismissible = false;
        $target = $(target);

        if (target === popper) {
          break;
        } else if ($target.data('dismiss') === 'popper') {
          this.hide();
          break;
        } else if ($target.is('tbody > tr')) {
          if (!$target.hasClass('active')) {
            $this.find('tbody > tr').removeClass('active');
            $target.addClass('active');
            this.load($target.find('.qor-action-edit').attr('href'));
          }

          break;
        } else if ($target.is('.qor-action-new')) {
          e.preventDefault();
          $this.find('tbody > tr').removeClass('active');
          this.load($target.attr('href'));
          break;
        } else if ($target.data('url')) {
          e.preventDefault();
          data = $target.data();
          this.load(data.url, data);
          break;
        } else {
          if ($target.is('.qor-action-edit') || $target.is('.qor-action-delete')) {
            e.preventDefault();
          }

          if (target) {
            dismissible = true;
            target = target.parentNode;
          } else {
            break;
          }
        }
      }

      if (dismissible) {
        $this.find('tbody > tr').removeClass('active');
        this.hide();
      }
    },

    submit: function (e) {
      var form = e.target,
          $form = $(form),
          _this = this;

      if (FormData) {
        e.preventDefault();

        $.ajax($form.prop('action'), {
          method: $form.prop('method'),
          data: new FormData(form),
          processData: false,
          contentType: false,
          success: function () {
            var returnUrl = $form.data('returnUrl');

            if (returnUrl) {
              _this.load(returnUrl);
              return;
            }

            _this.hide();

            setTimeout(function () {
              window.location.reload();
            }, 350);
          },
          error: function () {
            window.alert(arguments[1] + (arguments[2] || ''));
          }
        });
      }
    },

    load: function (url, options) {
      var _this = this,
          data = $.isPlainObject(options) ? options : {},
          method = data.method ? data.method : 'GET',
          load = function () {
            $.ajax(url, {
              method: method,
              data: data,
              success: function (response) {
                var $response,
                    $content;

                if (method === 'GET') {
                  $response = $(response);

                  if ($response.is('.qor-form-container')) {
                    $content = $response;
                  } else {
                    $content = $response.find('.qor-form-container');
                  }

                  $content.find('.qor-action-cancel').attr('data-dismiss', 'popper').removeAttr('href');
                  _this.$title.html($response.find('.qor-title').html());
                  _this.$body.html($content.html());
                  _this.$popper.one(EVENT_SHOWN, function () {
                    $(this).trigger('renew.qor.initiator'); // Renew Qor Components
                  });
                  _this.show();
                } else if (data.returnUrl) {
                  _this.disabled = false; // For reload
                  _this.load(data.returnUrl);
                }
              },
              complete: function () {
                _this.disabled = false;
              }
            });
          };

      if (!url || this.disabled) {
        return;
      }

      this.disabled = true;

      if (this.active) {
        this.hide();
        this.$popper.one(EVENT_HIDDEN, load);
      } else {
        load();
      }
    },

    show: function () {
      var $popper = this.$popper,
          showEvent;

      if (this.active || this.animating) {
        return;
      }

      showEvent = $.Event(EVENT_SHOW);
      $popper.trigger(showEvent);

      if (showEvent.isDefaultPrevented()) {
        return;
      }

      /*jshint expr:true */
      $popper.addClass('active').get(0).offsetWidth;
      $popper.addClass('in');
      this.animating = setTimeout($.proxy(this.shown, this), 350);
    },

    shown: function () {
      this.active = true;
      this.animating = false;
      this.$popper.trigger(EVENT_SHOWN);
    },

    hide: function () {
      var $popper = this.$popper,
          hideEvent;

      if (!this.active || this.animating) {
        return;
      }

      hideEvent = $.Event(EVENT_HIDE);
      $popper.trigger(hideEvent);

      if (hideEvent.isDefaultPrevented()) {
        return;
      }

      $popper.removeClass('in');
      this.animating = setTimeout($.proxy(this.hidden, this), 350);
    },

    hidden: function () {
      this.active = false;
      this.animating = false;
      this.$element.find('tbody > tr').removeClass('active');
      this.$popper.removeClass('active').trigger(EVENT_HIDDEN);
    },

    toggle: function () {
      if (this.active) {
        this.hide();
      } else {
        this.show();
      }
    },

    destroy: function () {
      this.unbind();
      this.$element.removeData(NAMESPACE);
    }
  };

  QorPopper.DEFAULTS = {
  };

  QorPopper.TEMPLATE = (
    '<div class="qor-popper">' +
      '<div class="popper-dialog">' +
        '<div class="popper-header">' +
          '<button type="button" class="popper-close" data-dismiss="popper" aria-div="Close"><span class="md md-24">close</span></button>' +
          '<h3 class="popper-title">Order Details</h3>' +
        '</div>' +
        '<div class="popper-body"></div>' +
        // '<div class="popper-footer"></div>' +
      '</div>' +
    '</div>'
  );

  QorPopper.plugin = function (options) {
    return this.each(function () {
      var $this = $(this);

      if (!$this.data(NAMESPACE)) {
        $this.data(NAMESPACE, new QorPopper(this, options));
      }
    });
  };

  $(function () {
    $(document)
      .on('renew.qor.initiator', function (e) {
        var $element = $('[data-toggle="qor.popper"]', e.target);

        if ($element.length) {
          QorPopper.plugin.call($element);
        }
      })
      .triggerHandler('renew.qor.initiator');
  });

  return QorPopper;

});
