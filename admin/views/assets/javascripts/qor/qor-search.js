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
  var location = window.location;
  var componentHandler = window.componentHandler;
  var history = window.history;
  var NAMESPACE = 'qor.globalSearch';
  var EVENT_ENABLE = 'enable.' + NAMESPACE;
  var EVENT_DISABLE = 'disable.' + NAMESPACE;
  var EVENT_CLICK = 'click.' + NAMESPACE;

  var SEARCH_RESOURCE = '.qor-global-search--resource';
  var SEARCH_RESULTS = '.qor-global-search--results';
  var QOR_TABLE = '.qor-table';
  var IS_ACTIVE = 'is-active';

  function QorSearchCenter(element, options) {
    this.$element = $(element);
    this.options = $.extend({}, QorSearchCenter.DEFAULTS, $.isPlainObject(options) && options);
    this.init();
  }

  QorSearchCenter.prototype = {
    constructor: QorSearchCenter,

    init: function () {
      this.bind();
      this.initTab();
    },

    bind: function () {
      this.$element.on(EVENT_CLICK, $.proxy(this.click, this));
    },

    unbind: function () {
      this.$element.off(EVENT_CLICK, this.check);
    },

    initTab: function () {
      var locationSearch = location.search;
      var resourceName;
      if (/resource_name/.test(locationSearch)){
        resourceName = locationSearch.match(/resource_name=\w+/g).toString().split('=')[1];
        $(SEARCH_RESOURCE).removeClass(IS_ACTIVE);
        $('[data-resource="' + resourceName + '"]').addClass(IS_ACTIVE);
      }
    },

    click : function (e) {
      var $target = $(e.target);
      var data = $target.data();

      if ($target.is(SEARCH_RESOURCE)){
        var oldUrl = location.href.replace(/#/g, '');
        var newUrl;
        var newResourceName = data.resource;
        var hasResource = /resource_name/.test(oldUrl);
        var hasKeyword = /keyword/.test(oldUrl);
        var resourceParam = 'resource_name=' + newResourceName;
        var searchSymbol = hasKeyword ? '&' : '?';

        if (newResourceName){
          if (hasResource){
            newUrl = oldUrl.replace(/resource_name=\w+/g, resourceParam);
          } else {
            newUrl = oldUrl + searchSymbol + resourceParam;
          }
        } else {
          newUrl = oldUrl.replace(/&resource_name=\w+/g, '');
        }

        if (history.pushState){
          this.fetchSearch(newUrl, $target);
        } else {
          location.href = newUrl;
        }

      }
    },

    fetchSearch: function (url,$target) {
      var title = document.title;

      $.ajax(url, {
        method: 'GET',
        dataType: 'html',
        beforeSend: function () {
          $('.mdl-spinner').remove();
          $(SEARCH_RESULTS).prepend('<div class="mdl-spinner mdl-js-spinner is-active"></div>').find('.qor-section').hide();
          componentHandler.upgradeElement(document.querySelector('.mdl-spinner'));
        },
        success: function (html) {
          var result = $(html).find(SEARCH_RESULTS).html();
          $(SEARCH_RESOURCE).removeClass(IS_ACTIVE);
          $target.addClass(IS_ACTIVE);
          // change location URL without refresh page
          history.pushState({ Page: url, Title: title }, title, url);
          $('.mdl-spinner').remove();
          $(SEARCH_RESULTS).removeClass('loading').html(result);
          componentHandler.upgradeElements(document.querySelectorAll(QOR_TABLE));
        },
        error: function (xhr, textStatus, errorThrown) {
          $(SEARCH_RESULTS).find('.qor-section').show();
          $('.mdl-spinner').remove();
          window.alert([textStatus, errorThrown].join(': '));
        }
      });
    },

    destroy: function () {
      this.unbind();
      this.$element.removeData(NAMESPACE);
    }

  };

  QorSearchCenter.DEFAULTS = {
  };

  QorSearchCenter.plugin = function (options) {
    return this.each(function () {
      var $this = $(this);
      var data = $this.data(NAMESPACE);
      var fn;

      if (!data) {
        $this.data(NAMESPACE, (data = new QorSearchCenter(this, options)));
      }

      if (typeof options === 'string' && $.isFunction(fn = data[options])) {
        fn.call(data);
      }
    });
  };

  $(function () {
    var selector = '[data-toggle="qor.global.search"]';
    var options = {};

    $(document).
      on(EVENT_DISABLE, function (e) {
        QorSearchCenter.plugin.call($(selector, e.target), 'destroy');
      }).
      on(EVENT_ENABLE, function (e) {
        QorSearchCenter.plugin.call($(selector, e.target), options);
      }).
      triggerHandler(EVENT_ENABLE);
  });

  return QorSearchCenter;

});
