$(document).ready(function() {
  $(".draft-item").click(function() {
    var url = document.location.href + "/diff/" + $(this).attr("id");
    $.ajax({url: url})
  })
})
