$(function () {

  'use strict';

  var location = window.location;
  var modal = (
    '<div class="qor-dialog qor-dialog--global-search" tabindex="-1" role="dialog" aria-hidden="true">' +
      '<div class="qor-dialog-content">' +
        '<form action=[[actionUrl]]>' +
          '<div class="mdl-textfield mdl-js-textfield" id="global-search-textfield">' +
            '<input class="mdl-textfield__input" name="keyword" id="globalSearch" value="" type="text" placeholder="" />' +
            '<label class="mdl-textfield__label" for="globalSearch">[[placeholder]]</label>' +
          '</div>' +
        '</form>' +
      '</div>' +
    '</div>'
  );

  $(document).on('click', '.qor-dialog--global-search', function(e){
    e.stopPropagation();
    if (!$(e.target).parents('.qor-dialog-content').size() && !$(e.target).is('.qor-dialog-content')){
      $('.qor-dialog--global-search').remove();
    }
  });

  $(document).on('click', '.qor-global-search--show', function(e){
      e.preventDefault();

      var data = $(this).data();
      var modalHTML = Mustache.render(modal, data);

      $('body').append(modalHTML);
      componentHandler.upgradeElement(document.getElementById('global-search-textfield'));
  });






  // $('.qor-search').each(function () {
  //   var $this = $(this);
  //   var $input = $this.find('.qor-search__input');
  //   var $clear = $this.find('.qor-search__clear');
  //   var isSearched = !!$input.val();

  //   $this.closest('.qor-page__header').addClass('has-search');

  //   $clear.on('click', function () {
  //     if ($input.val()) {
  //       $input.focus().val('');
  //     } else if (isSearched) {
  //       location.search = location.search.replace(new RegExp($input.attr('name') + '\\=?\\w*'), '');
  //     } else {
  //       $this.removeClass('is-dirty');
  //     }
  //   });
  // });

});
