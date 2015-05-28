(function (factory) {
  if (typeof define === 'function' && define.amd) {
    // AMD. Register as anonymous module.
    define('qor-replicator', ['jquery'], factory);
  } else if (typeof exports === 'object') {
    // Node / CommonJS
    factory(require('jquery'));
  } else {
    // Browser globals.
    factory(jQuery);
  }
})(function ($) {

  'use strict';

  var QorReplicator = function (element, options) {
        this.$element = $(element);
        this.options = $.extend({}, QorReplicator.DEFAULTS, options);
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
      this.template = $template.clone().removeClass('hide').prop('outerHTML');
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
      var $this = this.$element,
          options = this.options;

      $this.on('click', options.addClass, $.proxy(this.add, this));
      $this.on('click', options.delClass, $.proxy(this.del, this));
    },

    add: function (e) {
      var $template = this.$template,
          $target;

      if ($template.hasClass('hide')) {
        $template.removeClass('hide');
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
        $alert.find(options.undoClass).one('click', function () {
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
    }
  };

  QorReplicator.DEFAULTS = {
    itemClass: '',
    newClass: '',
    addClass: '',
    delClass: '',
    alertTemplate: ''
  };

  $(function () {
    $('.qor-collection-group').each(function () {
      var $this = $(this);

      if (!$this.data('qor.replicator')) {
        $this.data('qor.replicator', new QorReplicator(this, {
          itemClass: '.qor-collection',
          newClass: '.qor-collection-new',
          addClass: '.qor-collection-add',
          delClass: '.qor-collection-del',
          undoClass: '.qor-collection-undo',
          alertTemplate: '<div class="alert alert-danger"><input type="hidden" name="{{name}}._destroy" value="1"><a href="javascript:void(0);" class="alert-link qor-collection-undo">Undo Delete</a></div>'
        }));
      }
    });
  });

});
