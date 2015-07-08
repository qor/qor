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

  var NAMESPACE = 'qor.alert',
      EVENT_CLICK = 'click.' + NAMESPACE;

  $(function () {
    $(document).on(EVENT_CLICK, '[data-dismiss="alert"]', function () {
      $(this).closest('.qor-alert').remove();
    });

    setTimeout(function () {
      $('.qor-alert[data-dismissible="true"]').remove();
    }, 3000);
  });

});

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

  var NAMESPACE = 'qor.textviewer',
      EVENT_CLICK = 'click.' + NAMESPACE;

  function TextViewer(element, options) {
    this.$element = $(element);
    this.options = $.extend({}, TextViewer.DEFAULTS, $.isPlainObject(options) && options);
    this.$modal = null;
    this.built = false;
    this.init();
  }

  TextViewer.prototype = {
    constructor: TextViewer,

    init: function () {
      this.$element.find(this.options.toggle).each(function () {
        var $this = $(this);

        if (this.scrollHeight > $this.height()) {
          $this.addClass('active').wrapInner(TextViewer.INNER);
        }
      });
      this.bind();
    },

    build: function () {
      if (this.built) {
        return;
      }

      this.built = true;
      this.$modal = $(TextViewer.TEMPLATE).modal({
        show: false
      }).appendTo('body');
    },

    bind: function () {
      this.$element.on(EVENT_CLICK, this.options.toggle, $.proxy(this.click, this));
    },

    unbind: function () {
      this.$element.off(EVENT_CLICK, this.click);
    },

    click: function (e) {
      var target = e.currentTarget,
          $target = $(target),
          $modal;

      if (!this.built) {
        this.build();
      }

      if ($target.hasClass('active')) {
        $modal = this.$modal;
        $modal.find('.modal-title').text($target.closest('td').attr('title'));
        $modal.find('.modal-body').html($target.html());
        $modal.modal('show');
      }
    },

    destroy: function () {
      this.unbind();
      this.$element.removeData(NAMESPACE);
    }
  };

  TextViewer.DEFAULTS = {
    toggle: '.qor-list-text'
  };

  TextViewer.INNER = ('<div class="text-inner"></div>');

  TextViewer.TEMPLATE = (
    '<div class="modal fade qor-list-modal" id="qorListModal" tabindex="-1" role="dialog" aria-labelledby="qorListModalLabel" aria-hidden="true">' +
      '<div class="modal-dialog">' +
        '<div class="modal-content">' +
          '<div class="modal-header">' +
            '<button type="button" class="close" data-dismiss="modal" aria-label="Close"><span aria-hidden="true">&times;</span></button>' +
            '<h4 class="modal-title" id="qorPublishModalLabel"></h4>' +
          '</div>' +
          '<div class="modal-body"></div>' +
        '</div>' +
      '</div>' +
    '</div>'
  );

  TextViewer.plugin = function (options) {
    return this.each(function () {
      var $this = $(this),
          data = $this.data(NAMESPACE),
          fn;

      if (!data) {
        if (!$.fn.modal) {
          return;
        }

        $this.data(NAMESPACE, (data = new TextViewer(this, options)));
      }

      if (typeof options === 'string' && $.isFunction((fn = data[options]))) {
        fn.apply(data);
      }
    });
  };

  $(function () {
    TextViewer.plugin.call($('.qor-list'));
  });

});

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

$(function () {

  'use strict';

  // Add Bootstrap's classes dynamically
  $('.qor-locale-selector').on('change', function () {
    var url = $(this).val();

    if (url) {
      window.location.assign(url);
    }
  });

  // Toggle submenus
  $('.qor-menu-group').on('click', '> ul > li > a', function () {
    var $next = $(this).next();

    if ($next.is('ul') && $next.css('position') !== 'absolute') {
      if (!$next.hasClass('collapsable')) {
        $next.addClass('collapsable').height($next.prop('scrollHeight'));
      }

      if ($next.hasClass('collapsed')) {
        $next.height($next.prop('scrollHeight'));

        setTimeout(function () {
          $next.removeClass('collapsed');
        }, 350);
      } else {
        $next.addClass('collapsed').height(0);
      }
    }
  });

  $('.qor-search').each(function () {
    var $this = $(this),
        $label = $this.find('.qor-search-label'),
        $input = $this.find('.qor-search-input'),
        $clear = $this.find('.qor-search-clear');

    $label.on('click', function () {
      if (!$input.hasClass('focus')) {
        $this.addClass('active');
        $input.addClass('focus');
      }
    });

    $clear.on('click', function () {
      if ($input.val()) {
        $input.val('');
      } else {
        $this.removeClass('active');
        $input.removeClass('focus');
      }
    });

  });

  // Init Bootstrap Material Design
  $.material.init();
});

