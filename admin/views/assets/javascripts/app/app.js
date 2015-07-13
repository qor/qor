$(function () {

  'use strict';

  // Add Bootstrap's classes dynamically
  $('.qor-locale-selector').on('change', function () {
    var url = $(this).val();

    if (url) {
      window.location.assign(url);
    }
  });

  // Toggle submenus
  $('.qor-menu-group').on('click', '> ul > li > a', function () {
    var $next = $(this).next();

    if ($next.is('ul') && $next.css('position') !== 'absolute') {
      if (!$next.hasClass('collapsable')) {
        $next.addClass('collapsable').height($next.prop('scrollHeight'));
      }

      if ($next.hasClass('collapsed')) {
        $next.height($next.prop('scrollHeight'));

        setTimeout(function () {
          $next.removeClass('collapsed');
        }, 350);
      } else {
        $next.addClass('collapsed').height(0);
      }
    }
  });

  $('.qor-menu > li').each(function () {
    var $this = $(this),
        $ul = $this.find('> ul');

    if (!$ul.length) {
      $this.append($ul = $('<ul class="qor-menu"></ul>'));
    }

    $ul.attr('data-menu', $this.data('menuItem'));
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
