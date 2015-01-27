$(document).ready(function() {
  $(".draft-item td").not(".selector").click(function() {
    var self = $(this).parents(".draft-item");
    var url = document.location.pathname + "/diff/" + self.attr("id");
    $(".reveal-modal-box").foundation('reveal', 'open', url);
  })
})
