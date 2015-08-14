$(function () {

  'use strict';

  var location = window.location;
  var search = location.search;

  $('.qor-lang-selector, .qor-locale-selector').on('change', function () {
    var url = $(this).val();

    if (url) {
      location.assign(url);
    }
  });

  $('.qor-search').each(function () {
    var $this = $(this);
    var $input = $this.find('.qor-search__input');
    var $clear = $this.find('.qor-search__clear');
    var isSearched = !!$input.val();

    $clear.on('click', function (e) {
      if ($input.val()) {
        $input.focus().val('');
      } else if (isSearched) {
        location.search = search.replace(new RegExp($input.attr('name') + '\\=?\\w*'), '');
      } else {
        $this.removeClass('is-dirty');
      }
    });
  });
});
