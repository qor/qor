window.QorTab = {
  init : function() {
    this.initStatus();
    this.bindingEvents();
  },

  initStatus : function() {
    this.initTab();
  },

  bindingEvents : function() {
    $(".qor-js-action-tabs").on("click", ".qor-js-action-tab", this.switchTab);
  },

  initTab : function() {
    if(location.hash.match(/#tab-/)) {
      var $tab = $(".qor-js-action-tabs").find(".qor-js-action-tab[href='" + location.hash + "']");
      if($tab.get(0)) $.proxy(this.switchTab, $tab)();
    }
  },

  switchTab : function() {
    var $scoped = $(this).parents(".qor-js-action-tabs");
    $scoped.find('.qor-js-action-tab').removeClass('is-active');
    $scoped.find('.mdl-tabs__panel').removeClass('is-active');
    $(this).addClass('is-active');
    $scoped.find($(this).attr('href').replace("tab-", "")).addClass('is-active');
    location.hash = $('.qor-js-action-tab.is-active').attr('href');
    return false;
  }
}

$(document).ready(function() {
  window.QorTab.init();
});
