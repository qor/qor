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

  var NAMESPACE = 'qor.material';
  var EVENT_ENABLE = 'enable.' + NAMESPACE;
  var EVENT_DISABLE = 'disable.' + NAMESPACE;
  var EVENT_UPDATE = 'update.' + NAMESPACE;
  var SELECTOR_COMPONENT = '[class*="mdl-js"]';

  function enable(target) {

    /*jshint undef:false */
    if (componentHandler) {

      // Enable all MDL (Material Design Lite) components within the target element
      if ($(target).is(SELECTOR_COMPONENT)) {
        componentHandler.upgradeElements(target);
      } else {
        componentHandler.upgradeElements($(SELECTOR_COMPONENT, target).toArray());
      }
    }
  }

  function disable(target) {

    /*jshint undef:false */
    if (componentHandler) {

      // Destroy all MDL (Material Design Lite) components within the target element
      if ($(target).is(SELECTOR_COMPONENT)) {
        componentHandler.downgradeElements(target);
      } else {
        componentHandler.downgradeElements($(SELECTOR_COMPONENT, target).toArray());
      }
    }
  }

  $(function () {
    $(document).
      on(EVENT_ENABLE, function (e) {
        enable(e.target);
      }).
      on(EVENT_DISABLE, function (e) {
        disable(e.target);
      }).
      on(EVENT_UPDATE, function (e) {
        disable(e.target);
        enable(e.target);
      });
  });

});
