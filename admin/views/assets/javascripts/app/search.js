$(function () {

  'use strict';

  var location = window.location;
  var a='a';

  $('.qor-search').each(function () {
    var $this = $(this);
    var $input = $this.find('.qor-search__input');
    var $clear = $this.find('.qor-search__clear');
    var isSearched = !!$input.val();

    $this.closest('.qor-page__header').addClass('has-search');

    $clear.on('click', function () {
      if ($input.val()) {
        $input.focus().val('');
      } else if (isSearched) {
        location.search = location.search.replace(new RegExp($input.attr('name') + '\\=?\\w*'), '').replace('?','');
      } else {
        $this.removeClass('is-dirty');
      }
    });
  });
});
