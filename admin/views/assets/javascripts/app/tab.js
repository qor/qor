$(function () {

  'use strict';

  $('.qor-page__body .qor-js-action-tabs').on('click', '.mdl-tabs__tab', function() {
    var $scoped = $(this);
    $scoped.find('.mdl-tabs__tab').removeClass('is-active');
    $scoped.find('.mdl-tabs__panel').removeClass('is-active');
    $(this).addClass('is-active');
    $scoped.find($(this).attr('href')).addClass('is-active');
    var href = $('.mdl-tabs__tab.is-active').attr('href');
    location.hash = href.replace('-panel', '');
    return false;
  });

});
