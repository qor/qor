(function () {

  'use strict';

  var requireOptions = {
        urlArgs: 'bust=' + (new Date()).getTime(),
        paths: {
          // Plugins
          submitter: 'plugins/submitter',
          validator: 'plugins/validator',
          uploader: 'plugins/uploader'

          // Libraries
          // jquery: 'jquery.min',
          // bootstrap: 'bootstrap.min'
        }
      };

  require.config(requireOptions);

  require([
    'jquery'
  ], function ($) {
    require([
      'bootstrap',
      'submitter',
      'validator'
    ], function () {
      $(function () {

      });
    });
  });

})();
