(function (factory) {
  if (typeof define === 'function' && define.amd) {
    // AMD. Register as anonymous module.
    define(['jquery'], factory);
  } else if (typeof exports === 'object') {
    // Node / CommonJS
    factory(require('jquery'));
  } else {
    // Browser globals.
    factory(jQuery);
  }
})(function ($) {

  'use strict';

  var NAMESPACE = 'qor.action';
  var EVENT_ENABLE = 'enable.' + NAMESPACE;
  var EVENT_DISABLE = 'disable.' + NAMESPACE;
  var EVENT_CLICK = 'click.' + NAMESPACE;
  var ACTION_FORMS = '.qor-action-forms';
  var ACTION_HEADER = '.qor-page__header';
  var ACTION_SELECTORS = '.qor-actions';
  var BUTTON_BULKS = '.qor-action-bulk-buttons';
  var QOR_TABLE = '.qor-table-container';
  var QOR_SEARCH = '.qor-search-container';

  function QorAction(element, options) {
    this.$element = $(element);
    this.$wrap = $(ACTION_FORMS);
    this.options = $.extend({}, QorAction.DEFAULTS, $.isPlainObject(options) && options);
    this.init();
  }

  QorAction.prototype = {
    constructor: QorAction,

    init: function () {
      this.bind();
    },

    bind: function () {
      this.$element.on(EVENT_CLICK, $.proxy(this.click, this));
      $(document).on(EVENT_CLICK, '.qor-table--bulking tr', this.click);
    },

    unbind: function () {
      this.$element.off(EVENT_CLICK, this.check);
    },

    click : function (e) {
      var $target = $(e.target);

      if ($target.is('.qor-action--bulk span')) {
        this.$wrap.removeClass('hidden');
        $(BUTTON_BULKS).find('button').toggleClass('hidden');
        this.appendTableCheckbox();
        $(QOR_TABLE).addClass('qor-table--bulking');
        $(ACTION_HEADER).find(ACTION_SELECTORS).addClass('hidden');
        $(ACTION_HEADER).find(QOR_SEARCH).addClass('hidden');
      }

      if ($target.is('.qor-action--exit-bulk span')) {
        this.$wrap.addClass('hidden');
        $(BUTTON_BULKS).find('button').toggleClass('hidden');
        this.removeTableCheckbox();
        $(QOR_TABLE).removeClass('qor-table--bulking');
        $(ACTION_HEADER).find(ACTION_SELECTORS).removeClass('hidden');
        $(ACTION_HEADER).find(QOR_SEARCH).removeClass('hidden');
      }


      if ($(this).is('tr') && !$target.is('a')) {

        var $firstTd = $(this).find('td').first();

        // Manual make checkbox checked or not
        if ($firstTd.find('.mdl-checkbox__input').get(0)) {
          var $checkbox = $firstTd.find('.mdl-js-checkbox');
          var slideroutActionForm = $('[data-toggle="qor-action-slideout"]').find('form');
          var formValueInput = slideroutActionForm.find('.js-primary-value');
          var primaryValue = $(this).data('primary-key');
          var $alreadyHaveValue = formValueInput.filter('[value="' + primaryValue + '"]');

          $checkbox.toggleClass('is-checked');
          $firstTd.parents('tr').toggleClass('is-selected');

          var isChecked = $checkbox.hasClass('is-checked');

          $firstTd.find('input').prop('checked', isChecked);

          if (slideroutActionForm.size() && $('.qor-slideout').is(':visible')){

            if (isChecked && !$alreadyHaveValue.size()){
              slideroutActionForm.prepend('<input class="js-primary-value" type="hidden" name="primary_values[]" value="' + primaryValue + '" />');
            }

            if (!isChecked && $alreadyHaveValue.size()){
              $alreadyHaveValue.remove();
            }

          }

          return false;
        }

      }
    },

    destroy: function () {
      this.unbind();
      this.$element.removeData(NAMESPACE);
    },

    // Helper
    removeTableCheckbox : function () {
      $('.qor-page__body .mdl-data-table__select').each(function (i, e) { $(e).parents('td').remove(); });
      $('.qor-page__body .mdl-data-table__select').each(function (i, e) { $(e).parents('th').remove(); });
      $('.qor-table-container tr.is-selected').removeClass('is-selected');
      $('.qor-page__body table.mdl-data-table--selectable').removeClass('mdl-data-table--selectable');
      $('.qor-page__body tr.is-selected').removeClass('is-selected');
    },

    appendTableCheckbox : function () {
      // Only value change and the table isn't selectable will add checkboxes
      $('.qor-page__body .mdl-data-table__select').each(function (i, e) { $(e).parents('td').remove(); });
      $('.qor-page__body .mdl-data-table__select').each(function (i, e) { $(e).parents('th').remove(); });
      $('.qor-table-container tr.is-selected').removeClass('is-selected');

      $('.qor-page__body table').addClass('mdl-data-table--selectable');
      window.newQorMaterialDataTable = new window.MaterialDataTable($('.qor-page__body table').get(0));

      // The fixed head have checkbox but the visiual one doesn't, clone the head with checkbox from the fixed one
      $('thead.is-hidden tr th:not(".mdl-data-table__cell--non-numeric")').clone().prependTo($('thead:not(".is-hidden") tr'));

      // The clone one doesn't bind event, so binding event manual
      var $fixedHeadCheckBox = $('thead:not(".is-fixed") .mdl-checkbox__input').parents('label');
      $fixedHeadCheckBox.find('span').remove();
      window.newQorMaterialCheckbox = new window.MaterialCheckbox($fixedHeadCheckBox.get(0));
      $fixedHeadCheckBox.on('click', function () {
        $('thead.is-fixed tr th').eq(0).find('label').click();
        $(this).toggleClass('is-checked');

        var slideroutActionForm = $('[data-toggle="qor-action-slideout"]').find('form');
        var slideroutActionFormPrimaryValues = slideroutActionForm.find('.js-primary-value');

        slideroutActionFormPrimaryValues.remove();
        if ($(this).hasClass('is-checked') && slideroutActionForm.size() && $('.qor-slideout').is(':visible')){
          var allPrimaryValues = $('.qor-table--bulking tbody tr');
          allPrimaryValues.each(function () {
            var primaryValue = $(this).data('primary-key');
            if (primaryValue){
              slideroutActionForm.prepend('<input class="js-primary-value" type="hidden" name="primary_values[]" value="' + primaryValue + '" />');
            }
          });

        }

        return false;
      });
    }
  };

  QorAction.DEFAULTS = {
  };

  $.fn.qorSliderAfterShow.qorActionInit = function (url, html) {
    var hasAction = $(html).find('[data-toggle="qor-action-slideout"]').size();
    var $actionForm = $('[data-toggle="qor-action-slideout"]').find('form');
    var $checkedItem = $('.qor-page__body .mdl-checkbox__input:checked');

    if (hasAction && $checkedItem.size()){
      // insert checked value into sliderout form
      $checkedItem.each(function (i, e) {
        var id = $(e).parents('tbody tr').data('primary-key');
        id && $actionForm.prepend('<input class="js-primary-value" type="hidden" name="primary_values[]" value="' + id + '" />');
      });
    }

  };

  QorAction.plugin = function (options) {
    return this.each(function () {
      var $this = $(this);
      var data = $this.data(NAMESPACE);
      var fn;

      if (!data) {
        $this.data(NAMESPACE, (data = new QorAction(this, options)));
      }

      if (typeof options === 'string' && $.isFunction(fn = data[options])) {
        fn.call(data);
      }
    });
  };

  $(function () {
    var selector = '[data-toggle="qor.action.bulk"]';
    var options = {};

    $(document).
      on(EVENT_DISABLE, function (e) {
        QorAction.plugin.call($(selector, e.target), 'destroy');
      }).
      on(EVENT_ENABLE, function (e) {
        QorAction.plugin.call($(selector, e.target), options);
      }).
      triggerHandler(EVENT_ENABLE);
  });

  return QorAction;

});
