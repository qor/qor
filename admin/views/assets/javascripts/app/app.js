$(function () {

  'use strict';

  // Add Bootstrap's classes dynamically
  $('.qor-locale-selector').on('change', function () {
    var url = $(this).val();

    if (url) {
      window.location.assign(url);
    }
  });

  $('.qor-search').each(function () {
    var $this = $(this),
        $label = $this.find('.qor-search-label'),
        $input = $this.find('.qor-search-input'),
        $clear = $this.find('.qor-search-clear');

    $label.on('click', function () {
      if (!$input.hasClass('focus')) {
        $this.addClass('active');
        $input.addClass('focus');
      }
    });

    $clear.on('click', function () {
      if ($input.val()) {
        $input.val('');
      } else {
        $this.removeClass('active');
        $input.removeClass('focus');
      }
    });

  });

  // Init Bootstrap Material Design
  $.material.init();
});
