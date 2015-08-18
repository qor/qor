$(function () {

  'use strict';

  $(function () {
    $(document).on('click.qor.alert', '[data-dismiss="alert"]', function () {
      $(this).closest('.qor-alert').remove();
    });

    setTimeout(function () {
      $('.qor-alert[data-dismissible="true"]').remove();
    }, 3000);
  });

});
