// init for slideout after show event
$.fn.qorSliderAfterShow = {};

// change Mustache tags from {{}} to [[]]
window.Mustache.tags = ['[[', ']]'];

// Init for date time picker
$('.qor-datetime-picker').materialDatePicker({ format : 'YYYY-MM-DD HH:mm' });
$('.qor-date-picker').materialDatePicker({ format : 'YYYY-MM-DD', time: false });
