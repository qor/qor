$(function () {

  'use strict';

  $('.qor-menu-container').on('click', '> ul > li > a', function () {
    var $this = $(this);
    var $li = $this.parent();
    var $ul = $this.next('ul');

    if (!$ul.length) {
      return;
    }

    if ($ul.hasClass('in')) {
      $li.removeClass('is-expanded');
      $ul.one('transitionend', function () {
        $ul.removeClass('collapsing in');
      }).addClass('collapsing').height(0);
    } else {
      $li.addClass('is-expanded');
      $ul.one('transitionend', function () {
        $ul.removeClass('collapsing');
      }).addClass('collapsing in').height($ul.prop('scrollHeight'));
    }
  }).find('> ul > li > a').each(function () {
    var $this = $(this);
    var $li = $this.parent();
    var $ul = $this.next('ul');

    if (!$ul.length) {
      return;
    }

    $li.addClass('has-menu is-expanded');
    $ul.addClass('collapse in').height($ul.prop('scrollHeight'));
  });

  if ($('.qor-page').find('.qor-page__header').size()){
    $('.qor-page').addClass("has-header");
  }

});
