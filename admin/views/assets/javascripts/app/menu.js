$(function () {

  'use strict';

  var NAMESPACE = 'qor.menu';
  var EVENT_CLICK = 'click.' + NAMESPACE;
  var EVENT_TRANSITION_END = 'transitionend.' + NAMESPACE;

  function Menu(element, options) {
    this.$element = $(element);
    this.options = $.extend({}, Menu.DEFAULTS, $.isPlainObject(options) && options);
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
          $ul.addClass('collapse in').height($ul.prop('scrollHeight'));
        }
      });

      this.bind();
    },

    bind: function () {
      this.$element.on(EVENT_CLICK, '> ul > li > a', $.proxy(this.click, this));
    },

    unbind: function () {
      this.$element.off(EVENT_CLICK, this.click);
    },

    click: function (e) {
      var $li = $(e.currentTarget).closest('li'),
          $ul = $li.find('ul');

      if ($ul.hasClass('in')) {
        $li.removeClass('expanded');
        $ul.one(EVENT_TRANSITION_END, function () {
          $ul.removeClass('collapsing in');
        }).addClass('collapsing').height(0);
      } else {
        $li.addClass('expanded');
        $ul.one(EVENT_TRANSITION_END, function () {
          $ul.removeClass('collapsing');
        }).addClass('collapsing in').height($ul.prop('scrollHeight'));
      }
    },

    destroy: function () {
      this.unbind();
      this.$element.removeData(NAMESPACE);
    }
  };

  Menu.DEFAULTS = {};

  Menu.plugin = function (options) {
    return this.each(function () {
      var $this = $(this);
      var data = $this.data(NAMESPACE);
      var fn;

      if (!data) {
        $this.data(NAMESPACE, (data = new Menu(this, options)));
      }

      if (typeof options === 'string' && $.isFunction((fn = data[options]))) {
        fn.apply(data);
      }
    });
  };

  $(function () {
    Menu.plugin.call($('.qor-menu-container'));
  });

});
