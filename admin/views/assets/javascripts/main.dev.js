(function () {

  'use strict';

  var requireOptions = {
        urlArgs: 'bust=' + Date.now(),
        paths: {
          // Plugins
          // submitter: 'plugins/submitter',
          // validator: 'plugins/validator',
          // uploader: 'plugins/uploader'

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
      'qor'
    ], function () {

    });
  });

})();
