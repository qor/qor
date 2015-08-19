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
  var EVENT_KEYUP = 'keyup.' + NAMESPACE;
  var EVENT_CLICK = 'click.' + NAMESPACE;
  var EVENT_SUBMIT = 'submit.' + NAMESPACE;
  var EVENT_SHOW = 'show.' + NAMESPACE;
  var EVENT_SHOWN = 'shown.' + NAMESPACE;
  var EVENT_HIDE = 'hide.' + NAMESPACE;
  var EVENT_HIDDEN = 'hidden.' + NAMESPACE;
  var EVENT_TRANSITIONEND = 'transitionend';
  var CLASS_OPEN = 'qor-slideout-open';
  var CLASS_IS_SHOWN = 'is-shown';
  var CLASS_IS_SLIDED = 'is-slided';
  var CLASS_IS_SELECTED = 'is-selected';

  function QorSlideout(element, options) {
    this.$element = $(element);
    this.options = $.extend({}, QorSlideout.DEFAULTS, $.isPlainObject(options) && options);
    this.slided = false;
    this.disabled = false;
    this.init();
  }

  QorSlideout.prototype = {
    constructor: QorSlideout,

    init: function () {
      this.build();
      this.bind();
    },

    build: function () {
      var $slideout;

      this.$slideout = $slideout = $(QorSlideout.TEMPLATE).appendTo('body');
      this.$title = $slideout.find('.qor-slideout__title');
      this.$body = $slideout.find('.qor-slideout__body');
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
        on(EVENT_KEYUP, $.proxy(this.keyup, this)).
        on(EVENT_CLICK, $.proxy(this.click, this));
    },

    unbind: function () {
      this.$slideout.
        off(EVENT_SUBMIT, this.submit);

      $document.
        off(EVENT_KEYUP, this.keyup).
        off(EVENT_CLICK, this.click);
    },

    keyup: function (e) {
      if (e.which === 27) {
        this.hide();
      }
    },

    click: function (e) {
      var $this = this.$element;
      var slideout = this.$slideout.get(0);
      var target = e.target;
      var dismissible;
      var $target;
      var data;

      function toggleClass() {
        $this.find('tbody > tr').removeClass(CLASS_IS_SELECTED);
        $target.addClass(CLASS_IS_SELECTED);
      }

      if (e.isDefaultPrevented()) {
        return;
      }

      while (target !== document) {
        dismissible = false;
        $target = $(target);

        if ($target.prop('disabled')) {
          break;
        }

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
          if (!this.disabled && !$target.hasClass(CLASS_IS_SELECTED)) {
            $this.one(EVENT_SHOW, toggleClass);
            this.load($target.find('.qor-button--edit').attr('href'));
          }

          break;
        } else if ($target.is('.qor-button--new')) {
          e.preventDefault();
          this.load($target.attr('href'));
          break;
        } else {
          if ($target.is('.qor-button--edit') || $target.is('.qor-button--delete')) {
            e.preventDefault();
          } else if ($target.is('a')) {
            break;
          }

          if (target) {
            target = target.parentNode;
          } else {
            break;
          }
        }
      }
    },

    submit: function (e) {
      var $slideout = this.$slideout;
      var $body = this.$body;
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
            var $error;

            // Custom HTTP status code
            if (xhr.status === 422) {

              // Clear old errors
              $body.find('.qor-error').remove();
              $form.find('.qor-field').removeClass('is-error').find('.qor-field__error').remove();

              // Append new errors
              $error = $(xhr.responseText).find('.qor-error');
              $form.before($error);

              $error.find('> li > label').each(function () {
                var $label = $(this);
                var id = $label.attr('for');

                if (id) {
                  $form.find('#' + id).
                    closest('.qor-field').
                    addClass('is-error').
                    append($label.clone().addClass('qor-field__error'));
                }
              });

              // Scroll to top to view the errors
              $slideout.scrollTop(0);
            } else {
              window.alert([textStatus, errorThrown].join(': '));
            }
          },
          complete: function () {
            $submit.prop('disabled', false);
          },
        });
      }
    },

    load: function (url, data) {
      var options = this.options;
      var method;
      var load;

      if (!url || this.disabled) {
        return;
      }

      this.disabled = true;
      data = $.isPlainObject(data) ? data : {};
      method = data.method ? data.method : 'GET';

      load = $.proxy(function () {
        $.ajax(url, {
          method: method,
          data: data,
          success: $.proxy(function (response) {
            var $response;
            var $content;

            if (method === 'GET') {
              $response = $(response);

              if ($response.is(options.content)) {
                $content = $response;
              } else {
                $content = $response.find(options.content);
              }

              if (!$content.length) {
                return;
              }

              $content.find('.qor-button--cancel').attr('data-dismiss', 'slideout').removeAttr('href');
              this.$title.html($response.find(options.title).html());
              this.$body.html($content.html());

              this.$slideout.one(EVENT_SHOWN, function () {

                // Enable all Qor components within the slideout
                $(this).trigger('enable');
              }).one(EVENT_HIDDEN, function () {

                // Destroy all Qor components within the slideout
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

      if (this.slided) {
        this.hide();
        this.$slideout.one(EVENT_HIDDEN, load);
      } else {
        load();
      }
    },

    show: function () {
      var $slideout = this.$slideout;
      var showEvent;

      if (this.slided) {
        return;
      }

      showEvent = $.Event(EVENT_SHOW);
      $slideout.trigger(showEvent);

      if (showEvent.isDefaultPrevented()) {
        return;
      }

      /*jshint expr:true */
      $slideout.addClass(CLASS_IS_SHOWN).get(0).offsetWidth;
      $slideout.
        one(EVENT_TRANSITIONEND, $.proxy(this.shown, this)).
        addClass(CLASS_IS_SLIDED).
        scrollTop(0);
    },

    shown: function () {
      this.slided = true;

      // Disable to scroll body element
      $('body').addClass(CLASS_OPEN);

      this.$slideout.trigger(EVENT_SHOWN);
    },

    hide: function () {
      var $slideout = this.$slideout;
      var hideEvent;

      if (!this.slided) {
        return;
      }

      hideEvent = $.Event(EVENT_HIDE);
      $slideout.trigger(hideEvent);

      if (hideEvent.isDefaultPrevented()) {
        return;
      }

      $slideout.
        one(EVENT_TRANSITIONEND, $.proxy(this.hidden, this)).
        removeClass(CLASS_IS_SLIDED);
    },

    hidden: function () {
      this.slided = false;

      // Enable to scroll body element
      $('body').removeClass(CLASS_OPEN);

      this.$element.find('tbody > tr').removeClass(CLASS_IS_SELECTED);
      this.$slideout.removeClass(CLASS_IS_SHOWN).trigger(EVENT_HIDDEN);
    },

    refresh: function () {
      this.hide();

      setTimeout(function () {
        window.location.reload();
      }, 350);
    },

    toggle: function () {
      if (this.slided) {
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

  QorSlideout.DEFAULTS = {
    title: false,
    content: false,
  };

  QorSlideout.TEMPLATE = (
    '<div class="qor-slideout">' +
      '<div class="qor-slideout__header">' +
        '<button type="button" class="mdl-button mdl-button--icon mdl-js-button mdl-js-repple-effect qor-slideout__close" data-dismiss="slideout">' +
          '<span class="material-icons">close</span>' +
        '</button>' +
        '<h3 class="qor-slideout__title"></h3>' +
      '</div>' +
      '<div class="qor-slideout__body"></div>' +
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
    var options = {
          title: '.qor-form-title, .mdl-layout-title',
          content: '.qor-form-container',
        };

    $(document).
      on(EVENT_ENABLE, function (e) {

        if (/slideout/.test(e.namespace)) {
          QorSlideout.plugin.call($(selector, e.target), options);
        }
      }).
      on(EVENT_DISABLE, function (e) {

        if (/slideout/.test(e.namespace)) {
          QorSlideout.plugin.call($(selector, e.target), 'destroy');
        }
      }).
      triggerHandler(EVENT_ENABLE);
  });

  return QorSlideout;

});
