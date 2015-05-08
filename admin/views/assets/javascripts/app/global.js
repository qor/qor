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
});
