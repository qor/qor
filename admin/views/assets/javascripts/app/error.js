$(function () {

  'use strict';

  var $form = $('.qor-form-container > form');

  $('.qor-error > li > label').each(function (i) {
    var $label = $(this);
    var $input = $form.find('#' + $label.attr('for'));

    if ($input.length) {
      $input.after($label.clone().addClass('mdl-textfield__error'));
      $input.closest('.form-group').addClass('has-error');

      if (i === 0) {
        $input.focus();
      }
    }
  });

});
