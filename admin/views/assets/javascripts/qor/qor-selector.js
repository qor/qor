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
  var NAMESPACE = 'qor.selector';
  var EVENT_ENABLE = 'enable.' + NAMESPACE;
  var EVENT_DISABLE = 'disable.' + NAMESPACE;
  var EVENT_CLICK = 'click.' + NAMESPACE;
  var CLASS_OPEN = 'open';
  var CLASS_ACTIVE = 'active';
  var CLASS_SELECTED = 'selected';
  var CLASS_DISABLED = 'disabled';
  var CLASS_CLEARABLE = 'clearable';
  var SELECTOR_SELECTED = '.' + CLASS_SELECTED;
  var SELECTOR_TOGGLE = '.qor-selector-toggle';
  var SELECTOR_LABEL = '.qor-selector-label';
  var SELECTOR_CLEAR = '.qor-selector-clear';
  var SELECTOR_MENU = '.qor-selector-menu';

  function QorSelector(element, options) {
    this.options = options;
    this.$element = $(element);
    this.init();
  }

  QorSelector.prototype = {
    constructor: QorSelector,

    init: function () {
      var $this = this.$element;

      this.placeholder = $this.attr('placeholder') || $this.attr('name') || 'Select';
      this.build();
    },

    build: function () {
      var $this = this.$element;
      var $selector = $(QorSelector.TEMPLATE);
      var alignedClass = this.options.aligned + '-aligned';
      var data = {};

      $selector.addClass(alignedClass).find(SELECTOR_MENU).html(function () {
        var list = [];

        $this.children().each(function () {
          var $this = $(this);
          var selected = $this.attr('selected');
          var disabled = $this.attr('disabled');
          var value = $this.attr('value');
          var label = $this.text();
          var classNames = [];

          if (selected) {
            classNames.push(CLASS_SELECTED);
            data.value = value;
            data.label = label;
          }

          if (disabled) {
            classNames.push(CLASS_DISABLED);
          }

          list.push(
            '<li' +
              (classNames.length ? ' class="' + classNames.join(' ') + '"' : '') +
              ' data-value="' + value + '"' +
              ' data-label="' + label + '"' +
            '>' +
              label +
            '</li>'
          );
        });

        return list.join('');
      });

      this.$selector = $selector;
      $this.hide().after($selector);
      this.pick(data, true);
      this.bind();
    },

    unbuild: function () {
      this.unbind();
      this.$selector.remove();
      this.$element.show();
    },

    bind: function () {
      this.$selector.on(EVENT_CLICK, $.proxy(this.click, this));
      $document.on(EVENT_CLICK, $.proxy(this.close, this));
    },

    unbind: function () {
      this.$selector.off(EVENT_CLICK, this.click);
      $document.off(EVENT_CLICK, this.close);
    },

    click: function (e) {
      var $target = $(e.target);

      e.stopPropagation();

      if ($target.is(SELECTOR_CLEAR)) {
        this.clear();
      } else if ($target.is('li')) {
        if (!$target.hasClass(CLASS_SELECTED) && !$target.hasClass(CLASS_DISABLED)) {
          this.pick($target.data());
        }

        this.close();
      } else if ($target.closest(SELECTOR_TOGGLE).length) {
        this.open();
      }
    },

    pick: function (data, initialized) {
      var $selector = this.$selector;
      var selected = !!data.value;

      $selector.
        find(SELECTOR_TOGGLE).
        toggleClass(CLASS_ACTIVE, selected).
        toggleClass(CLASS_CLEARABLE, selected && this.options.clearable).
          find(SELECTOR_LABEL).
          text(data.label || this.placeholder);

      if (!initialized) {
        $selector.
          find(SELECTOR_MENU).
            children('[data-value="' + data.value + '"]').
            addClass(CLASS_SELECTED).
            siblings(SELECTOR_SELECTED).
            removeClass(CLASS_SELECTED);

        this.$element.val(data.value).trigger('change');
      }
    },

    clear: function () {
      this.$selector.
        find(SELECTOR_TOGGLE).
        removeClass(CLASS_ACTIVE).
        removeClass(CLASS_CLEARABLE).
          find(SELECTOR_LABEL).
          text(this.placeholder).
          end().
        end().
        find(SELECTOR_MENU).
          children(SELECTOR_SELECTED).
          removeClass(CLASS_SELECTED);

      this.$element.val('').trigger('change');
    },

    open: function () {

      // Close other opened dropdowns first
      $document.triggerHandler(EVENT_CLICK);

      // Open the current dropdown
      this.$selector.addClass(CLASS_OPEN);
    },

    close: function () {
      this.$selector.removeClass(CLASS_OPEN);
    },

    destroy: function () {
      this.unbuild();
      this.$element.removeData(NAMESPACE);
    },
  };

  QorSelector.DEFAULTS = {
    aligned: 'left',
    clearable: false,
  };

  QorSelector.TEMPLATE = (
    '<div class="qor-selector">' +
      '<a class="qor-selector-toggle">' +
        '<span class="qor-selector-label"></span>' +
        '<i class="material-icons qor-selector-arrow">arrow_drop_down</i>' +
        '<i class="material-icons qor-selector-clear">clear</i>' +
      '</a>' +
      '<ul class="qor-selector-menu"></ul>' +
    '</div>'
  );

  QorSelector.plugin = function (option) {
    return this.each(function () {
      var $this = $(this);
      var data = $this.data(NAMESPACE);
      var options;
      var fn;

      if (!data) {
        if (/destroy/.test(option)) {
          return;
        }

        options = $.extend({}, QorSelector.DEFAULTS, $this.data(), typeof option === 'object' && option);
        $this.data(NAMESPACE, (data = new QorSelector(this, options)));
      }

      if (typeof option === 'string' && $.isFunction(fn = data[option])) {
        fn.apply(data);
      }
    });
  };

  $(function () {
    var selector = '[data-toggle="qor.selector"]';

    $(document).
      on(EVENT_DISABLE, function (e) {
        QorSelector.plugin.call($(selector, e.target), 'destroy');
      }).
      on(EVENT_ENABLE, function (e) {
        QorSelector.plugin.call($(selector, e.target));
      }).
      triggerHandler(EVENT_ENABLE);
  });

  return QorSelector;

});
