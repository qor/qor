$(function () {

  'use strict';

  var $form = $('.qor-form-container > form');

  $('.qor-error > li > label').each(function () {
    var $label = $(this);
    var $input = $form.find('#' + $label.attr('for'));

    if ($input.length) {
      $input.closest('.form-group').addClass('has-error').append($label.clone().addClass('mdl-textfield__error'));
    }
  });

});
