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

  var NAMESPACE = 'qor.replicator';
  var EVENT_ENABLE = 'enable.' + NAMESPACE;
  var EVENT_DISABLE = 'disable.' + NAMESPACE;
  var EVENT_CLICK = 'click.' + NAMESPACE;
  var IS_TEMPLATE = 'is-template';

  function QorReplicator(element, options) {
    this.$element = $(element);
    this.options = $.extend({}, QorReplicator.DEFAULTS, $.isPlainObject(options) && options);
    this.index = 0;
    this.init();
  }

  QorReplicator.prototype = {
    constructor: QorReplicator,

    init: function () {
      var $this = this.$element;
      var options = this.options;
      var $all = $this.find(options.itemClass);
      var $template;

      if (!$all.length) {
        return;
      }

      $template = $all.filter(options.newClass);

      if (!$template.length) {
        $template = $all.last();
      }

      // Should destroy all components here
      $template.trigger('disable');

      this.$template = $template;
      this.template = $template.filter($this.children(options.childrenClass).children(options.newClass)).prop('outerHTML');
      $template.data(IS_TEMPLATE, true).hide();

      this.parse();
      this.bind();
    },

    parse: function () {
      var i = 0;

      this.template = this.template.replace(/(\w+)\="(\S*\[\d+\]\S*)"/g, function (attribute, name, value) {
        value = value.replace(/^(\S*)\[(\d+)\]([^\[\]]*)$/, function (input, prefix, index, suffix) {
          if (input === value) {
            if (name === 'name') {
              i = index;
            }

            return (prefix + '[{{index}}]' + suffix);
          }
        });

        return (name + '="' + value + '"');
      });

      this.index = parseFloat(i);
    },

    bind: function () {
      var options = this.options;

      this.$element.
        on(EVENT_CLICK, options.addClass, $.proxy(this.add, this)).
        on(EVENT_CLICK, options.delClass, $.proxy(this.del, this));
    },

    unbind: function () {
      this.$element.
        off(EVENT_CLICK, this.add).
        off(EVENT_CLICK, this.del);
    },

    add: function (e) {
      var options = this.options;
      var $target = $(e.target).closest(this.options.addClass);
      var $template = this.$template.filter($target.closest(this.$element).children(options.childrenClass).children(options.newClass));
      var $item = $template;

      if ($template && $template.is(':hidden')) {
        $template.show();
      } else {
        if ($target.length) {
          $item = $(this.template.replace(/\{\{index\}\}/g, ++this.index));
          $target.before($item.show());
        }
      }

      if ($item) {

        // Enable all JavaScript components within the fieldset
        $item.trigger('enable');
      }
      e.stopPropagation();
    },

    del: function (e) {
      var options = this.options;
      var $item = $(e.target).closest(options.itemClass);
      var $alert;

      if ($item.is(options.newClass)) {
        if ($item.data(IS_TEMPLATE)) {
          this.$template = null;
        }

        // Destroy all JavaScript components within the fieldset
        $item.trigger('disable').remove();
      } else {
        $item.children(':visible').addClass('hidden').hide();
        $alert = $(options.alertTemplate.replace('{{name}}', this.parseName($item)));
        $alert.find(options.undoClass).one(EVENT_CLICK, function () {
          $alert.remove();
          $item.children('.hidden').removeClass('hidden').show();
        });
        $item.append($alert);
      }
    },

    parseName: function ($item) {
      var name = $item.find('input[name]').attr('name');

      if (name) {
        return name.replace(/[^\[\]]+$/, '');
      }
    },

    destroy: function () {
      this.unbind();
      this.$element.removeData(NAMESPACE);
    },
  };

  QorReplicator.DEFAULTS = {
    itemClass: false,
    newClass: false,
    addClass: false,
    delClass: false,
    childrenClass: false,
    alertTemplate: '',
  };

  QorReplicator.plugin = function (options) {
    return this.each(function () {
      var $this = $(this);
      var data = $this.data(NAMESPACE);
      var fn;

      if (!data) {
        $this.data(NAMESPACE, (data = new QorReplicator(this, options)));
      }

      if (typeof options === 'string' && $.isFunction(fn = data[options])) {
        fn.call(data);
      }
    });
  };

  $(function () {
    var selector = '.qor-fieldset-container';
    var options = {
          itemClass: '.qor-fieldset',
          newClass: '.qor-fieldset--new',
          addClass: '.qor-fieldset__add',
          delClass: '.qor-fieldset__delete',
          childrenClass: '.qor-field__block',
          undoClass: '.qor-fieldset__undo',
          alertTemplate: (
            '<div class="qor-fieldset__alert">' +
              '<input type="hidden" name="{{name}}._destroy" value="1">' +
              '<button class="mdl-button mdl-button--accent mdl-js-button mdl-js-ripple-effect qor-fieldset__undo" type="button">Undo delete</button>' +
            '</div>'
          ),
        };

    $(document).
      on(EVENT_DISABLE, function (e) {
        QorReplicator.plugin.call($(selector, e.target), 'destroy');
      }).
      on(EVENT_ENABLE, function (e) {
        QorReplicator.plugin.call($(selector, e.target), options);
      }).
      triggerHandler(EVENT_ENABLE);
  });

  return QorReplicator;

});
