// init for slideout after show event
$.fn.qorSliderAfterShow = {};

// change Mustache tags from {{}} to [[]]
window.Mustache.tags = ['[[', ']]'];

// Init for date time picker
$('[data-toggle="qor.datetimepicker"]').materialDatePicker({ format : 'YYYY-MM-DD HH:mm' });
$('[data-toggle="qor.datepicker"]').materialDatePicker({ format : 'YYYY-MM-DD', time: false });
