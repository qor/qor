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

  $('td > .qor-button--actions').on('mouseover',function(){
      var $this = $(this);
      var viewHeight = $(window).height();
      var buttonOffsetTop = $this.offset().top;
      var $buttonDropdown = $this.closest("td").find('.mdl-menu');
      var CLASS_TOP = 'mdl-menu--top-right';
      var CLASS_BOTTOM = 'mdl-menu--bottom-right';

      if (buttonOffsetTop > viewHeight * 0.6){
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
