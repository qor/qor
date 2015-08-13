$(function () {

  'use strict';

  $('.qor-lang-selector, .qor-locale-selector').on('change', function () {
    var url = $(this).val();

    if (url) {
      window.location.assign(url);
    }
  });

  $('.qor-search').each(function () {
    var $this = $(this);
    var $input = $this.find('.qor-search-input');
    var $clear = $this.find('.qor-search-clear');

    $clear.on('click', function () {
      $this.removeClass('is-dirty');

      if ($input.val()) {
        $input.focus().val('');
      }
    });
  });
});
