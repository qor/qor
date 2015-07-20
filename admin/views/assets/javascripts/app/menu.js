$(function () {

  'use strict';

  var $window = $(window);
  var NAMESPACE = 'qor.menu';
  var EVENT_CLICK = 'click.' + NAMESPACE;
  var EVENT_RESIZE = 'resize.' + NAMESPACE;

  function Menu(element, options) {
    this.$element = $(element);
    this.options = $.extend({}, Menu.DEFAULTS, $.isPlainObject(options) && options);
    this.disabled = false;
    this.init();
  }

  Menu.prototype = {
    constructor: Menu,

    init: function () {
      var $this = this.$element;

      $this.find('> ul > li').each(function () {
        var $this = $(this);
        var $ul = $this.find('> ul');

        if ($ul.length) {
          $this.addClass('expandable expanded');
        } else {
          $this.append($ul = $(Menu.TEMPLATE));
        }

        $ul.attr('data-menu', $this.data('menuItem'));
      });

      this.resize();
      this.bind();
    },

    bind: function () {
      this.$element.on(EVENT_CLICK, '> ul > li > a', $.proxy(this.click, this));
      $window.on(EVENT_RESIZE, $.proxy(this.resize, this));
    },

    unbind: function () {
      this.$element.off(EVENT_CLICK, this.click);
      $window.off(EVENT_RESIZE, this.resize);
    },

    click: function (e) {
      var $li = $(e.currentTarget).closest('li'),
          $ul = $li.find('ul');

      if (this.disabled) {
        return;
      }

      if (!$ul.hasClass('collapsable')) {
        $ul.addClass('collapsable').height($ul.prop('scrollHeight'));
      }

      if ($ul.hasClass('collapsed')) {
        $li.addClass('expanded');
        $ul.height($ul.prop('scrollHeight'));

        setTimeout(function () {
          $ul.removeClass('collapsed');
        }, 350);
      } else {
        $li.removeClass('expanded');
        $ul.addClass('collapsed').height(0);
      }
    },

    resize: function (e) {
      this.disabled = $window.innerWidth() <= this.options.gridFloatBreakpoint;

      if (e) {
        this.$element.find('> ul > li > ul').removeClass('collapsable collapsed').height('auto');
      }
    },

    destroy: function () {
      this.unbind();
      this.$element.removeData(NAMESPACE);
    }
  };

  Menu.DEFAULTS = {
    gridFloatBreakpoint: 768
  };

  Menu.TEMPLATE = '<ul class="qor-menu"></ul>';

  Menu.plugin = function (options) {
    return this.each(function () {
      var $this = $(this);
      var data = $this.data(NAMESPACE);
      var fn;

      if (!data) {
        if (!$.fn.modal) {
          return;
        }

        $this.data(NAMESPACE, (data = new Menu(this, options)));
      }

      if (typeof options === 'string' && $.isFunction((fn = data[options]))) {
        fn.apply(data);
      }
    });
  };

  $(function () {
    Menu.plugin.call($('.qor-menu-group'));
  });

});
