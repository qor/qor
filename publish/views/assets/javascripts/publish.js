$(document).ready(function() {
  $(".draft-item").click(function() {
    var url = document.location.pathname + "/diff/" + $(this).attr("id");
    console.log(url)
    $(".reveal-modal-box").foundation('reveal', 'open', url);
  })
})
