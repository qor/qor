$(document).ready(function() {
  $(".draft-item").click(function() {
    var url = document.location.pathname + "/diff/" + $(this).attr("id");
    $(this).foundation('reveal', 'open', url);
  })
})
