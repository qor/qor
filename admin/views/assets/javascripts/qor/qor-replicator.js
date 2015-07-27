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

  var NAMESPACE = 'qor.replicator',
      EVENT_ENABLE = 'enable.' + NAMESPACE,
      EVENT_DISABLE = 'disable.' + NAMESPACE,
      EVENT_CLICK = 'click.' + NAMESPACE,

      QorReplicator = function (element, options) {
        this.$element = $(element);
        this.options = $.extend({}, QorReplicator.DEFAULTS, $.isPlainObject(options) && options);
        this.index = 0;
        this.init();
      };

  QorReplicator.prototype = {
    constructor: QorReplicator,

    init: function () {
      var $this = this.$element,
          options = this.options,
          $all = $this.find(options.itemClass),
          $template;

      if (!$all.length) {
        return;
      }

      $template = $all.filter(options.newClass);

      if (!$template.length) {
        $template = $all.last();
      }

      this.$template = $template;
      this.template = $template.prop('outerHTML');
      $template.hide();

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
      var $template = this.$template,
          $target;

      if ($template.is(':hidden')) {
        $template.show();
        return;
      }

      $target = $(e.target).closest(this.options.addClass);

      if ($target.length) {
        $target.before(this.template.replace(/\{\{index\}\}/g, ++this.index));
      }
    },

    del: function (e) {
      var options = this.options,
          $item = $(e.target).closest(options.itemClass),
          $alert;

      if ($item.is(options.newClass)) {
        $item.remove();
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
    }
  };

  QorReplicator.DEFAULTS = {
    itemClass: false,
    newClass: false,
    addClass: false,
    delClass: false,
    alertTemplate: ''
  };

  QorReplicator.plugin = function (options) {
    return this.each(function () {
      var $this = $(this),
          data = $this.data(NAMESPACE),
          fn;

      if (!data) {
        $this.data(NAMESPACE, (data = new QorReplicator(this, options)));
      }

      if (typeof options === 'string' && $.isFunction(fn = data[options])) {
        fn.call(data);
      }
    });
  };

  $(function () {
    var selector = '.qor-collection-group',
        options = {
          itemClass: '.qor-collection',
          newClass: '.qor-collection-new',
          addClass: '.qor-collection-add',
          delClass: '.qor-collection-del',
          undoClass: '.qor-collection-undo',
          alertTemplate: ('<div class="alert alert-danger"><input type="hidden" name="{{name}}._destroy" value="1"><a href="javascript:void(0);" class="alert-link qor-collection-undo">Undo Delete</a></div>')
        };

    $(document)
      .on(EVENT_CLICK, selector, function () {
        QorReplicator.plugin.call($(this), options);
      })
      .on(EVENT_DISABLE, function (e) {
        QorReplicator.plugin.call($(selector, e.target), 'destroy');
      })
      .on(EVENT_ENABLE, function (e) {
        QorReplicator.plugin.call($(selector, e.target), options);
      })
      .triggerHandler(EVENT_ENABLE);
  });

  return QorReplicator;

});
