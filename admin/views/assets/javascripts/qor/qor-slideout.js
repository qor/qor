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

  var $document = $(document);
  var FormData = window.FormData;
  var NAMESPACE = 'qor.slideout';
  var EVENT_ENABLE = 'enable.' + NAMESPACE;
  var EVENT_DISABLE = 'disable.' + NAMESPACE;
  var EVENT_CLICK = 'click.' + NAMESPACE;
  var EVENT_SUBMIT = 'submit.' + NAMESPACE;
  var EVENT_SHOW = 'show.' + NAMESPACE;
  var EVENT_SHOWN = 'shown.' + NAMESPACE;
  var EVENT_HIDE = 'hide.' + NAMESPACE;
  var EVENT_HIDDEN = 'hidden.' + NAMESPACE;
  var CLASS_OPEN = 'qor-slideout-open';

  function QorSlideout(element, options) {
    this.$element = $(element);
    this.options = $.extend({}, QorSlideout.DEFAULTS, $.isPlainObject(options) && options);
    this.active = false;
    this.disabled = false;
    this.animating = false;
    this.init();
  }

  QorSlideout.prototype = {
    constructor: QorSlideout,

    init: function () {
      if (!this.$element.find('.qor-list').length) {
        return;
      }

      this.build();
      this.bind();
    },

    build: function () {
      var $slideout;

      this.$slideout = $slideout = $(QorSlideout.TEMPLATE).appendTo('body');
      this.$title = $slideout.find('.slideout-title');
      this.$body = $slideout.find('.slideout-body');
      this.$documentBody = $('body');
    },

    unbuild: function () {
      this.$title = null;
      this.$body = null;
      this.$slideout.remove();
    },

    bind: function () {
      this.$slideout.
        on(EVENT_SUBMIT, 'form', $.proxy(this.submit, this));

      $document.
        on(EVENT_CLICK, $.proxy(this.click, this));
    },

    unbind: function () {
      this.$slideout.
        off(EVENT_SUBMIT, this.submit);

      $document.
        off(EVENT_CLICK, this.click);
    },

    click: function (e) {
      var $this = this.$element;
      var slideout = this.$slideout.get(0);
      var target = e.target;
      var dismissible;
      var $target;
      var data;

      function toggleClass() {
        $this.find('tbody > tr').removeClass('active');
        $target.addClass('active');
      }

      if (e.isDefaultPrevented()) {
        return;
      }

      while (target !== document) {
        dismissible = false;
        $target = $(target);

        if (target === slideout) {
          break;
        } else if ($target.data('url')) {
          e.preventDefault();
          data = $target.data();
          this.load(data.url, data);
          break;
        } else if ($target.data('dismiss') === 'slideout') {
          this.hide();
          break;
        } else if ($target.is('tbody > tr')) {
          if (!$target.hasClass('active')) {
            $this.one(EVENT_SHOW, toggleClass);
            this.load($target.find('.qor-action-edit').attr('href'));
          }

          break;
        } else if ($target.is('.qor-action-new')) {
          e.preventDefault();
          this.load($target.attr('href'));
          break;
        } else {
          if ($target.is('.qor-action-edit') || $target.is('.qor-action-delete')) {
            e.preventDefault();
          }

          if (target) {
            // dismissible = true;
            target = target.parentNode;
          } else {
            break;
          }
        }
      }

      // if (dismissible) {
      //   $this.find('tbody > tr').removeClass('active');
      //   this.hide();
      // }
    },

    submit: function (e) {
      var form = e.target;
      var $form = $(form);
      var _this = this;
      var $submit = $form.find(':submit');

      if (FormData) {
        e.preventDefault();

        $.ajax($form.prop('action'), {
          method: $form.prop('method'),
          data: new FormData(form),
          processData: false,
          contentType: false,
          beforeSend: function () {
            $submit.prop('disabled', true);
          },
          success: function () {
            var returnUrl = $form.data('returnUrl');

            if (returnUrl) {
              _this.load(returnUrl);
            } else {
              _this.refresh();
            }
          },
          error: function (xhr, textStatus, errorThrown) {
            if (xhr.status === 422) {
              $(xhr.responseText).find('.qor-error > li > label').each(function (i) {
                var $label = $(this);
                var $input = $form.find('#' + $label.attr('for'));

                if ($input.length) {
                  $input.after($label.clone().addClass('mdl-textfield__error'));
                  $input.closest('.form-group').addClass('has-error');

                  if (i === 0) {
                    $input.focus();
                  }
                }
              });
            } else {
              window.alert(textStatus + (errorThrown ? (' (' + (errorThrown || '') + ')') : ''));
            }
          },
          complete: function () {
            $submit.prop('disabled', false);
          },
        });
      }
    },

    load: function (url, options) {
      var data = $.isPlainObject(options) ? options : {};
      var method = data.method ? data.method : 'GET';
      var load;

      if (!url || this.disabled) {
        return;
      }

      this.disabled = true;

      load = $.proxy(function () {
        $.ajax(url, {
          method: method,
          data: data,
          success: $.proxy(function (response) {
            var $response;
            var $content;

            if (method === 'GET') {
              $response = $(response);

              if ($response.is('.qor-form-container')) {
                $content = $response;
              } else {
                $content = $response.find('.qor-form-container');
              }

              $content.find('.qor-action-cancel').attr('data-dismiss', 'slideout').removeAttr('href');
              this.$title.html($response.find('.qor-title').html());
              this.$body.empty().html($content.html());
              this.$slideout.one(EVENT_SHOWN, function () {

                // Enable all JavaScript components within the slideout
                $(this).trigger('enable');
              }).one(EVENT_HIDDEN, function () {

                // Destroy all JavaScript components within the slideout
                $(this).trigger('disable');
              });

              this.show();
            } else {
              if (data.returnUrl) {
                this.disabled = false; // For reload
                this.load(data.returnUrl);
              } else {
                this.refresh();
              }
            }
          }, this),
          complete: $.proxy(function () {
            this.disabled = false;
          }, this),
        });
      }, this);

      if (this.active) {
        this.hide();
        this.$slideout.one(EVENT_HIDDEN, load);
      } else {
        load();
      }
    },

    show: function () {
      var $slideout = this.$slideout;
      var showEvent;

      if (this.active || this.animating) {
        return;
      }

      showEvent = $.Event(EVENT_SHOW);
      $slideout.trigger(showEvent);

      if (showEvent.isDefaultPrevented()) {
        return;
      }

      /*jshint expr:true */
      $slideout.addClass('active').get(0).offsetWidth;
      $slideout.addClass('in');
      this.animating = setTimeout($.proxy(this.shown, this), 350);
    },

    shown: function () {
      this.active = true;
      this.animating = false;
      this.$slideout.trigger(EVENT_SHOWN);

      // Disable to scroll body element
      this.$documentBody.addClass(CLASS_OPEN);
    },

    hide: function () {
      var $slideout = this.$slideout;
      var hideEvent;

      if (!this.active || this.animating) {
        return;
      }

      hideEvent = $.Event(EVENT_HIDE);
      $slideout.trigger(hideEvent);

      if (hideEvent.isDefaultPrevented()) {
        return;
      }

      $slideout.removeClass('in');
      this.animating = setTimeout($.proxy(this.hidden, this), 350);
    },

    hidden: function () {
      this.active = false;
      this.animating = false;
      this.$element.find('tbody > tr').removeClass('active');
      this.$slideout.removeClass('active').trigger(EVENT_HIDDEN);

      // Enable to scroll body element
      this.$documentBody.removeClass(CLASS_OPEN);
    },

    refresh: function () {
      this.hide();

      setTimeout(function () {
        window.location.reload();
      }, 350);
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
      this.unbuild();
      this.$element.removeData(NAMESPACE);
    },
  };

  QorSlideout.DEFAULTS = {};

  QorSlideout.TEMPLATE = (
    '<div class="qor-slideout">' +
      '<div class="slideout-dialog">' +
        '<div class="slideout-header">' +
          '<button type="button" class="slideout-close" data-dismiss="slideout" aria-div="Close">' +
            '<span class="material-icons">close</span>' +
          '</button>' +
          '<h3 class="slideout-title"></h3>' +
        '</div>' +
        '<div class="slideout-body"></div>' +
      '</div>' +
    '</div>'
  );

  QorSlideout.plugin = function (options) {
    return this.each(function () {
      var $this = $(this);
      var data = $this.data(NAMESPACE);
      var fn;

      if (!data) {
        if (/destroy/.test(options)) {
          return;
        }

        $this.data(NAMESPACE, (data = new QorSlideout(this, options)));
      }

      if (typeof options === 'string' && $.isFunction(fn = data[options])) {
        fn.apply(data);
      }
    });
  };

  $(function () {
    var selector = '.qor-theme-slideout';

    $(document)
      .on(EVENT_DISABLE, function (e) {

        if (/slideout/.test(e.namespace)) {
          QorSlideout.plugin.call($(selector, e.target), 'destroy');
        }
      })
      .on(EVENT_ENABLE, function (e) {

        if (/slideout/.test(e.namespace)) {
          QorSlideout.plugin.call($(selector, e.target));
        }
      })
      .triggerHandler(EVENT_ENABLE);
  });

  return QorSlideout;

});
