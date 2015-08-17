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
