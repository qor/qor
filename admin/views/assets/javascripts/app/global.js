$(function () {
  $('.table').each(function () {
    var self = this,
        $ths = $(this).find('.thr-inner .th');

    $ths.each(function () {
      var col = $(this).data('col'),
          wid = $(this).outerWidth();

      $(self).find('.tr-inner .' + col).outerWidth(wid);
    });
  });

  $('.grid-trigger-wrapper .trigger').on('click', function () {
    var state = $(this).attr('state');

    $('.table, table').attr('state', state);
    $('.grid-trigger-wrapper .trigger').removeClass('cur');
    $(this).addClass('cur');
  });

  $('.dropdown.select .dropdown-option').on('click', function() {
    var text = $(this).text(),
        value = $(this).data('value'),
        $parent = $(this).parents('.dropdown');

    $parent.find('.selectedInput').val(value);
    $parent.find('.selected').text(text);

  });
});
