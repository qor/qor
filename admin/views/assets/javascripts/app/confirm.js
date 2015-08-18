$(function () {

  'use strict';

  $(document).on('click.qor.confirm', '[data-confirm]', function (e) {
    var $this = $(this);
    var data = $this.data();
    var url;

    if (data.confirm) {
      if (window.confirm(data.confirm)) {
        if (/DELETE/i.test(data.method)) {
          e.preventDefault();

          url = data.url || $this.attr('href');
          data = $.extend({}, data, {
            _method: 'DELETE'
          });

          $.post(url, data, function () {
            window.location.reload();
          });

        }
      } else {
        e.preventDefault();
      }
    }
  });

});
