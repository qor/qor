$(function () {

  'use strict';

  var $form = $('.qor-form-container > form');

  $('.qor-error > li > label').each(function () {
    var $label = $(this);

    $form.
      find('#' + $label.attr('for')).
      after($label.clone().addClass('mdl-textfield__error')).
      closest('.form-group').
      addClass('has-error');
  });

});
