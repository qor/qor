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

  var NAMESPACE = 'qor.sorter';
  var EVENT_ENABLE = 'enable.' + NAMESPACE;
  var EVENT_DISABLE = 'disable.' + NAMESPACE;
  var EVENT_CHANGE = 'change.' + NAMESPACE;
  var EVENT_MOUSE_DOWN = 'mousedown.' + NAMESPACE;
  var EVENT_MOUSE_UP = 'mouseup.' + NAMESPACE;
  var EVENT_DRAG_START = 'dragstart.' + NAMESPACE;
  var EVENT_DRAG_OVER = 'dragover.' + NAMESPACE;
  var EVENT_DROP = 'drop.' + NAMESPACE;
  var SELECTOR_TR = 'tbody> tr';

  function QorSorter(element, options) {
    this.$element = $(element);
    this.options = $.extend({}, QorSorter.DEFAULTS, $.isPlainObject(options) && options);
    this.$source = null;
    this.init();
  }

  QorSorter.prototype = {
    constructor: QorSorter,

    init: function () {
      $('body').addClass('qor-sorter');
      this.$element.find('tbody .qor-list-action').append(QorSorter.TEMPLATE);
      this.bind();
    },

    bind: function () {
      var options = this.options;

      this.$element.
        on(EVENT_CHANGE, options.input, $.proxy(this.change, this)).
        on(EVENT_MOUSE_DOWN, options.toggle, $.proxy(this.mousedown, this)).
        on(EVENT_MOUSE_UP, $.proxy(this.mouseup, this)).
        on(EVENT_DRAG_START, SELECTOR_TR, $.proxy(this.dragstart, this)).
        on(EVENT_DRAG_OVER, SELECTOR_TR, $.proxy(this.dragover, this)).
        on(EVENT_DROP, SELECTOR_TR, $.proxy(this.drop, this));
    },

    unbind: function () {
      this.$element.
        off(EVENT_CHANGE, this.change).
        off(EVENT_MOUSE_DOWN, this.mousedown);
    },

    change: function (e) {
      var options = this.options;
      var $sourceInput = $(e.currentTarget);
      var $source = $sourceInput.closest('tr');
      var $tbody = $source.parent();
      var source = $sourceInput.data();
      var sourcePosition = source.position;
      var targetPosition = parseInt($sourceInput.val(), 10);
      var largethan = targetPosition > sourcePosition;
      var moved;

      this.$element.find(SELECTOR_TR).each(function () {
        var $this = $(this);
        var $input = $this.find(options.input);
        var position = $input.data('position');

        if (largethan) {
          if (position > sourcePosition && position <= targetPosition) {
            if (position === targetPosition) {
              $this.before($source);
              moved = true;
            }

            $input.data('position', --position).val(position);
          }
        } else {
          if (position < sourcePosition && position >= targetPosition) {
            if (position === targetPosition) {
              $this.after($source);
              moved = true;
            }

            $input.data('position', ++position).val(position);
          }
        }
      });

      $sourceInput.data('position', targetPosition);

      if (!moved) {
        if (largethan) {
          $tbody.prepend($source);
        } else {
          $tbody.append($source);
        }
      }

      this.sort(source.sortingUrl, sourcePosition, targetPosition);
    },

    mousedown: function (e) {
      $(e.currentTarget).closest('tr').prop('draggable', true);
    },

    mouseup: function () {
      this.$element.find(SELECTOR_TR).prop('draggable', false);
    },

    dragstart: function (e) {
      var event = e.originalEvent,
          $target = $(e.currentTarget);

      if ($target.prop('draggable') && event.dataTransfer) {
        event.dataTransfer.effectAllowed = 'move';
        this.$source = $target;
      }
    },

    dragover: function (e) {
      var $source = this.$source;

      if (!$source || e.currentTarget === this.$source[0]) {
        return;
      }

      e.preventDefault();
    },

    drop: function (e) {
      var options = this.options;
      var $source = this.$source;
      var $sourceInput;
      var $target;
      var source;
      var sourcePosition;
      var targetPosition;
      var largethan;

      if (!$source || e.currentTarget === this.$source[0]) {
        return;
      }

      e.preventDefault();

      $target = $(e.currentTarget);

      $sourceInput = $source.find(options.input);
      source = $sourceInput.data();
      sourcePosition = source.position;
      targetPosition = $target.find(options.input).data('position');
      largethan = targetPosition > sourcePosition;

      this.$element.find(SELECTOR_TR).each(function () {
        var $input = $(this).find(options.input);
        var position = $input.data('position');

        if (largethan) {
          if (position > sourcePosition && position <= targetPosition) {
            $input.data('position', --position).val(position);
          }
        } else {
          if (position < sourcePosition && position >= targetPosition) {
            $input.data('position', ++position).val(position);
          }
        }
      });

      $sourceInput.data('position', targetPosition).val(targetPosition);

      if (largethan) {
        $target.before($source);
      } else {
        $target.after($source);
      }

      this.sort(source.sortingUrl, sourcePosition, targetPosition);
    },

    sort: function (url, from, to) {
      if (url) {
        $.post(url, {
          from: from,
          to: to,
        });
      }
    },

    destroy: function () {
      this.unbind();
      this.$element.removeData(NAMESPACE);
    },
  };

  QorSorter.DEFAULTS = {
    toggle: false,
    input: false,
  };

  QorSorter.TEMPLATE = '<a class="qor-sorter-toggle"><i class="material-icons">swap_vert</i></a>';

  QorSorter.plugin = function (options) {
    return this.each(function () {
      var $this = $(this);
      var data = $this.data(NAMESPACE);
      var fn;

      if (!data) {
        if (/destroy/.test(options)) {
          return;
        }

        $this.data(NAMESPACE, (data = new QorSorter(this, options)));
      }

      if (typeof options === 'string' && $.isFunction(fn = data[options])) {
        fn.apply(data);
      }
    });
  };

  $(function () {
    if (!/sorting\=true/.test(window.location.search)) {
      return;
    }

    var selector = '.qor-list';
    var options = {
          toggle: '.qor-sorter-toggle',
          input: '.qor-sorting-position',
        };

    $(document).
      on(EVENT_DISABLE, function (e) {
        QorSorter.plugin.call($(selector, e.target), 'destroy');
      }).
      on(EVENT_ENABLE, function (e) {
        QorSorter.plugin.call($(selector, e.target), options);
      }).
      trigger('disable.qor.slideout').
      triggerHandler(EVENT_ENABLE);
  });

  return QorSorter;

});
