(function () {

  'use strict';

  var requireOptions = {
        paths: {
          // Plugins
          checkbox: 'plugins/checkbox',
          radio: 'plugins/radio',
          placeholder: 'plugins/placeholder',
          submitter: 'plugins/submitter',
          validator: 'plugins/validator',

          // Libraries
          jquery: 'jquery.min',
          underscore: 'underscore.min',
          bootstrap: 'bootstrap.min'
        }
      };

  require.config(requireOptions);

  require([
    'jquery'
  ], function ($) {
    require([
      'welife',
      'bootstrap',
      'checkbox',
      'radio',
      'placeholder',
      'submitter',
      'validator'
    ], function (WeLife) {
      $(function () {
        var $main = $('.main');

        $main.data('welife', new WeLife($main[0], window._pageConfig || {}));
      });
    });
  });

})();
