$(function () {

  'use strict';

  $("#dtBox").DateTimePicker({
    dateTimeFormat: "yyyy-MM-dd hh:mm",
    dateFormat: "yyyy-MM-dd",
    shortMonthNames: ["1","2","3","4","5","6","7","8","9","10","11","12"],
    fullMonthNames: ["1","2","3","4","5","6","7","8","9","10","11","12"],
    titleContentDateTime : "",
    setButtonContent : document.QorI18n.datetimePickerOKButton,
    incrementButtonContent: "add",
    decrementButtonContent: "remove",
    buttonsToDisplay: ["HeaderCloseButton", "SetButton"],
    formatHumanDate: function(oDate, sMode, sFormat){
      if(sMode === "date"){
        return oDate.yyyy + "-" + oDate.month + "-" + oDate.dd;
      }else if(sMode === "datetime"){
        return oDate.yyyy + "-" + oDate.month + "-" + oDate.dd + "<br>" + oDate.HH + ":" + oDate.mm;
      }
    }
  });

});