//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJzb3VyY2VzIjpbInFvci1hbGVydC5qcyIsInFvci1wb3BwZXIuanMiLCJxb3ItdGV4dHZpZXdlci5qcyIsInFvci1vcmRlci5qcyIsInFvci5qcyJdLCJuYW1lcyI6W10sIm1hcHBpbmdzIjoiQUFBQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUM3QkE7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FDcFRBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUMvSEE7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQ2pKQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBIiwiZmlsZSI6InFvci5qcyIsInNvdXJjZXNDb250ZW50IjpbIihmdW5jdGlvbiAoZmFjdG9yeSkge1xuICBpZiAodHlwZW9mIGRlZmluZSA9PT0gJ2Z1bmN0aW9uJyAmJiBkZWZpbmUuYW1kKSB7XG4gICAgLy8gQU1ELiBSZWdpc3RlciBhcyBhbm9ueW1vdXMgbW9kdWxlLlxuICAgIGRlZmluZShbJ2pxdWVyeSddLCBmYWN0b3J5KTtcbiAgfSBlbHNlIGlmICh0eXBlb2YgZXhwb3J0cyA9PT0gJ29iamVjdCcpIHtcbiAgICAvLyBOb2RlIC8gQ29tbW9uSlNcbiAgICBmYWN0b3J5KHJlcXVpcmUoJ2pxdWVyeScpKTtcbiAgfSBlbHNlIHtcbiAgICAvLyBCcm93c2VyIGdsb2JhbHMuXG4gICAgZmFjdG9yeShqUXVlcnkpO1xuICB9XG59KShmdW5jdGlvbiAoJCkge1xuXG4gICd1c2Ugc3RyaWN0JztcblxuICB2YXIgTkFNRVNQQUNFID0gJ3Fvci5hbGVydCcsXG4gICAgICBFVkVOVF9DTElDSyA9ICdjbGljay4nICsgTkFNRVNQQUNFO1xuXG4gICQoZnVuY3Rpb24gKCkge1xuICAgICQoZG9jdW1lbnQpLm9uKEVWRU5UX0NMSUNLLCAnW2RhdGEtZGlzbWlzcz1cImFsZXJ0XCJdJywgZnVuY3Rpb24gKCkge1xuICAgICAgJCh0aGlzKS5jbG9zZXN0KCcucW9yLWFsZXJ0JykucmVtb3ZlKCk7XG4gICAgfSk7XG5cbiAgICBzZXRUaW1lb3V0KGZ1bmN0aW9uICgpIHtcbiAgICAgICQoJy5xb3ItYWxlcnRbZGF0YS1kaXNtaXNzaWJsZT1cInRydWVcIl0nKS5yZW1vdmUoKTtcbiAgICB9LCAzMDAwKTtcbiAgfSk7XG5cbn0pO1xuIiwiKGZ1bmN0aW9uIChmYWN0b3J5KSB7XG4gIGlmICh0eXBlb2YgZGVmaW5lID09PSAnZnVuY3Rpb24nICYmIGRlZmluZS5hbWQpIHtcbiAgICAvLyBBTUQuIFJlZ2lzdGVyIGFzIGFub255bW91cyBtb2R1bGUuXG4gICAgZGVmaW5lKFsnanF1ZXJ5J10sIGZhY3RvcnkpO1xuICB9IGVsc2UgaWYgKHR5cGVvZiBleHBvcnRzID09PSAnb2JqZWN0Jykge1xuICAgIC8vIE5vZGUgLyBDb21tb25KU1xuICAgIGZhY3RvcnkocmVxdWlyZSgnanF1ZXJ5JykpO1xuICB9IGVsc2Uge1xuICAgIC8vIEJyb3dzZXIgZ2xvYmFscy5cbiAgICBmYWN0b3J5KGpRdWVyeSk7XG4gIH1cbn0pKGZ1bmN0aW9uICgkKSB7XG5cbiAgJ3VzZSBzdHJpY3QnO1xuXG4gIHZhciAkZG9jdW1lbnQgPSAkKGRvY3VtZW50KSxcbiAgICAgIEZvcm1EYXRhID0gd2luZG93LkZvcm1EYXRhLFxuXG4gICAgICBOQU1FU1BBQ0UgPSAncW9yLnBvcHBlcicsXG4gICAgICBFVkVOVF9DTElDSyA9ICdjbGljay4nICsgTkFNRVNQQUNFLFxuICAgICAgRVZFTlRfU1VCTUlUID0gJ3N1Ym1pdC4nICsgTkFNRVNQQUNFLFxuICAgICAgRVZFTlRfU0hPVyA9ICdzaG93LicgKyBOQU1FU1BBQ0UsXG4gICAgICBFVkVOVF9TSE9XTiA9ICdzaG93bi4nICsgTkFNRVNQQUNFLFxuICAgICAgRVZFTlRfSElERSA9ICdoaWRlLicgKyBOQU1FU1BBQ0UsXG4gICAgICBFVkVOVF9ISURERU4gPSAnaGlkZGVuLicgKyBOQU1FU1BBQ0UsXG5cbiAgICAgIFFvclBvcHBlciA9IGZ1bmN0aW9uIChlbGVtZW50LCBvcHRpb25zKSB7XG4gICAgICAgIHRoaXMuJGVsZW1lbnQgPSAkKGVsZW1lbnQpO1xuICAgICAgICB0aGlzLm9wdGlvbnMgPSAkLmV4dGVuZCh7fSwgUW9yUG9wcGVyLkRFRkFVTFRTLCBvcHRpb25zKTtcbiAgICAgICAgdGhpcy5hY3RpdmUgPSBmYWxzZTtcbiAgICAgICAgdGhpcy5kaXNhYmxlZCA9IGZhbHNlO1xuICAgICAgICB0aGlzLmFuaW1hdGluZyA9IGZhbHNlO1xuICAgICAgICB0aGlzLmluaXQoKTtcbiAgICAgICAgY29uc29sZS5sb2codGhpcyk7XG4gICAgICB9O1xuXG4gIFFvclBvcHBlci5wcm90b3R5cGUgPSB7XG4gICAgY29uc3RydWN0b3I6IFFvclBvcHBlcixcblxuICAgIGluaXQ6IGZ1bmN0aW9uICgpIHtcbiAgICAgIHZhciAkcG9wcGVyO1xuXG4gICAgICB0aGlzLiRwb3BwZXIgPSAkcG9wcGVyID0gJChRb3JQb3BwZXIuVEVNUExBVEUpLmFwcGVuZFRvKCdib2R5Jyk7XG4gICAgICB0aGlzLiR0aXRsZSA9ICRwb3BwZXIuZmluZCgnLnBvcHBlci10aXRsZScpO1xuICAgICAgdGhpcy4kYm9keSA9ICRwb3BwZXIuZmluZCgnLnBvcHBlci1ib2R5Jyk7XG4gICAgICB0aGlzLmJpbmQoKTtcbiAgICB9LFxuXG4gICAgYmluZDogZnVuY3Rpb24gKCkge1xuICAgICAgdGhpcy4kcG9wcGVyLm9uKEVWRU5UX1NVQk1JVCwgJ2Zvcm0nLCAkLnByb3h5KHRoaXMuc3VibWl0LCB0aGlzKSk7XG4gICAgICAkZG9jdW1lbnQub24oRVZFTlRfQ0xJQ0ssICQucHJveHkodGhpcy5jbGljaywgdGhpcykpO1xuICAgIH0sXG5cbiAgICB1bmJpbmQ6IGZ1bmN0aW9uICgpIHtcbiAgICAgIHRoaXMuJHBvcHBlci5vZmYoRVZFTlRfU1VCTUlULCB0aGlzLnN1Ym1pdCk7XG4gICAgICAkZG9jdW1lbnQub2ZmKEVWRU5UX0NMSUNLLCB0aGlzLmNsaWNrKTtcbiAgICB9LFxuXG4gICAgY2xpY2s6IGZ1bmN0aW9uIChlKSB7XG4gICAgICB2YXIgJHRoaXMgPSB0aGlzLiRlbGVtZW50LFxuICAgICAgICAgIHBvcHBlciA9IHRoaXMuJHBvcHBlci5nZXQoMCksXG4gICAgICAgICAgdGFyZ2V0ID0gZS50YXJnZXQsXG4gICAgICAgICAgZGlzbWlzc2libGUsXG4gICAgICAgICAgJHRhcmdldCxcbiAgICAgICAgICBkYXRhO1xuXG4gICAgICB3aGlsZSAodGFyZ2V0ICE9PSBkb2N1bWVudCkge1xuICAgICAgICBkaXNtaXNzaWJsZSA9IGZhbHNlO1xuICAgICAgICAkdGFyZ2V0ID0gJCh0YXJnZXQpO1xuXG4gICAgICAgIGlmICh0YXJnZXQgPT09IHBvcHBlcikge1xuICAgICAgICAgIGJyZWFrO1xuICAgICAgICB9IGVsc2UgaWYgKCR0YXJnZXQuZGF0YSgnZGlzbWlzcycpID09PSAncG9wcGVyJykge1xuICAgICAgICAgIHRoaXMuaGlkZSgpO1xuICAgICAgICAgIGJyZWFrO1xuICAgICAgICB9IGVsc2UgaWYgKCR0YXJnZXQuaXMoJ3Rib2R5ID4gdHInKSkge1xuICAgICAgICAgIGlmICghJHRhcmdldC5oYXNDbGFzcygnYWN0aXZlJykpIHtcbiAgICAgICAgICAgICR0aGlzLmZpbmQoJ3Rib2R5ID4gdHInKS5yZW1vdmVDbGFzcygnYWN0aXZlJyk7XG4gICAgICAgICAgICAkdGFyZ2V0LmFkZENsYXNzKCdhY3RpdmUnKTtcbiAgICAgICAgICAgIHRoaXMubG9hZCgkdGFyZ2V0LmZpbmQoJy5xb3ItYWN0aW9uLWVkaXQnKS5hdHRyKCdocmVmJykpO1xuICAgICAgICAgIH1cblxuICAgICAgICAgIGJyZWFrO1xuICAgICAgICB9IGVsc2UgaWYgKCR0YXJnZXQuaXMoJy5xb3ItYWN0aW9uLW5ldycpKSB7XG4gICAgICAgICAgZS5wcmV2ZW50RGVmYXVsdCgpO1xuICAgICAgICAgICR0aGlzLmZpbmQoJ3Rib2R5ID4gdHInKS5yZW1vdmVDbGFzcygnYWN0aXZlJyk7XG4gICAgICAgICAgdGhpcy5sb2FkKCR0YXJnZXQuYXR0cignaHJlZicpKTtcbiAgICAgICAgICBicmVhaztcbiAgICAgICAgfSBlbHNlIGlmICgkdGFyZ2V0LmRhdGEoJ3VybCcpKSB7XG4gICAgICAgICAgZS5wcmV2ZW50RGVmYXVsdCgpO1xuICAgICAgICAgIGRhdGEgPSAkdGFyZ2V0LmRhdGEoKTtcbiAgICAgICAgICB0aGlzLmxvYWQoZGF0YS51cmwsIGRhdGEpO1xuICAgICAgICAgIGJyZWFrO1xuICAgICAgICB9IGVsc2Uge1xuICAgICAgICAgIGlmICgkdGFyZ2V0LmlzKCcucW9yLWFjdGlvbi1lZGl0JykgfHwgJHRhcmdldC5pcygnLnFvci1hY3Rpb24tZGVsZXRlJykpIHtcbiAgICAgICAgICAgIGUucHJldmVudERlZmF1bHQoKTtcbiAgICAgICAgICB9XG5cbiAgICAgICAgICBpZiAodGFyZ2V0KSB7XG4gICAgICAgICAgICBkaXNtaXNzaWJsZSA9IHRydWU7XG4gICAgICAgICAgICB0YXJnZXQgPSB0YXJnZXQucGFyZW50Tm9kZTtcbiAgICAgICAgICB9IGVsc2Uge1xuICAgICAgICAgICAgYnJlYWs7XG4gICAgICAgICAgfVxuICAgICAgICB9XG4gICAgICB9XG5cbiAgICAgIGlmIChkaXNtaXNzaWJsZSkge1xuICAgICAgICAkdGhpcy5maW5kKCd0Ym9keSA+IHRyJykucmVtb3ZlQ2xhc3MoJ2FjdGl2ZScpO1xuICAgICAgICB0aGlzLmhpZGUoKTtcbiAgICAgIH1cbiAgICB9LFxuXG4gICAgc3VibWl0OiBmdW5jdGlvbiAoZSkge1xuICAgICAgdmFyIGZvcm0gPSBlLnRhcmdldCxcbiAgICAgICAgICAkZm9ybSA9ICQoZm9ybSksXG4gICAgICAgICAgX3RoaXMgPSB0aGlzO1xuXG4gICAgICBpZiAoRm9ybURhdGEpIHtcbiAgICAgICAgZS5wcmV2ZW50RGVmYXVsdCgpO1xuXG4gICAgICAgICQuYWpheCgkZm9ybS5wcm9wKCdhY3Rpb24nKSwge1xuICAgICAgICAgIG1ldGhvZDogJGZvcm0ucHJvcCgnbWV0aG9kJyksXG4gICAgICAgICAgZGF0YTogbmV3IEZvcm1EYXRhKGZvcm0pLFxuICAgICAgICAgIHByb2Nlc3NEYXRhOiBmYWxzZSxcbiAgICAgICAgICBjb250ZW50VHlwZTogZmFsc2UsXG4gICAgICAgICAgc3VjY2VzczogZnVuY3Rpb24gKCkge1xuICAgICAgICAgICAgdmFyIHJldHVyblVybCA9ICRmb3JtLmRhdGEoJ3JldHVyblVybCcpO1xuXG4gICAgICAgICAgICBpZiAocmV0dXJuVXJsKSB7XG4gICAgICAgICAgICAgIF90aGlzLmxvYWQocmV0dXJuVXJsKTtcbiAgICAgICAgICAgICAgcmV0dXJuO1xuICAgICAgICAgICAgfVxuXG4gICAgICAgICAgICBfdGhpcy5oaWRlKCk7XG5cbiAgICAgICAgICAgIHNldFRpbWVvdXQoZnVuY3Rpb24gKCkge1xuICAgICAgICAgICAgICB3aW5kb3cubG9jYXRpb24ucmVsb2FkKCk7XG4gICAgICAgICAgICB9LCAzNTApO1xuICAgICAgICAgIH0sXG4gICAgICAgICAgZXJyb3I6IGZ1bmN0aW9uICgpIHtcbiAgICAgICAgICAgIHdpbmRvdy5hbGVydChhcmd1bWVudHNbMV0gKyAoYXJndW1lbnRzWzJdIHx8ICcnKSk7XG4gICAgICAgICAgfVxuICAgICAgICB9KTtcbiAgICAgIH1cbiAgICB9LFxuXG4gICAgbG9hZDogZnVuY3Rpb24gKHVybCwgb3B0aW9ucykge1xuICAgICAgdmFyIF90aGlzID0gdGhpcyxcbiAgICAgICAgICBkYXRhID0gJC5pc1BsYWluT2JqZWN0KG9wdGlvbnMpID8gb3B0aW9ucyA6IHt9LFxuICAgICAgICAgIG1ldGhvZCA9IGRhdGEubWV0aG9kID8gZGF0YS5tZXRob2QgOiAnR0VUJyxcbiAgICAgICAgICBsb2FkID0gZnVuY3Rpb24gKCkge1xuICAgICAgICAgICAgJC5hamF4KHVybCwge1xuICAgICAgICAgICAgICBtZXRob2Q6IG1ldGhvZCxcbiAgICAgICAgICAgICAgZGF0YTogZGF0YSxcbiAgICAgICAgICAgICAgc3VjY2VzczogZnVuY3Rpb24gKHJlc3BvbnNlKSB7XG4gICAgICAgICAgICAgICAgdmFyICRyZXNwb25zZSxcbiAgICAgICAgICAgICAgICAgICAgJGNvbnRlbnQ7XG5cbiAgICAgICAgICAgICAgICBpZiAobWV0aG9kID09PSAnR0VUJykge1xuICAgICAgICAgICAgICAgICAgJHJlc3BvbnNlID0gJChyZXNwb25zZSk7XG5cbiAgICAgICAgICAgICAgICAgIGlmICgkcmVzcG9uc2UuaXMoJy5xb3ItZm9ybS1jb250YWluZXInKSkge1xuICAgICAgICAgICAgICAgICAgICAkY29udGVudCA9ICRyZXNwb25zZTtcbiAgICAgICAgICAgICAgICAgIH0gZWxzZSB7XG4gICAgICAgICAgICAgICAgICAgICRjb250ZW50ID0gJHJlc3BvbnNlLmZpbmQoJy5xb3ItZm9ybS1jb250YWluZXInKTtcbiAgICAgICAgICAgICAgICAgIH1cblxuICAgICAgICAgICAgICAgICAgJGNvbnRlbnQuZmluZCgnLnFvci1hY3Rpb24tY2FuY2VsJykuYXR0cignZGF0YS1kaXNtaXNzJywgJ3BvcHBlcicpLnJlbW92ZUF0dHIoJ2hyZWYnKTtcbiAgICAgICAgICAgICAgICAgIF90aGlzLiR0aXRsZS5odG1sKCRyZXNwb25zZS5maW5kKCcucW9yLXRpdGxlJykuaHRtbCgpKTtcbiAgICAgICAgICAgICAgICAgIF90aGlzLiRib2R5Lmh0bWwoJGNvbnRlbnQuaHRtbCgpKTtcbiAgICAgICAgICAgICAgICAgIF90aGlzLiRwb3BwZXIub25lKEVWRU5UX1NIT1dOLCBmdW5jdGlvbiAoKSB7XG4gICAgICAgICAgICAgICAgICAgICQodGhpcykudHJpZ2dlcigncmVuZXcucW9yLmluaXRpYXRvcicpOyAvLyBSZW5ldyBRb3IgQ29tcG9uZW50c1xuICAgICAgICAgICAgICAgICAgfSk7XG4gICAgICAgICAgICAgICAgICBfdGhpcy5zaG93KCk7XG4gICAgICAgICAgICAgICAgfSBlbHNlIGlmIChkYXRhLnJldHVyblVybCkge1xuICAgICAgICAgICAgICAgICAgX3RoaXMuZGlzYWJsZWQgPSBmYWxzZTsgLy8gRm9yIHJlbG9hZFxuICAgICAgICAgICAgICAgICAgX3RoaXMubG9hZChkYXRhLnJldHVyblVybCk7XG4gICAgICAgICAgICAgICAgfVxuICAgICAgICAgICAgICB9LFxuICAgICAgICAgICAgICBjb21wbGV0ZTogZnVuY3Rpb24gKCkge1xuICAgICAgICAgICAgICAgIF90aGlzLmRpc2FibGVkID0gZmFsc2U7XG4gICAgICAgICAgICAgIH1cbiAgICAgICAgICAgIH0pO1xuICAgICAgICAgIH07XG5cbiAgICAgIGlmICghdXJsIHx8IHRoaXMuZGlzYWJsZWQpIHtcbiAgICAgICAgcmV0dXJuO1xuICAgICAgfVxuXG4gICAgICB0aGlzLmRpc2FibGVkID0gdHJ1ZTtcblxuICAgICAgaWYgKHRoaXMuYWN0aXZlKSB7XG4gICAgICAgIHRoaXMuaGlkZSgpO1xuICAgICAgICB0aGlzLiRwb3BwZXIub25lKEVWRU5UX0hJRERFTiwgbG9hZCk7XG4gICAgICB9IGVsc2Uge1xuICAgICAgICBsb2FkKCk7XG4gICAgICB9XG4gICAgfSxcblxuICAgIHNob3c6IGZ1bmN0aW9uICgpIHtcbiAgICAgIHZhciAkcG9wcGVyID0gdGhpcy4kcG9wcGVyLFxuICAgICAgICAgIHNob3dFdmVudDtcblxuICAgICAgaWYgKHRoaXMuYWN0aXZlIHx8IHRoaXMuYW5pbWF0aW5nKSB7XG4gICAgICAgIHJldHVybjtcbiAgICAgIH1cblxuICAgICAgc2hvd0V2ZW50ID0gJC5FdmVudChFVkVOVF9TSE9XKTtcbiAgICAgICRwb3BwZXIudHJpZ2dlcihzaG93RXZlbnQpO1xuXG4gICAgICBpZiAoc2hvd0V2ZW50LmlzRGVmYXVsdFByZXZlbnRlZCgpKSB7XG4gICAgICAgIHJldHVybjtcbiAgICAgIH1cblxuICAgICAgLypqc2hpbnQgZXhwcjp0cnVlICovXG4gICAgICAkcG9wcGVyLmFkZENsYXNzKCdhY3RpdmUnKS5nZXQoMCkub2Zmc2V0V2lkdGg7XG4gICAgICAkcG9wcGVyLmFkZENsYXNzKCdpbicpO1xuICAgICAgdGhpcy5hbmltYXRpbmcgPSBzZXRUaW1lb3V0KCQucHJveHkodGhpcy5zaG93biwgdGhpcyksIDM1MCk7XG4gICAgfSxcblxuICAgIHNob3duOiBmdW5jdGlvbiAoKSB7XG4gICAgICB0aGlzLmFjdGl2ZSA9IHRydWU7XG4gICAgICB0aGlzLmFuaW1hdGluZyA9IGZhbHNlO1xuICAgICAgdGhpcy4kcG9wcGVyLnRyaWdnZXIoRVZFTlRfU0hPV04pO1xuICAgIH0sXG5cbiAgICBoaWRlOiBmdW5jdGlvbiAoKSB7XG4gICAgICB2YXIgJHBvcHBlciA9IHRoaXMuJHBvcHBlcixcbiAgICAgICAgICBoaWRlRXZlbnQ7XG5cbiAgICAgIGlmICghdGhpcy5hY3RpdmUgfHwgdGhpcy5hbmltYXRpbmcpIHtcbiAgICAgICAgcmV0dXJuO1xuICAgICAgfVxuXG4gICAgICBoaWRlRXZlbnQgPSAkLkV2ZW50KEVWRU5UX0hJREUpO1xuICAgICAgJHBvcHBlci50cmlnZ2VyKGhpZGVFdmVudCk7XG5cbiAgICAgIGlmIChoaWRlRXZlbnQuaXNEZWZhdWx0UHJldmVudGVkKCkpIHtcbiAgICAgICAgcmV0dXJuO1xuICAgICAgfVxuXG4gICAgICAkcG9wcGVyLnJlbW92ZUNsYXNzKCdpbicpO1xuICAgICAgdGhpcy5hbmltYXRpbmcgPSBzZXRUaW1lb3V0KCQucHJveHkodGhpcy5oaWRkZW4sIHRoaXMpLCAzNTApO1xuICAgIH0sXG5cbiAgICBoaWRkZW46IGZ1bmN0aW9uICgpIHtcbiAgICAgIHRoaXMuYWN0aXZlID0gZmFsc2U7XG4gICAgICB0aGlzLmFuaW1hdGluZyA9IGZhbHNlO1xuICAgICAgdGhpcy4kZWxlbWVudC5maW5kKCd0Ym9keSA+IHRyJykucmVtb3ZlQ2xhc3MoJ2FjdGl2ZScpO1xuICAgICAgdGhpcy4kcG9wcGVyLnJlbW92ZUNsYXNzKCdhY3RpdmUnKS50cmlnZ2VyKEVWRU5UX0hJRERFTik7XG4gICAgfSxcblxuICAgIHRvZ2dsZTogZnVuY3Rpb24gKCkge1xuICAgICAgaWYgKHRoaXMuYWN0aXZlKSB7XG4gICAgICAgIHRoaXMuaGlkZSgpO1xuICAgICAgfSBlbHNlIHtcbiAgICAgICAgdGhpcy5zaG93KCk7XG4gICAgICB9XG4gICAgfSxcblxuICAgIGRlc3Ryb3k6IGZ1bmN0aW9uICgpIHtcbiAgICAgIHRoaXMudW5iaW5kKCk7XG4gICAgICB0aGlzLiRlbGVtZW50LnJlbW92ZURhdGEoTkFNRVNQQUNFKTtcbiAgICB9XG4gIH07XG5cbiAgUW9yUG9wcGVyLkRFRkFVTFRTID0ge1xuICB9O1xuXG4gIFFvclBvcHBlci5URU1QTEFURSA9IChcbiAgICAnPGRpdiBjbGFzcz1cInFvci1wb3BwZXJcIj4nICtcbiAgICAgICc8ZGl2IGNsYXNzPVwicG9wcGVyLWRpYWxvZ1wiPicgK1xuICAgICAgICAnPGRpdiBjbGFzcz1cInBvcHBlci1oZWFkZXJcIj4nICtcbiAgICAgICAgICAnPGJ1dHRvbiB0eXBlPVwiYnV0dG9uXCIgY2xhc3M9XCJwb3BwZXItY2xvc2VcIiBkYXRhLWRpc21pc3M9XCJwb3BwZXJcIiBhcmlhLWRpdj1cIkNsb3NlXCI+PHNwYW4gY2xhc3M9XCJtZCBtZC0yNFwiPmNsb3NlPC9zcGFuPjwvYnV0dG9uPicgK1xuICAgICAgICAgICc8aDMgY2xhc3M9XCJwb3BwZXItdGl0bGVcIj5PcmRlciBEZXRhaWxzPC9oMz4nICtcbiAgICAgICAgJzwvZGl2PicgK1xuICAgICAgICAnPGRpdiBjbGFzcz1cInBvcHBlci1ib2R5XCI+PC9kaXY+JyArXG4gICAgICAgIC8vICc8ZGl2IGNsYXNzPVwicG9wcGVyLWZvb3RlclwiPjwvZGl2PicgK1xuICAgICAgJzwvZGl2PicgK1xuICAgICc8L2Rpdj4nXG4gICk7XG5cbiAgUW9yUG9wcGVyLnBsdWdpbiA9IGZ1bmN0aW9uIChvcHRpb25zKSB7XG4gICAgcmV0dXJuIHRoaXMuZWFjaChmdW5jdGlvbiAoKSB7XG4gICAgICB2YXIgJHRoaXMgPSAkKHRoaXMpO1xuXG4gICAgICBpZiAoISR0aGlzLmRhdGEoTkFNRVNQQUNFKSkge1xuICAgICAgICAkdGhpcy5kYXRhKE5BTUVTUEFDRSwgbmV3IFFvclBvcHBlcih0aGlzLCBvcHRpb25zKSk7XG4gICAgICB9XG4gICAgfSk7XG4gIH07XG5cbiAgJChmdW5jdGlvbiAoKSB7XG4gICAgJChkb2N1bWVudClcbiAgICAgIC5vbigncmVuZXcucW9yLmluaXRpYXRvcicsIGZ1bmN0aW9uIChlKSB7XG4gICAgICAgIHZhciAkZWxlbWVudCA9ICQoJ1tkYXRhLXRvZ2dsZT1cInFvci5wb3BwZXJcIl0nLCBlLnRhcmdldCk7XG5cbiAgICAgICAgaWYgKCRlbGVtZW50Lmxlbmd0aCkge1xuICAgICAgICAgIFFvclBvcHBlci5wbHVnaW4uY2FsbCgkZWxlbWVudCk7XG4gICAgICAgIH1cbiAgICAgIH0pXG4gICAgICAudHJpZ2dlckhhbmRsZXIoJ3JlbmV3LnFvci5pbml0aWF0b3InKTtcbiAgfSk7XG5cbiAgcmV0dXJuIFFvclBvcHBlcjtcblxufSk7XG4iLCIoZnVuY3Rpb24gKGZhY3RvcnkpIHtcbiAgaWYgKHR5cGVvZiBkZWZpbmUgPT09ICdmdW5jdGlvbicgJiYgZGVmaW5lLmFtZCkge1xuICAgIC8vIEFNRC4gUmVnaXN0ZXIgYXMgYW5vbnltb3VzIG1vZHVsZS5cbiAgICBkZWZpbmUoWydqcXVlcnknXSwgZmFjdG9yeSk7XG4gIH0gZWxzZSBpZiAodHlwZW9mIGV4cG9ydHMgPT09ICdvYmplY3QnKSB7XG4gICAgLy8gTm9kZSAvIENvbW1vbkpTXG4gICAgZmFjdG9yeShyZXF1aXJlKCdqcXVlcnknKSk7XG4gIH0gZWxzZSB7XG4gICAgLy8gQnJvd3NlciBnbG9iYWxzLlxuICAgIGZhY3RvcnkoalF1ZXJ5KTtcbiAgfVxufSkoZnVuY3Rpb24gKCQpIHtcblxuICAndXNlIHN0cmljdCc7XG5cbiAgdmFyIE5BTUVTUEFDRSA9ICdxb3IudGV4dHZpZXdlcicsXG4gICAgICBFVkVOVF9DTElDSyA9ICdjbGljay4nICsgTkFNRVNQQUNFO1xuXG4gIGZ1bmN0aW9uIFRleHRWaWV3ZXIoZWxlbWVudCwgb3B0aW9ucykge1xuICAgIHRoaXMuJGVsZW1lbnQgPSAkKGVsZW1lbnQpO1xuICAgIHRoaXMub3B0aW9ucyA9ICQuZXh0ZW5kKHt9LCBUZXh0Vmlld2VyLkRFRkFVTFRTLCAkLmlzUGxhaW5PYmplY3Qob3B0aW9ucykgJiYgb3B0aW9ucyk7XG4gICAgdGhpcy4kbW9kYWwgPSBudWxsO1xuICAgIHRoaXMuYnVpbHQgPSBmYWxzZTtcbiAgICB0aGlzLmluaXQoKTtcbiAgfVxuXG4gIFRleHRWaWV3ZXIucHJvdG90eXBlID0ge1xuICAgIGNvbnN0cnVjdG9yOiBUZXh0Vmlld2VyLFxuXG4gICAgaW5pdDogZnVuY3Rpb24gKCkge1xuICAgICAgdGhpcy4kZWxlbWVudC5maW5kKHRoaXMub3B0aW9ucy50b2dnbGUpLmVhY2goZnVuY3Rpb24gKCkge1xuICAgICAgICB2YXIgJHRoaXMgPSAkKHRoaXMpO1xuXG4gICAgICAgIGlmICh0aGlzLnNjcm9sbEhlaWdodCA+ICR0aGlzLmhlaWdodCgpKSB7XG4gICAgICAgICAgJHRoaXMuYWRkQ2xhc3MoJ2FjdGl2ZScpLndyYXBJbm5lcihUZXh0Vmlld2VyLklOTkVSKTtcbiAgICAgICAgfVxuICAgICAgfSk7XG4gICAgICB0aGlzLmJpbmQoKTtcbiAgICB9LFxuXG4gICAgYnVpbGQ6IGZ1bmN0aW9uICgpIHtcbiAgICAgIGlmICh0aGlzLmJ1aWx0KSB7XG4gICAgICAgIHJldHVybjtcbiAgICAgIH1cblxuICAgICAgdGhpcy5idWlsdCA9IHRydWU7XG4gICAgICB0aGlzLiRtb2RhbCA9ICQoVGV4dFZpZXdlci5URU1QTEFURSkubW9kYWwoe1xuICAgICAgICBzaG93OiBmYWxzZVxuICAgICAgfSkuYXBwZW5kVG8oJ2JvZHknKTtcbiAgICB9LFxuXG4gICAgYmluZDogZnVuY3Rpb24gKCkge1xuICAgICAgdGhpcy4kZWxlbWVudC5vbihFVkVOVF9DTElDSywgdGhpcy5vcHRpb25zLnRvZ2dsZSwgJC5wcm94eSh0aGlzLmNsaWNrLCB0aGlzKSk7XG4gICAgfSxcblxuICAgIHVuYmluZDogZnVuY3Rpb24gKCkge1xuICAgICAgdGhpcy4kZWxlbWVudC5vZmYoRVZFTlRfQ0xJQ0ssIHRoaXMuY2xpY2spO1xuICAgIH0sXG5cbiAgICBjbGljazogZnVuY3Rpb24gKGUpIHtcbiAgICAgIHZhciB0YXJnZXQgPSBlLmN1cnJlbnRUYXJnZXQsXG4gICAgICAgICAgJHRhcmdldCA9ICQodGFyZ2V0KSxcbiAgICAgICAgICAkbW9kYWw7XG5cbiAgICAgIGlmICghdGhpcy5idWlsdCkge1xuICAgICAgICB0aGlzLmJ1aWxkKCk7XG4gICAgICB9XG5cbiAgICAgIGlmICgkdGFyZ2V0Lmhhc0NsYXNzKCdhY3RpdmUnKSkge1xuICAgICAgICAkbW9kYWwgPSB0aGlzLiRtb2RhbDtcbiAgICAgICAgJG1vZGFsLmZpbmQoJy5tb2RhbC10aXRsZScpLnRleHQoJHRhcmdldC5jbG9zZXN0KCd0ZCcpLmF0dHIoJ3RpdGxlJykpO1xuICAgICAgICAkbW9kYWwuZmluZCgnLm1vZGFsLWJvZHknKS5odG1sKCR0YXJnZXQuaHRtbCgpKTtcbiAgICAgICAgJG1vZGFsLm1vZGFsKCdzaG93Jyk7XG4gICAgICB9XG4gICAgfSxcblxuICAgIGRlc3Ryb3k6IGZ1bmN0aW9uICgpIHtcbiAgICAgIHRoaXMudW5iaW5kKCk7XG4gICAgICB0aGlzLiRlbGVtZW50LnJlbW92ZURhdGEoTkFNRVNQQUNFKTtcbiAgICB9XG4gIH07XG5cbiAgVGV4dFZpZXdlci5ERUZBVUxUUyA9IHtcbiAgICB0b2dnbGU6ICcucW9yLWxpc3QtdGV4dCdcbiAgfTtcblxuICBUZXh0Vmlld2VyLklOTkVSID0gKCc8ZGl2IGNsYXNzPVwidGV4dC1pbm5lclwiPjwvZGl2PicpO1xuXG4gIFRleHRWaWV3ZXIuVEVNUExBVEUgPSAoXG4gICAgJzxkaXYgY2xhc3M9XCJtb2RhbCBmYWRlIHFvci1saXN0LW1vZGFsXCIgaWQ9XCJxb3JMaXN0TW9kYWxcIiB0YWJpbmRleD1cIi0xXCIgcm9sZT1cImRpYWxvZ1wiIGFyaWEtbGFiZWxsZWRieT1cInFvckxpc3RNb2RhbExhYmVsXCIgYXJpYS1oaWRkZW49XCJ0cnVlXCI+JyArXG4gICAgICAnPGRpdiBjbGFzcz1cIm1vZGFsLWRpYWxvZ1wiPicgK1xuICAgICAgICAnPGRpdiBjbGFzcz1cIm1vZGFsLWNvbnRlbnRcIj4nICtcbiAgICAgICAgICAnPGRpdiBjbGFzcz1cIm1vZGFsLWhlYWRlclwiPicgK1xuICAgICAgICAgICAgJzxidXR0b24gdHlwZT1cImJ1dHRvblwiIGNsYXNzPVwiY2xvc2VcIiBkYXRhLWRpc21pc3M9XCJtb2RhbFwiIGFyaWEtbGFiZWw9XCJDbG9zZVwiPjxzcGFuIGFyaWEtaGlkZGVuPVwidHJ1ZVwiPiZ0aW1lczs8L3NwYW4+PC9idXR0b24+JyArXG4gICAgICAgICAgICAnPGg0IGNsYXNzPVwibW9kYWwtdGl0bGVcIiBpZD1cInFvclB1Ymxpc2hNb2RhbExhYmVsXCI+PC9oND4nICtcbiAgICAgICAgICAnPC9kaXY+JyArXG4gICAgICAgICAgJzxkaXYgY2xhc3M9XCJtb2RhbC1ib2R5XCI+PC9kaXY+JyArXG4gICAgICAgICc8L2Rpdj4nICtcbiAgICAgICc8L2Rpdj4nICtcbiAgICAnPC9kaXY+J1xuICApO1xuXG4gIFRleHRWaWV3ZXIucGx1Z2luID0gZnVuY3Rpb24gKG9wdGlvbnMpIHtcbiAgICByZXR1cm4gdGhpcy5lYWNoKGZ1bmN0aW9uICgpIHtcbiAgICAgIHZhciAkdGhpcyA9ICQodGhpcyksXG4gICAgICAgICAgZGF0YSA9ICR0aGlzLmRhdGEoTkFNRVNQQUNFKSxcbiAgICAgICAgICBmbjtcblxuICAgICAgaWYgKCFkYXRhKSB7XG4gICAgICAgIGlmICghJC5mbi5tb2RhbCkge1xuICAgICAgICAgIHJldHVybjtcbiAgICAgICAgfVxuXG4gICAgICAgICR0aGlzLmRhdGEoTkFNRVNQQUNFLCAoZGF0YSA9IG5ldyBUZXh0Vmlld2VyKHRoaXMsIG9wdGlvbnMpKSk7XG4gICAgICB9XG5cbiAgICAgIGlmICh0eXBlb2Ygb3B0aW9ucyA9PT0gJ3N0cmluZycgJiYgJC5pc0Z1bmN0aW9uKChmbiA9IGRhdGFbb3B0aW9uc10pKSkge1xuICAgICAgICBmbi5hcHBseShkYXRhKTtcbiAgICAgIH1cbiAgICB9KTtcbiAgfTtcblxuICAkKGZ1bmN0aW9uICgpIHtcbiAgICBUZXh0Vmlld2VyLnBsdWdpbi5jYWxsKCQoJy5xb3ItbGlzdCcpKTtcbiAgfSk7XG5cbn0pO1xuIiwiKGZ1bmN0aW9uIChmYWN0b3J5KSB7XG4gIGlmICh0eXBlb2YgZGVmaW5lID09PSAnZnVuY3Rpb24nICYmIGRlZmluZS5hbWQpIHtcbiAgICAvLyBBTUQuIFJlZ2lzdGVyIGFzIGFub255bW91cyBtb2R1bGUuXG4gICAgZGVmaW5lKFsnanF1ZXJ5J10sIGZhY3RvcnkpO1xuICB9IGVsc2UgaWYgKHR5cGVvZiBleHBvcnRzID09PSAnb2JqZWN0Jykge1xuICAgIC8vIE5vZGUgLyBDb21tb25KU1xuICAgIGZhY3RvcnkocmVxdWlyZSgnanF1ZXJ5JykpO1xuICB9IGVsc2Uge1xuICAgIC8vIEJyb3dzZXIgZ2xvYmFscy5cbiAgICBmYWN0b3J5KGpRdWVyeSk7XG4gIH1cbn0pKGZ1bmN0aW9uICgkKSB7XG5cbiAgJ3VzZSBzdHJpY3QnO1xuXG4gIHZhciBGb3JtRGF0YSA9IHdpbmRvdy5Gb3JtRGF0YSxcbiAgICAgIE5BTUVTUEFDRSA9ICdxb3Iub3JkZXInLFxuICAgICAgRVZFTlRfQ0hBTkdFID0gJ2NoYW5nZS4nICsgTkFNRVNQQUNFO1xuXG4gIGZ1bmN0aW9uIFFvck9yZGVyKGVsZW1lbnQsIG9wdGlvbnMpIHtcbiAgICB0aGlzLiRlbGVtZW50ID0gJChlbGVtZW50KTtcbiAgICB0aGlzLm9wdGlvbnMgPSAkLmV4dGVuZCh7fSwgUW9yT3JkZXIuREVGQVVMVFMsICQuaXNQbGFpbk9iamVjdChvcHRpb25zKSAmJiBvcHRpb25zKTtcbiAgICB0aGlzLnN1Ym1pdHRpbmcgPSBmYWxzZTtcbiAgICB0aGlzLmluaXQoKTtcbiAgfVxuXG4gIFFvck9yZGVyLnByb3RvdHlwZSA9IHtcbiAgICBjb25zdHJ1Y3RvcjogUW9yT3JkZXIsXG5cbiAgICBpbml0OiBmdW5jdGlvbiAoKSB7XG4gICAgICB0aGlzLiRmb3JtID0gdGhpcy4kZWxlbWVudC5maW5kKCdmb3JtJyk7XG4gICAgICB0aGlzLmJpbmQoKTtcbiAgICB9LFxuXG4gICAgYmluZDogZnVuY3Rpb24gKCkge1xuICAgICAgdGhpcy4kZWxlbWVudC5vbihFVkVOVF9DSEFOR0UsICQucHJveHkodGhpcy5jaGFuZ2UsIHRoaXMpKTtcbiAgICB9LFxuXG4gICAgdW5iaW5kOiBmdW5jdGlvbiAoKSB7XG4gICAgICB0aGlzLiRlbGVtZW50Lm9mZihFVkVOVF9DSEFOR0UsIHRoaXMuY2hhbmdlKTtcbiAgICB9LFxuXG4gICAgY2hhbmdlOiBmdW5jdGlvbiAoZSkge1xuICAgICAgdmFyICR0aGlzID0gdGhpcy4kZWxlbWVudCxcbiAgICAgICAgICAkdGFyZ2V0ID0gJChlLnRhcmdldCksXG4gICAgICAgICAgJGZvcm0gPSB0aGlzLiRmb3JtLFxuICAgICAgICAgICRyZWxhdGVkVGFyZ2V0LFxuICAgICAgICAgIGZvcm1EYXRhO1xuXG4gICAgICBpZiAoJHRhcmdldC5pcygnOmNoZWNrYm94JykpIHtcbiAgICAgICAgc3dpdGNoICgkdGFyZ2V0LmRhdGEoJ25hbWUnKSkge1xuICAgICAgICAgIGNhc2UgJ3RvZ2dsZS5mZWUnOlxuICAgICAgICAgICAgJHJlbGF0ZWRUYXJnZXQgPSAkZm9ybS5maW5kKCdbZGF0YS1uYW1lPVwiYWRqdXN0LmZlZVwiXScpO1xuICAgICAgICAgICAgYnJlYWs7XG5cbiAgICAgICAgICBjYXNlICd0b2dnbGUucHJpY2UnOlxuICAgICAgICAgICAgJHJlbGF0ZWRUYXJnZXQgPSAkZm9ybS5maW5kKCdbZGF0YS1uYW1lPVwiYWRqdXN0LnByaWNlXCJdJyk7XG4gICAgICAgICAgICBicmVhaztcbiAgICAgICAgfVxuXG4gICAgICAgIGlmICgkdGFyZ2V0LnByb3AoJ2NoZWNrZWQnKSkge1xuICAgICAgICAgICRyZWxhdGVkVGFyZ2V0LnByb3AoJ2Rpc2FibGVkJywgZmFsc2UpO1xuICAgICAgICB9IGVsc2Uge1xuICAgICAgICAgICRyZWxhdGVkVGFyZ2V0LnZhbCgwKS50cmlnZ2VyKEVWRU5UX0NIQU5HRSkucHJvcCgnZGlzYWJsZWQnLCB0cnVlKTtcbiAgICAgICAgfVxuXG4gICAgICAgIHJldHVybjtcbiAgICAgIH1cblxuICAgICAgaWYgKCFGb3JtRGF0YSkge1xuICAgICAgICByZXR1cm47XG4gICAgICB9XG5cbiAgICAgIGZvcm1EYXRhID0gbmV3IEZvcm1EYXRhKCRmb3JtWzBdKTtcbiAgICAgIGZvcm1EYXRhLmFwcGVuZCgnTm9TYXZlJywgdHJ1ZSk7XG5cbiAgICAgIGlmICh0aGlzLnN1Ym1pdHRpbmcpIHtcbiAgICAgICAgY2xlYXJUaW1lb3V0KHRoaXMuc3VibWl0dGluZyk7XG4gICAgICB9XG5cbiAgICAgIHRoaXMuc3VibWl0dGluZyA9IHNldFRpbWVvdXQoZnVuY3Rpb24gKCkge1xuICAgICAgICAkLmFqYXgoJGZvcm0ucHJvcCgnYWN0aW9uJyksIHtcbiAgICAgICAgICBtZXRob2Q6ICRmb3JtLnByb3AoJ21ldGhvZCcpLFxuICAgICAgICAgIGRhdGE6IGZvcm1EYXRhLFxuICAgICAgICAgIGRhdGFUeXBlOiAnanNvbicsXG4gICAgICAgICAgcHJvY2Vzc0RhdGE6IGZhbHNlLFxuICAgICAgICAgIGNvbnRlbnRUeXBlOiBmYWxzZSxcbiAgICAgICAgICBzdWNjZXNzOiBmdW5jdGlvbiAoZGF0YSkge1xuICAgICAgICAgICAgdmFyICRmaWVsZDtcblxuICAgICAgICAgICAgaWYgKCQuaXNQbGFpbk9iamVjdChkYXRhKSkge1xuICAgICAgICAgICAgICAkZmllbGQgPSAkdGhpcy5maW5kKCcuZGV0YWlsLXByb2R1Y3QtbGlzdCcpO1xuICAgICAgICAgICAgICAkZmllbGQuZmluZCgnW2RhdGEtbmFtZT1cImZlZVwiXScpLnRleHQoZGF0YS5TaGlwcGluZ0ZlZSk7XG4gICAgICAgICAgICAgICRmaWVsZC5maW5kKCdbZGF0YS1uYW1lPVwidG90YWxcIl0nKS50ZXh0KGRhdGEuVG90YWwpO1xuICAgICAgICAgICAgfVxuICAgICAgICAgIH0sXG4gICAgICAgICAgY29tcGxldGU6IGZ1bmN0aW9uICgpIHtcbiAgICAgICAgICAgIHRoaXMuc3VibWl0dGluZyA9IGZhbHNlO1xuICAgICAgICAgIH1cbiAgICAgICAgfSk7XG4gICAgICB9LCA1MDApO1xuICAgIH0sXG5cbiAgICBkZXN0cm95OiBmdW5jdGlvbiAoKSB7XG4gICAgICB0aGlzLnVuYmluZCgpO1xuICAgICAgdGhpcy4kZWxlbWVudC5yZW1vdmVEYXRhKE5BTUVTUEFDRSk7XG4gICAgfVxuICB9O1xuXG4gIFFvck9yZGVyLkRFRkFVTFRTID0ge1xuICB9O1xuXG4gIFFvck9yZGVyLnBsdWdpbiA9IGZ1bmN0aW9uIChvcHRpb25zKSB7XG4gICAgcmV0dXJuIHRoaXMuZWFjaChmdW5jdGlvbiAoKSB7XG4gICAgICB2YXIgJHRoaXMgPSAkKHRoaXMpLFxuICAgICAgICAgIGRhdGEgPSAkdGhpcy5kYXRhKE5BTUVTUEFDRSksXG4gICAgICAgICAgZm47XG5cbiAgICAgIGlmICghZGF0YSkge1xuICAgICAgICBpZiAoISQuZm4ubW9kYWwpIHtcbiAgICAgICAgICByZXR1cm47XG4gICAgICAgIH1cblxuICAgICAgICAkdGhpcy5kYXRhKE5BTUVTUEFDRSwgKGRhdGEgPSBuZXcgUW9yT3JkZXIodGhpcywgb3B0aW9ucykpKTtcbiAgICAgIH1cblxuICAgICAgaWYgKHR5cGVvZiBvcHRpb25zID09PSAnc3RyaW5nJyAmJiAkLmlzRnVuY3Rpb24oKGZuID0gZGF0YVtvcHRpb25zXSkpKSB7XG4gICAgICAgIGZuLmFwcGx5KGRhdGEpO1xuICAgICAgfVxuICAgIH0pO1xuICB9O1xuXG4gICQoZnVuY3Rpb24gKCkge1xuICAgICQoZG9jdW1lbnQpXG4gICAgICAub24oJ3JlbmV3LnFvci5pbml0aWF0b3InLCBmdW5jdGlvbiAoZSkge1xuICAgICAgICB2YXIgJGVsZW1lbnQgPSAkKCdbZGF0YS10b2dnbGU9XCJxb3Iub3JkZXJcIl0nLCBlLnRhcmdldCk7XG5cbiAgICAgICAgaWYgKCRlbGVtZW50Lmxlbmd0aCkge1xuICAgICAgICAgIFFvck9yZGVyLnBsdWdpbi5jYWxsKCRlbGVtZW50KTtcbiAgICAgICAgfVxuICAgICAgfSlcbiAgICAgIC50cmlnZ2VySGFuZGxlcigncmVuZXcucW9yLmluaXRpYXRvcicpO1xuICB9KTtcblxufSk7XG4iLCIkKGZ1bmN0aW9uICgpIHtcblxuICAndXNlIHN0cmljdCc7XG5cbiAgLy8gQWRkIEJvb3RzdHJhcCdzIGNsYXNzZXMgZHluYW1pY2FsbHlcbiAgJCgnLnFvci1sb2NhbGUtc2VsZWN0b3InKS5vbignY2hhbmdlJywgZnVuY3Rpb24gKCkge1xuICAgIHZhciB1cmwgPSAkKHRoaXMpLnZhbCgpO1xuXG4gICAgaWYgKHVybCkge1xuICAgICAgd2luZG93LmxvY2F0aW9uLmFzc2lnbih1cmwpO1xuICAgIH1cbiAgfSk7XG5cbiAgLy8gVG9nZ2xlIHN1Ym1lbnVzXG4gICQoJy5xb3ItbWVudS1ncm91cCcpLm9uKCdjbGljaycsICc+IHVsID4gbGkgPiBhJywgZnVuY3Rpb24gKCkge1xuICAgIHZhciAkbmV4dCA9ICQodGhpcykubmV4dCgpO1xuXG4gICAgaWYgKCRuZXh0LmlzKCd1bCcpICYmICRuZXh0LmNzcygncG9zaXRpb24nKSAhPT0gJ2Fic29sdXRlJykge1xuICAgICAgaWYgKCEkbmV4dC5oYXNDbGFzcygnY29sbGFwc2FibGUnKSkge1xuICAgICAgICAkbmV4dC5hZGRDbGFzcygnY29sbGFwc2FibGUnKS5oZWlnaHQoJG5leHQucHJvcCgnc2Nyb2xsSGVpZ2h0JykpO1xuICAgICAgfVxuXG4gICAgICBpZiAoJG5leHQuaGFzQ2xhc3MoJ2NvbGxhcHNlZCcpKSB7XG4gICAgICAgICRuZXh0LmhlaWdodCgkbmV4dC5wcm9wKCdzY3JvbGxIZWlnaHQnKSk7XG5cbiAgICAgICAgc2V0VGltZW91dChmdW5jdGlvbiAoKSB7XG4gICAgICAgICAgJG5leHQucmVtb3ZlQ2xhc3MoJ2NvbGxhcHNlZCcpO1xuICAgICAgICB9LCAzNTApO1xuICAgICAgfSBlbHNlIHtcbiAgICAgICAgJG5leHQuYWRkQ2xhc3MoJ2NvbGxhcHNlZCcpLmhlaWdodCgwKTtcbiAgICAgIH1cbiAgICB9XG4gIH0pO1xuXG4gICQoJy5xb3Itc2VhcmNoJykuZWFjaChmdW5jdGlvbiAoKSB7XG4gICAgdmFyICR0aGlzID0gJCh0aGlzKSxcbiAgICAgICAgJGxhYmVsID0gJHRoaXMuZmluZCgnLnFvci1zZWFyY2gtbGFiZWwnKSxcbiAgICAgICAgJGlucHV0ID0gJHRoaXMuZmluZCgnLnFvci1zZWFyY2gtaW5wdXQnKSxcbiAgICAgICAgJGNsZWFyID0gJHRoaXMuZmluZCgnLnFvci1zZWFyY2gtY2xlYXInKTtcblxuICAgICRsYWJlbC5vbignY2xpY2snLCBmdW5jdGlvbiAoKSB7XG4gICAgICBpZiAoISRpbnB1dC5oYXNDbGFzcygnZm9jdXMnKSkge1xuICAgICAgICAkdGhpcy5hZGRDbGFzcygnYWN0aXZlJyk7XG4gICAgICAgICRpbnB1dC5hZGRDbGFzcygnZm9jdXMnKTtcbiAgICAgIH1cbiAgICB9KTtcblxuICAgICRjbGVhci5vbignY2xpY2snLCBmdW5jdGlvbiAoKSB7XG4gICAgICBpZiAoJGlucHV0LnZhbCgpKSB7XG4gICAgICAgICRpbnB1dC52YWwoJycpO1xuICAgICAgfSBlbHNlIHtcbiAgICAgICAgJHRoaXMucmVtb3ZlQ2xhc3MoJ2FjdGl2ZScpO1xuICAgICAgICAkaW5wdXQucmVtb3ZlQ2xhc3MoJ2ZvY3VzJyk7XG4gICAgICB9XG4gICAgfSk7XG5cbiAgfSk7XG5cbiAgLy8gSW5pdCBCb290c3RyYXAgTWF0ZXJpYWwgRGVzaWduXG4gICQubWF0ZXJpYWwuaW5pdCgpO1xufSk7XG4iXSwic291cmNlUm9vdCI6Ii9zb3VyY2UvIn0=