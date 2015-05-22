$(function () {

  $('.dropdown.select .dropdown-option').on('click', function() {
    var text = $(this).text(),
        value = $(this).data('value'),
        $parent = $(this).parents('.dropdown');

    $parent.find('.selectedInput').val(value);
    $parent.find('.selected').text(text);

    var primaryLocale = $('.dropdown.select.origin .selectedInput').val(),
        toLocale = $('.dropdown.select.target .selectedInput').val(),
        path = '?primary_locale=' + primaryLocale + '&to_locale=' + toLocale;

    location.replace(path);
  });

  $('.translation-entry textarea').on('change', function() {
    var value = $.trim($(this).val());
    if (value) {
      var data = $($(this)[0].form).serializeArray();
      $.ajax({
        url: location.pathname,
        type: 'POST',
        cache: false,
        timeout: 7777,
        data: data
      });
    }
  });
});
