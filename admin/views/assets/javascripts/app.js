$(function () {

  'use strict';

  $(document).on('click.qor.alert', '[data-dismiss="alert"]', function () {
    $(this).closest('.qor-alert').remove();
  });

  setTimeout(function () {
    $('.qor-alert[data-dismissible="true"]').remove();
  }, 5000);

});

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

$(function () {

  'use strict';

  var $form = $('.qor-page__body > .qor-form-container > form');

  $('.qor-error > li > label').each(function () {
    var $label = $(this);
    var id = $label.attr('for');

    if (id) {
      $form.find('#' + id).
        closest('.qor-field').
        addClass('is-error').
        append($label.clone().addClass('qor-field__error'));
    }
  });

});

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

});

$(function () {

  'use strict';

  var location = window.location;

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
        location.search = location.search.replace(new RegExp($input.attr('name') + '\\=?\\w*'), '');
      } else {
        $this.removeClass('is-dirty');
      }
    });
  });
});

$(function () {

  'use strict';

  $('.qor-js-table .qor-table__content').each(function () {
    var $this = $(this);
    var width = $this.width();
    var parentWidth = $this.parent().width();

    if (width >= 180 && width < parentWidth) {
      $this.css('max-width', parentWidth);
    }
  });

  // @Jason weng
  // Reset dropdown menu position in MDL Table
  // Normal is Bottom Right
  // If button top position + dropdown menu height > table height
  // will display dropdown as Top Right
  $('td > .qor-button--actions').on('mouseover',function(){
      var $this = $(this);
      var tableHeight = $this.closest("table").height();
      var buttonTop = $this.closest("td").position().top;
      var $buttonDropdown = $this.closest("td").find('.mdl-menu');
      var isNeedChangePosition = buttonTop + $buttonDropdown.height() > tableHeight * 0.9;

      var CLASS_TOP = 'mdl-menu--top-right';
      var CLASS_BOTTOM = 'mdl-menu--bottom-right';

      if (isNeedChangePosition){
        if ($this.hasClass(CLASS_TOP)){
          return;
        }
        $buttonDropdown.removeClass(CLASS_BOTTOM).addClass(CLASS_TOP);
      } else {
        if ($this.hasClass(CLASS_BOTTOM)){
          return;
        }
        $buttonDropdown.removeClass(CLASS_TOP).addClass(CLASS_BOTTOM);
      }
  });

});

//# sourceMappingURL=app.js.map
