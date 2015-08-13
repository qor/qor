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

  var NAMESPACE = 'qor.sorting';
  var EVENT_ENABLE = 'enable.' + NAMESPACE;
  var EVENT_DISABLE = 'disable.' + NAMESPACE;
  var EVENT_CHANGE = 'change.' + NAMESPACE;
  var EVENT_MOUSE_DOWN = 'mousedown.' + NAMESPACE;
  var EVENT_MOUSE_UP = 'mouseup.' + NAMESPACE;
  var EVENT_DRAG_START = 'dragstart.' + NAMESPACE;
  var EVENT_DRAG_OVER = 'dragover.' + NAMESPACE;
  var EVENT_DROP = 'drop.' + NAMESPACE;
  var CLASS_SORTING = 'qor-sorting';
  var CLASS_HIGHLIGHT = 'qor-sorting__highlight';
  var SELECTOR_TR = 'tbody> tr';

  function QorSorter(element, options) {
    this.$element = $(element);
    this.options = $.extend({}, QorSorter.DEFAULTS, $.isPlainObject(options) && options);
    this.$source = null;
    this.ascending = true;
    // this.descending = false;
    this.init();
  }

  QorSorter.prototype = {
    constructor: QorSorter,

    init: function () {
      var options = this.options;
      var $this = this.$element;
      var $rows = $this.find(SELECTOR_TR);
      var firstPosition = $rows.first().find(options.input).data('position');
      var lastPosition = $rows.last().find(options.input).data('position');

      $('body').addClass(CLASS_SORTING);
      $this.find('tbody .qor-table__actions').append(QorSorter.TEMPLATE);
      this.ascending = firstPosition < lastPosition;
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
      var $rows = this.$element.find(SELECTOR_TR);
      var $sourceInput = $(e.currentTarget);
      var $source = $sourceInput.closest('tr');
      var $tbody = $source.parent();
      var source = $sourceInput.data();
      var sourcePosition = source.position;
      var targetPosition = parseInt($sourceInput.val(), 10);
      var largethan = targetPosition > sourcePosition;
      var ascending = this.ascending;
      var targetIndex;
      var $target;

      $rows.each(function (i) {
        var $this = $(this);
        var $input = $this.find(options.input);
        var position = $input.data('position');

        if (position === targetPosition) {
          targetIndex = i;
        }

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

      $sourceInput.data('position', targetPosition);

      if (typeof targetIndex === 'number') {
        $target = $rows.eq(targetIndex);

        if (largethan) {
          if (ascending) {
            $target.after($source);
          } else {
            $target.before($source);
          }
        } else {
          if (ascending) {
            $target.before($source);
          } else {
            $target.after($source);
          }
        }
      } else {
        if (largethan) {
          if (ascending) {
            $tbody.append($source);
          } else {
            $tbody.prepend($source);
          }
        } else {
          if (ascending) {
            $tbody.prepend($source);
          } else {
            $tbody.append($source);
          }
        }
      }

      this.sort($source, {
        url: source.sortingUrl,
        from: sourcePosition,
        to: targetPosition,
      });
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
      var ascending = this.ascending;
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
        if (ascending) {
          $target.after($source);
        } else {
          $target.before($source);
        }
      } else {
        if (ascending) {
          $target.before($source);
        } else {
          $target.after($source);
        }
      }

      this.sort($source, {
        url: source.sortingUrl,
        from: sourcePosition,
        to: targetPosition,
      });
    },

    sort: function ($row, data) {
      var options = this.options;

      if (data.url) {
        this.highlight($row);

        $.ajax(data.url, {
          method: 'post',
          data: {
            from: data.from,
            to: data.to,
          },
          success: function (actualPosition, textStatus, jqXHR) {
            if (jqXHR.status === 200) {
              $row.find(options.input).data('position', actualPosition).val(actualPosition);
            }
          },
          error:function () {
            if (windwo.alert('Fail to sort!')) {
              window.location.reload();
            }
          }
        });
      }
    },

    highlight: function ($row) {
      $row.addClass(CLASS_HIGHLIGHT);

      setTimeout(function () {
        $row.removeClass(CLASS_HIGHLIGHT);
      }, 2000);
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

  QorSorter.TEMPLATE = '<a class="qor-sorting__toggle"><i class="material-icons">swap_vert</i></a>';

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

    var selector = '.qor-table';
    var options = {
          toggle: '.qor-sorting__toggle',
          input: '.qor-sorting__position',
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
