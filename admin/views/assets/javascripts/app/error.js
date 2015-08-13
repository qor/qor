$(function () {

  'use strict';

  var $form = $('.qor-form-container > form');

  $('.qor-error > li > label').each(function () {
    var $label = $(this);
    var $input = $form.find('#' + $label.attr('for'));

    if ($input.length) {
      $input.
        closest('.mdl-textfield, .qor-field').
        addClass('is-error').
        append('<span class="mdl-textfield__error"></span>').
          find('.mdl-textfield__error').
          html($label.html());
    }
  });

});
