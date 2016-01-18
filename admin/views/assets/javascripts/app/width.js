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
  if ($('.qor-js-table tbody').find('tr').size() > 6){
    $('td > .qor-button--actions').on('mouseover',function(){
        var $this = $(this);

        var tableHeight = $this.closest("table").height();
        var buttonTop = $this.closest("td").position().top;
        var $buttonDropdown = $this.closest("td").find('.mdl-menu');
        var isNeedChangePosition = (buttonTop + $buttonDropdown.outerHeight()) > (tableHeight * 0.8);

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
  }

});
