// init for slideout after show event
$.fn.qorSliderAfterShow = {};

// change Mustache tags from {{}} to [[]]
window.Mustache.tags = ['[[', ']]'];

// Init for date time picker
$('.qor-datetime-picker').materialDatePicker({ format : 'YYYY-MM-DD HH:mm' });
$('.qor-date-picker').materialDatePicker({ format : 'YYYY-MM-DD', time: false });

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
  var Mustache = window.Mustache;
  var NAMESPACE = 'qor.action';
  var EVENT_ENABLE = 'enable.' + NAMESPACE;
  var EVENT_DISABLE = 'disable.' + NAMESPACE;
  var EVENT_CLICK = 'click.' + NAMESPACE;
  var ACTION_FORMS = '.qor-action-forms';
  var ACTION_HEADER = '.qor-page__header';
  var ACTION_BODY = '.qor-page__body';
  var ACTION_BUTTON = '.qor-action-button';
  var MDL_BODY = '.mdl-layout__content';
  var ACTION_SELECTORS = '.qor-actions';
  var BUTTON_BULKS = '.qor-action-bulk-buttons';
  var QOR_TABLE = '.qor-table-container';
  var QOR_TABLE_BULK = '.qor-table--bulking';
  var QOR_SEARCH = '.qor-search-container';
  var QOR_SLIDEOUT = '.qor-slideout';

  var ACTION_FORM_DATA = 'primary_values[]';

  function QorAction(element, options) {
    this.$element = $(element);
    this.$wrap = $(ACTION_FORMS);
    this.options = $.extend({}, QorAction.DEFAULTS, $.isPlainObject(options) && options);
    this.ajaxForm = {};
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

    collectFormData: function () {
      var checkedInputs = $(QOR_TABLE_BULK).find('.mdl-checkbox__input:checked');
      var formData = [];

      if (checkedInputs.size()){
        checkedInputs.each(function () {
          var id = $(this).closest('tr').data('primary-key');
          if (id){
            formData.push({
              name: ACTION_FORM_DATA,
              value: id.toString()
            });
          }
        });
        this.ajaxForm.formData = formData;
      } else {
        this.ajaxForm.formData = [];
      }

      return this.ajaxForm;
    },

    click : function (e) {
      var $target = $(e.target);

      if ($target.data().ajaxForm) {
        this.collectFormData();
        this.ajaxForm.properties = $target.data();
        this.submit();
        return false;
      }

      if ($target.is('.qor-action--bulk')) {
        this.$wrap.removeClass('hidden');
        $(BUTTON_BULKS).find('button').toggleClass('hidden');
        this.appendTableCheckbox();
        $(QOR_TABLE).addClass('qor-table--bulking');
        $(ACTION_HEADER).find(ACTION_SELECTORS).addClass('hidden');
        $(ACTION_HEADER).find(QOR_SEARCH).addClass('hidden');
      }

      if ($target.is('.qor-action--exit-bulk')) {
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

    renderFlashMessage: function (data) {
      var flashMessageTmpl = QorAction.FLASHMESSAGETMPL;
      Mustache.parse(flashMessageTmpl);
      return Mustache.render(flashMessageTmpl, data);
    },

    submit: function () {
      var self = this;
      var $parent;

      var ajaxForm = this.ajaxForm;
      var properties = ajaxForm.properties;

      if (!ajaxForm.formData.length && properties.fromIndex){
        window.alert(ajaxForm.properties.errorNoProduct);
        return;
      }

      $.ajax(properties.url, {
        method: properties.method,
        data: ajaxForm.formData,
        dataType: properties.datatype,
        beforeSend: function () {
          self.$element.find(ACTION_BUTTON).attr('disabled', true);
        },
        success: function (data) {

          if (properties.fromIndex){
            window.location.reload();
          } else {
            self.$element.find(ACTION_BUTTON).attr('disabled', false);
            if ($(QOR_SLIDEOUT).is(':visible')){
              $parent = $(QOR_SLIDEOUT);
            } else {
              $parent = $(MDL_BODY);
            }
            $('.qor-alert').remove();
            $parent.find(ACTION_BODY).prepend(self.renderFlashMessage(data));
          }

        },
        error: function (xhr, textStatus, errorThrown) {
          self.$element.find(ACTION_BUTTON).attr('disabled', false);
          window.alert([textStatus, errorThrown].join(': '));
        }
      });
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

  QorAction.FLASHMESSAGETMPL = (
    '<div class="qor-alert qor-action-alert qor-alert--success [[#error]]qor-alert--error[[/error]]" [[#message]]data-dismissible="true"[[/message]] role="alert">' +
      '<button type="button" class="mdl-button mdl-button--icon" data-dismiss="alert">'  +
        '<i class="material-icons">close</i>'  +
      '</button>'  +
      '<span class="qor-alert-message">'  +
        '[[#message]]' +
          '[[message]]' +
        '[[/message]]' +
        '[[#error]]' +
          '[[error]]' +
        '[[/error]]' +
      '</span>'  +
    '</div>'
  );

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
        if (id){
          $actionForm.prepend('<input class="js-primary-value" type="hidden" name="primary_values[]" value="' + id + '" />');
        }
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

  var NAMESPACE = 'qor.autoheight';
  var EVENT_ENABLE = 'enable.' + NAMESPACE;
  var EVENT_DISABLE = 'disable.' + NAMESPACE;
  var EVENT_INPUT = 'input';

  function QorAutoheight(element, options) {
    this.$element = $(element);
    this.options = $.extend({}, QorAutoheight.DEFAULTS, $.isPlainObject(options) && options);
    this.init();
  }

  QorAutoheight.prototype = {
    constructor: QorAutoheight,

    init: function () {
      var $this = this.$element;

      this.overflow = $this.css('overflow');
      this.paddingTop = parseInt($this.css('padding-top'), 10);
      this.paddingBottom = parseInt($this.css('padding-bottom'), 10);
      $this.css('overflow', 'hidden');
      this.resize();
      this.bind();
    },

    bind: function () {
      this.$element.on(EVENT_INPUT, $.proxy(this.resize, this));
    },

    unbind: function () {
      this.$element.off(EVENT_INPUT, this.resize);
    },

    resize: function () {
      var $this = this.$element;

      if ($this.is(':hidden')) {
        return;
      }

      $this.height('auto').height($this.prop('scrollHeight') - this.paddingTop - this.paddingBottom);
    },

    destroy: function () {
      this.unbind();
      this.$element.css('overflow', this.overflow).removeData(NAMESPACE);
    },
  };

  QorAutoheight.DEFAULTS = {};

  QorAutoheight.plugin = function (options) {
    return this.each(function () {
      var $this = $(this);
      var data = $this.data(NAMESPACE);
      var fn;

      if (!data) {
        if (/destroy/.test(options)) {
          return;
        }

        $this.data(NAMESPACE, (data = new QorAutoheight(this, options)));
      }

      if (typeof options === 'string' && $.isFunction(fn = data[options])) {
        fn.apply(data);
      }
    });
  };

  $(function () {
    var selector = 'textarea.qor-js-autoheight';

    $(document).
      on(EVENT_DISABLE, function (e) {
        QorAutoheight.plugin.call($(selector, e.target), 'destroy');
      }).
      on(EVENT_ENABLE, function (e) {
        QorAutoheight.plugin.call($(selector, e.target));
      }).
      triggerHandler(EVENT_ENABLE);
  });

  return QorAutoheight;

});

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

  var NAMESPACE = 'qor.chooser';
  var EVENT_ENABLE = 'enable.' + NAMESPACE;
  var EVENT_DISABLE = 'disable.' + NAMESPACE;

  function QorChooser(element, options) {
    this.$element = $(element);
    this.options = $.extend({}, QorChooser.DEFAULTS, $.isPlainObject(options) && options);
    this.init();
  }

  QorChooser.prototype = {
    constructor: QorChooser,

    init: function () {
      var $this = this.$element;

      if (!$this.prop('multiple')) {
        if ($this.children('[selected]').length) {
          $this.prepend('<option value=""></option>');
        } else {
          $this.prepend('<option value="" selected></option>');
        }
      }

      $this.chosen({
        // jscs:disable requireCamelCaseOrUpperCaseIdentifiers
        allow_single_deselect: true,
        search_contains: true,
        disable_search_threshold: 10,
        width: '100%'
      });
    },

    destroy: function () {
      this.$element.chosen('destroy').removeData(NAMESPACE);
    },
  };

  QorChooser.DEFAULTS = {};

  QorChooser.plugin = function (options) {
    return this.each(function () {
      var $this = $(this);
      var data = $this.data(NAMESPACE);
      var fn;

      if (!data) {
        if (!$.fn.chosen) {
          return;
        }

        if (/destroy/.test(options)) {
          return;
        }

        $this.data(NAMESPACE, (data = new QorChooser(this, options)));
      }

      if (typeof options === 'string' && $.isFunction(fn = data[options])) {
        fn.apply(data);
      }
    });
  };

  $(function () {
    var selector = 'select[data-toggle="qor.chooser"]';

    $(document).
      on(EVENT_DISABLE, function (e) {
        QorChooser.plugin.call($(selector, e.target), 'destroy');
      }).
      on(EVENT_ENABLE, function (e) {
        QorChooser.plugin.call($(selector, e.target));
      }).
      triggerHandler(EVENT_ENABLE);
  });

  return QorChooser;

});

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

  var URL = window.URL || window.webkitURL;
  var NAMESPACE = 'qor.cropper';

  // Events
  var EVENT_ENABLE = 'enable.' + NAMESPACE;
  var EVENT_DISABLE = 'disable.' + NAMESPACE;
  var EVENT_CHANGE = 'change.' + NAMESPACE;
  var EVENT_CLICK = 'click.' + NAMESPACE;
  var EVENT_SHOWN = 'shown.qor.modal';
  var EVENT_HIDDEN = 'hidden.qor.modal';

  // Classes
  var CLASS_TOGGLE = '.qor-cropper__toggle';
  var CLASS_CANVAS = '.qor-cropper__canvas';
  var CLASS_WRAPPER = '.qor-cropper__wrapper';
  var CLASS_OPTIONS = '.qor-cropper__options';
  var CLASS_SAVE = '.qor-cropper__save';

  function capitalize(str) {
    if (typeof str === 'string') {
      str = str.charAt(0).toUpperCase() + str.substr(1);
    }

    return str;
  }

  function getLowerCaseKeyObject(obj) {
    var newObj = {};
    var key;

    if ($.isPlainObject(obj)) {
      for (key in obj) {
        if (obj.hasOwnProperty(key)) {
          newObj[String(key).toLowerCase()] = obj[key];
        }
      }
    }

    return newObj;
  }

  function getValueByNoCaseKey(obj, key) {
    var originalKey = String(key);
    var lowerCaseKey = originalKey.toLowerCase();
    var upperCaseKey = originalKey.toUpperCase();
    var capitalizeKey = capitalize(originalKey);

    if ($.isPlainObject(obj)) {
      return (obj[lowerCaseKey] || obj[capitalizeKey] || obj[upperCaseKey]);
    }
  }

  function replaceText(str, data) {
    if (typeof str === 'string') {
      if (typeof data === 'object') {
        $.each(data, function (key, val) {
          str = str.replace('${' + String(key).toLowerCase() + '}', val);
        });
      }
    }

    return str;
  }

  function QorCropper(element, options) {
    this.$element = $(element);
    this.options = $.extend(true, {}, QorCropper.DEFAULTS, $.isPlainObject(options) && options);
    this.data = null;
    this.init();
  }

  QorCropper.prototype = {
    constructor: QorCropper,

    init: function () {
      var options = this.options;
      var $this = this.$element;
      var $parent = $this.closest(options.parent);
      var $list;
      var data;

      if (!$parent.length) {
        $parent = $this.parent();
      }

      this.$parent = $parent;
      this.$output = $parent.find(options.output);
      this.$list = $list = $parent.find(options.list);

      if (!$list.find('img').attr('src')) {
        $list.find('ul').hide();
      }

      try {
        data = JSON.parse($.trim(this.$output.val()));
      } catch (e) {}

      this.data = data || {};
      this.build();
      this.bind();
    },

    build: function () {
      this.wrap();
      this.$modal = $(replaceText(QorCropper.MODAL, this.options.text)).appendTo('body');
    },

    unbuild: function () {
      this.$modal.remove();
      this.unwrap();
    },

    wrap: function () {
      var $list = this.$list;
      var $img;

      $list.find('li').append(QorCropper.TOGGLE);
      $img = $list.find('img');
      $img.wrap(QorCropper.CANVAS);
      this.center($img);
    },

    unwrap: function () {
      var $list = this.$list;

      $list.find(CLASS_TOGGLE).remove();
      $list.find(CLASS_CANVAS).each(function () {
        var $this = $(this);

        $this.before($this.html()).remove();
      });
    },

    bind: function () {
      this.$element.
        on(EVENT_CHANGE, $.proxy(this.read, this));

      this.$list.
        on(EVENT_CLICK, $.proxy(this.click, this));

      this.$modal.
        on(EVENT_SHOWN, $.proxy(this.start, this)).
        on(EVENT_HIDDEN, $.proxy(this.stop, this));
    },

    unbind: function () {
      this.$element.
        off(EVENT_CHANGE, this.read);

      this.$list.
        off(EVENT_CLICK, this.click);

      this.$modal.
        off(EVENT_SHOWN, this.start).
        off(EVENT_HIDDEN, this.stop);
    },

    click: function (e) {
      var target = e.target;
      var $target;

      if (target === this.$list[0]) {
        return;
      }

      $target = $(target);

      if (!$target.is('img')) {
        $target = $target.closest('li').find('img');
      }

      this.$target = $target;
      this.$modal.qorModal('show');
    },

    read: function (e) {
      var files = e.target.files;
      var file;

      if (files && files.length) {
        file = files[0];

        if (/^image\/\w+$/.test(file.type) && URL) {
          this.load(URL.createObjectURL(file));
        } else {
          this.$list.empty().text(file.name);
        }
      }
    },

    load: function (url) {
      var options = this.options;
      var _this = this;
      var $list = this.$list;
      var $ul = $list.find('ul');
      var data = this.data;
      var $image;

      if (!$ul.length) {
        $ul  = $(QorCropper.LIST);
        $list.html($ul);
        this.wrap();
      }

      $ul.show(); // show ul when it is hidden

      $image = $list.find('img');
      $image.one('load', function () {
        var $this = $(this);
        var naturalWidth = this.naturalWidth;
        var naturalHeight = this.naturalHeight;
        var sizeData = $this.data();
        var sizeResolution = sizeData.sizeResolution;
        var sizeName = sizeData.sizeName;
        var emulateImageData = {};
        var emulateCropData = {};
        var aspectRatio;
        var width;
        var height;

        if (sizeResolution) {
          width = getValueByNoCaseKey(sizeResolution, 'width');
          height = getValueByNoCaseKey(sizeResolution, 'height');
          aspectRatio = width / height;

          if (naturalHeight * aspectRatio > naturalWidth) {
            width = naturalWidth;
            height = width / aspectRatio;
          } else {
            height = naturalHeight;
            width = height * aspectRatio;
          }

          width *= 0.8;
          height *= 0.8;

          emulateImageData = {
            naturalWidth: naturalWidth,
            naturalHeight: naturalHeight,
          };

          emulateCropData = {
            x: Math.round((naturalWidth - width) / 2),
            y: Math.round((naturalHeight - height) / 2),
            width: Math.round(width),
            height: Math.round(height),
          };

          _this.preview($this, emulateImageData, emulateCropData);

          if (sizeName) {
            data.crop = true;

            if (!data[options.key]) {
              data[options.key] = {};
            }

            data[options.key][sizeName] = emulateCropData;
          }
        } else {
          _this.center($this);
        }

        _this.$output.val(JSON.stringify(data));
      }).attr('src', url).data('originalUrl', url);
    },

    start: function () {
      var options = this.options;
      var $modal = this.$modal;
      var $target = this.$target;
      var sizeData = $target.data();
      var sizeName = sizeData.sizeName || 'original';
      var sizeResolution = sizeData.sizeResolution;
      var $clone = $('<img>').attr('src', sizeData.originalUrl);
      var data = this.data;
      var _this = this;
      var sizeAspectRatio = NaN;
      var sizeWidth;
      var sizeHeight;
      var list;

      if (sizeResolution) {
        sizeWidth = getValueByNoCaseKey(sizeResolution, 'width');
        sizeHeight = getValueByNoCaseKey(sizeResolution, 'height');
        sizeAspectRatio = sizeWidth / sizeHeight;
      }

      if (!data[options.key]) {
        data[options.key] = {};
      }

      $modal.trigger('enable.qor.material').find(CLASS_WRAPPER).html($clone);

      list = this.getList(sizeAspectRatio);

      if (list) {
        $modal.find(CLASS_OPTIONS).show().append(list);
      }

      $clone.cropper({
        aspectRatio: sizeAspectRatio,
        data: getLowerCaseKeyObject(data[options.key][sizeName]),
        background: false,
        movable: false,
        zoomable: false,
        scalable: false,
        rotatable: false,
        checkImageOrigin: false,

        built: function () {
          $modal.find(CLASS_SAVE).one(EVENT_CLICK, function () {
            var cropData = $clone.cropper('getData', true);
            var syncData = [];
            var url;

            data.crop = true;
            data[options.key][sizeName] = cropData;
            _this.imageData = $clone.cropper('getImageData');
            _this.cropData = cropData;

            try {
              url = $clone.cropper('getCroppedCanvas').toDataURL();
            } catch (e) {}

            $modal.find(CLASS_OPTIONS + ' input').each(function () {
              var $this = $(this);

              if ($this.prop('checked')) {
                syncData.push($this.attr('name'));
              }
            });

            _this.output(url, syncData);
            $modal.qorModal('hide');
          });
        },
      });
    },

    stop: function () {
      this.$modal.
        trigger('disable.qor.material').
        find(CLASS_WRAPPER + ' > img').
          cropper('destroy').
          remove().
          end().
        find(CLASS_OPTIONS).
          hide().
          find('ul').
            remove();
    },

    getList: function (aspectRatio) {
      var list = [];

      this.$list.find('img').not(this.$target).each(function () {
        var data = $(this).data();
        var resolution = data.sizeResolution;
        var name = data.sizeName;
        var width;
        var height;

        if (resolution) {
          width = getValueByNoCaseKey(resolution, 'width');
          height = getValueByNoCaseKey(resolution, 'height');

          if (width / height === aspectRatio) {
            list.push(
              '<label>' +
                '<input type="checkbox" name="' + name + '" checked> ' +
                '<span>' + name +
                  '<small>(' + width + '&times;' + height + ' px)</small>' +
                '</span>' +
              '</label>'
            );
          }
        }
      });

      return list.length ? ('<ul><li>' + list.join('</li><li>') + '</li></ul>') : '';
    },

    output: function (url, data) {
      var $target = this.$target;

      if (url) {
        this.center($target.attr('src', url), true);
      } else {
        this.preview($target);
      }

      if ($.isArray(data) && data.length) {
        this.autoCrop(url, data);
      }

      this.$output.val(JSON.stringify(this.data));
    },

    preview: function ($target, emulateImageData, emulateCropData) {
      var $canvas = $target.parent();
      var $container = $canvas.parent();
      var containerWidth = $container.width();
      var containerHeight = $container.height();
      var imageData = emulateImageData || this.imageData;
      var cropData = $.extend({}, emulateCropData || this.cropData); // Clone one to avoid changing it
      var aspectRatio = cropData.width / cropData.height;
      var canvasWidth = containerWidth;
      var canvasHeight = containerHeight;
      var scaledRatio;

      if (containerHeight * aspectRatio > containerWidth) {
        canvasHeight = containerWidth / aspectRatio;
      } else {
        canvasWidth = containerHeight * aspectRatio;
      }

      scaledRatio = cropData.width / canvasWidth;

      $canvas.css({
        width: canvasWidth,
        height: canvasHeight,
      });

      $target.css({
        maxWidth: 'none',
        maxHeight: 'none',
        width: imageData.naturalWidth / scaledRatio,
        height: imageData.naturalHeight / scaledRatio,
        marginLeft: -cropData.x / scaledRatio,
        marginTop: -cropData.y / scaledRatio,
      });

      this.center($target);
    },

    center: function ($target, reset) {
      $target.each(function () {
        var $this = $(this);
        var $canvas = $this.parent();
        var $container = $canvas.parent();

        function center() {
          var containerHeight = $container.height();
          var canvasHeight = $canvas.height();
          var marginTop = 'auto';

          if (canvasHeight < containerHeight) {
            marginTop = (containerHeight - canvasHeight) / 2;
          }

          $canvas.css('margin-top', marginTop);
        }

        if (reset) {
          $canvas.add($this).removeAttr('style');
        }

        if (this.complete) {
          center.call(this);
        } else {
          this.onload = center;
        }
      });
    },

    autoCrop: function (url, data) {
      var cropData = this.cropData;
      var cropOptions = this.data[this.options.key];
      var _this = this;

      this.$list.find('img').not(this.$target).each(function () {
        var $this = $(this);
        var sizeName = $this.data('sizeName');

        if ($.inArray(sizeName, data) > -1) {
          cropOptions[sizeName] = $.extend({}, cropData);

          if (url) {
            _this.center($this.attr('src', url), true);
          } else {
            _this.preview($this);
          }
        }
      });
    },

    destroy: function () {
      this.unbind();
      this.unbuild();
      this.$element.removeData(NAMESPACE);
    },
  };

  QorCropper.DEFAULTS = {
    parent: false,
    output: false,
    list: false,
    key: 'data',
    data: null,
    text: {
      title: 'Crop the image',
      ok: 'OK',
      cancel: 'Cancel',
    },
  };

  QorCropper.TOGGLE = '<div class="qor-cropper__toggle"><i class="material-icons">crop</i></div>';
  QorCropper.CANVAS = '<div class="qor-cropper__canvas"></div>';
  QorCropper.LIST = '<ul><li><img></li></ul>';
  QorCropper.MODAL = (
    '<div class="qor-modal fade" tabindex="-1" role="dialog" aria-hidden="true">' +
      '<div class="mdl-card mdl-shadow--2dp" role="document">' +
        '<div class="mdl-card__title">' +
          '<h2 class="mdl-card__title-text">${title}</h2>' +
        '</div>' +
        '<div class="mdl-card__supporting-text">' +
          '<div class="qor-cropper__wrapper"></div>' +
          '<div class="qor-cropper__options">' +
            '<p>Sync cropping result to:</p>' +
          '</div>' +
        '</div>' +
        '<div class="mdl-card__actions mdl-card--border">' +
          '<a class="mdl-button mdl-button--colored mdl-js-button mdl-js-ripple-effect qor-cropper__save">${ok}</a>' +
          '<a class="mdl-button mdl-button--colored mdl-js-button mdl-js-ripple-effect" data-dismiss="modal">${cancel}</a>' +
        '</div>' +
        '<div class="mdl-card__menu">' +
          '<button class="mdl-button mdl-button--icon mdl-js-button mdl-js-ripple-effect" data-dismiss="modal" aria-label="close">' +
            '<i class="material-icons">close</i>' +
          '</button>' +
        '</div>' +
      '</div>' +
    '</div>'
  );

  QorCropper.plugin = function (option) {
    return this.each(function () {
      var $this = $(this);
      var data = $this.data(NAMESPACE);
      var options;
      var fn;

      if (!data) {
        if (!$.fn.cropper) {
          return;
        }

        if (/destroy/.test(option)) {
          return;
        }

        options = $.extend(true, {}, $this.data(), typeof option === 'object' && option);
        $this.data(NAMESPACE, (data = new QorCropper(this, options)));
      }

      if (typeof option === 'string' && $.isFunction(fn = data[option])) {
        fn.apply(data);
      }
    });
  };

  $(function () {
    var selector = '.qor-file__input';
    var options = {
          parent: '.qor-file',
          output: '.qor-file__options',
          list: '.qor-file__list',
          key: 'CropOptions',
        };

    $(document).
      on(EVENT_ENABLE, function (e) {
        QorCropper.plugin.call($(selector, e.target), options);
      }).
      on(EVENT_DISABLE, function (e) {
        QorCropper.plugin.call($(selector, e.target), 'destroy');
      }).
      triggerHandler(EVENT_ENABLE);
  });

  return QorCropper;

});

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

  var location = window.location;
  var NAMESPACE = 'qor.filter';
  var EVENT_ENABLE = 'enable.' + NAMESPACE;
  var EVENT_DISABLE = 'disable.' + NAMESPACE;
  var EVENT_CLICK = 'click.' + NAMESPACE;
  var EVENT_CHANGE = 'change.' + NAMESPACE;
  var CLASS_IS_ACTIVE = 'is-active';

  function QorFilter(element, options) {
    this.$element = $(element);
    this.options = $.extend({}, QorFilter.DEFAULTS, $.isPlainObject(options) && options);
    this.init();
  }

  function encodeSearch(data, detached) {
    var search = location.search;
    var params;

    if ($.isArray(data)) {
      params = decodeSearch(search);

      $.each(data, function (i, param) {
        i = $.inArray(param, params);

        if (i === -1) {
          params.push(param);
        } else if (detached) {
          params.splice(i, 1);
        }
      });

      search = '?' + params.join('&');
    }

    return search;
  }

  function decodeSearch(search) {
    var data = [];

    if (search && search.indexOf('?') > -1) {
      search = search.split('?')[1];

      if (search && search.indexOf('#') > -1) {
        search = search.split('#')[0];
      }

      if (search) {
        // search = search.toLowerCase();
        data = $.map(search.split('&'), function (n) {
          var param = [];
          var value;

          n = n.split('=');
          value = n[1];
          param.push(n[0]);

          if (value) {
            value = $.trim(decodeURIComponent(value));

            if (value) {
              param.push(value);
            }
          }

          return param.join('=');
        });
      }
    }

    return data;
  }

  QorFilter.prototype = {
    constructor: QorFilter,

    init: function () {
      // this.parse();
      this.bind();
    },

    bind: function () {
      var options = this.options;

      this.$element.
        on(EVENT_CLICK, options.label, $.proxy(this.toggle, this)).
        on(EVENT_CHANGE, options.group, $.proxy(this.toggle, this));
    },

    unbind: function () {
      this.$element.
        off(EVENT_CLICK, this.toggle).
        off(EVENT_CHANGE, this.toggle);
    },

    toggle: function (e) {
      var $target = $(e.currentTarget);
      var data = [];
      var params;
      var param;
      var search;
      var name;
      var value;
      var index;
      var matched;

      if ($target.is('select')) {
        params = decodeSearch(location.search);
        name = $target.attr('name');
        value = $target.val();

        param = [name];

        if (value) {
          param.push(value);
        }

        param = param.join('=');

        if (value) {
          data.push(param);
        }

        $target.children().each(function () {
          var $this = $(this);
          var param = [name];
          var value = $.trim($this.prop('value'));

          if (value) {
            param.push(value);
          }

          param = param.join('=');
          index = $.inArray(param, params);

          if (index > -1) {
            matched = param;
            return false;
          }
        });

        if (matched) {
          data.push(matched);
          search = encodeSearch(data, true);
        } else {
          search = encodeSearch(data);
        }
      } else if ($target.is('a')) {
        e.preventDefault();
        data = decodeSearch($target.attr('href'));

        if ($target.hasClass(CLASS_IS_ACTIVE)) {
          search = encodeSearch(data, true); // set `true` to detach
        } else {
          search = encodeSearch(data);
        }
      }

      location.search = search;
    },

    destroy: function () {
      this.unbind();
      this.$element.removeData(NAMESPACE);
    },
  };

  QorFilter.DEFAULTS = {
    label: false,
    group: false
  };

  QorFilter.plugin = function (options) {
    return this.each(function () {
      var $this = $(this);
      var data = $this.data(NAMESPACE);
      var fn;

      if (!data) {
        if (/destroy/.test(options)) {
          return;
        }

        $this.data(NAMESPACE, (data = new QorFilter(this, options)));
      }

      if (typeof options === 'string' && $.isFunction(fn = data[options])) {
        fn.apply(data);
      }
    });
  };

  $(function () {
    var selector = '[data-toggle="qor.filter"]';
    var options = {
          label: 'a',
          group: 'select',
        };

    $(document).
      on(EVENT_DISABLE, function (e) {
        QorFilter.plugin.call($(selector, e.target), 'destroy');
      }).
      on(EVENT_ENABLE, function (e) {
        QorFilter.plugin.call($(selector, e.target), options);
      }).
      triggerHandler(EVENT_ENABLE);
  });

  return QorFilter;

});

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

  var NAMESPACE = 'qor.fixer';
  var EVENT_ENABLE = 'enable.' + NAMESPACE;
  var EVENT_DISABLE = 'disable.' + NAMESPACE;
  var EVENT_CLICK = 'click.' + NAMESPACE;
  var EVENT_RESIZE = 'resize.' + NAMESPACE;
  var EVENT_SCROLL = 'scroll.' + NAMESPACE;
  var CLASS_IS_HIDDEN = 'is-hidden';
  var CLASS_IS_FIXED = 'is-fixed';

  function QorFixer(element, options) {
    this.$element = $(element);
    this.options = $.extend({}, QorFixer.DEFAULTS, $.isPlainObject(options) && options);
    this.$clone = null;
    this.init();
  }

  QorFixer.prototype = {
    constructor: QorFixer,

    init: function () {
      var options = this.options;
      var $this = this.$element;

      // disable fixer if have multiple tables
      if ($('.qor-page__body .qor-js-table').size() > 1) {
        return;
      }

      if ($this.is(':hidden') || $this.find('tbody > tr:visible').length <= 1) {
        return;
      }

      this.$thead = $this.find('thead:first');
      this.$tbody = $this.find('tbody:first');
      this.$header = $(options.header);
      this.$subHeader = $(options.subHeader);
      this.$content = $(options.content);
      this.marginBottomPX = parseInt(this.$subHeader.css('marginBottom'));
      this.paddingHeight = options.paddingHeight;

      this.resize();
      this.bind();
    },

    bind: function () {
      this.$element.on(EVENT_CLICK, $.proxy(this.check, this));

      this.$content.
        on(EVENT_SCROLL, $.proxy(this.toggle, this)).
        on(EVENT_RESIZE, $.proxy(this.resize, this));
    },

    unbind: function () {
      this.$element.off(EVENT_CLICK, this.check);

      this.$content.
        off(EVENT_SCROLL, this.toggle).
        off(EVENT_RESIZE, this.resize);
    },

    build: function () {
      var $this = this.$element;
      var $thead = this.$thead;
      var $clone = this.$clone;
      var $items = $thead.find('> tr').children();

      if (!$clone) {
        this.$clone = $clone = $thead.clone().prependTo($this);
      }

      $clone.
        addClass([CLASS_IS_FIXED, CLASS_IS_HIDDEN].join(' ')).
        find('> tr').
          children().
            each(function (i) {
              $(this).width($items.eq(i).width());
            });
    },

    unbuild: function () {
      this.$clone.remove();
    },

    check: function (e) {
      var $target = $(e.target);
      var checked;

      if ($target.is('.qor-js-check-all')) {
        checked = $target.prop('checked');

        $target.
          closest('thead').
          siblings('thead').
            find('.qor-js-check-all').prop('checked', checked).
            closest('.mdl-checkbox').toggleClass('is-checked', checked);
      }
    },

    toggle: function () {
      var $this = this.$element;
      var $clone = this.$clone;
      var theadHeight = this.$thead.outerHeight();
      var tbodyLastRowHeight = this.$tbody.find('tr:last').outerHeight();
      var scrollTop = this.$content.scrollTop();
      var minTop = 0;
      var maxTop = $this.outerHeight() - theadHeight - tbodyLastRowHeight;
      var offsetTop = this.$subHeader.outerHeight() + this.paddingHeight + this.marginBottomPX;
      var headerHeight = $('.qor-page__header').outerHeight();
      var showTop = Math.min(scrollTop - offsetTop, maxTop) + headerHeight;

      if (scrollTop > offsetTop - headerHeight) {
        $clone.css('top', showTop).removeClass(CLASS_IS_HIDDEN);
      } else {
        $clone.css('top', minTop).addClass(CLASS_IS_HIDDEN);
      }
    },

    resize: function () {
      this.build();
      this.toggle();
    },

    destroy: function () {
      this.unbind();
      this.unbuild();
      this.$element.removeData(NAMESPACE);
    },
  };

  QorFixer.DEFAULTS = {
    header: false,
    content: false,
  };

  QorFixer.plugin = function (options) {
    return this.each(function () {
      var $this = $(this);
      var data = $this.data(NAMESPACE);
      var fn;

      if (!data) {
        $this.data(NAMESPACE, (data = new QorFixer(this, options)));
      }

      if (typeof options === 'string' && $.isFunction(fn = data[options])) {
        fn.call(data);
      }
    });
  };

  $(function () {
    var selector = '.qor-js-table';
    var options = {
          header: '.mdl-layout__header',
          subHeader: '.qor-page__header',
          content: '.mdl-layout__content',
          paddingHeight: 2, // Fix sub header height bug
        };

    $(document).
      on(EVENT_DISABLE, function (e) {
        QorFixer.plugin.call($(selector, e.target), 'destroy');
      }).
      on(EVENT_ENABLE, function (e) {
        QorFixer.plugin.call($(selector, e.target), options);
      }).
      triggerHandler(EVENT_ENABLE);
  });

  return QorFixer;

});

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

  var NAMESPACE = 'qor.material';
  var EVENT_ENABLE = 'enable.' + NAMESPACE;
  var EVENT_DISABLE = 'disable.' + NAMESPACE;
  var EVENT_UPDATE = 'update.' + NAMESPACE;
  var SELECTOR_COMPONENT = '[class*="mdl-js"]';

  function enable(target) {

    /*jshint undef:false */
    if (componentHandler) {

      // Enable all MDL (Material Design Lite) components within the target element
      if ($(target).is(SELECTOR_COMPONENT)) {
        componentHandler.upgradeElements(target);
      } else {
        componentHandler.upgradeElements($(SELECTOR_COMPONENT, target).toArray());
      }
    }
  }

  function disable(target) {

    /*jshint undef:false */
    if (componentHandler) {

      // Destroy all MDL (Material Design Lite) components within the target element
      if ($(target).is(SELECTOR_COMPONENT)) {
        componentHandler.downgradeElements(target);
      } else {
        componentHandler.downgradeElements($(SELECTOR_COMPONENT, target).toArray());
      }
    }
  }

  $(function () {
    $(document).
      on(EVENT_ENABLE, function (e) {
        enable(e.target);
      }).
      on(EVENT_DISABLE, function (e) {
        disable(e.target);
      }).
      on(EVENT_UPDATE, function (e) {
        disable(e.target);
        enable(e.target);
      });
  });

});

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

  var $document = $(document);
  var NAMESPACE = 'qor.modal';
  var EVENT_ENABLE = 'enable.' + NAMESPACE;
  var EVENT_DISABLE = 'disable.' + NAMESPACE;
  var EVENT_CLICK = 'click.' + NAMESPACE;
  var EVENT_KEYUP = 'keyup.' + NAMESPACE;
  var EVENT_SHOW = 'show.' + NAMESPACE;
  var EVENT_SHOWN = 'shown.' + NAMESPACE;
  var EVENT_HIDE = 'hide.' + NAMESPACE;
  var EVENT_HIDDEN = 'hidden.' + NAMESPACE;
  var EVENT_TRANSITION_END = 'transitionend';
  var CLASS_OPEN = 'qor-modal-open';
  var CLASS_SHOWN = 'shown';
  var CLASS_FADE = 'fade';
  var CLASS_IN = 'in';
  var ARIA_HIDDEN = 'aria-hidden';

  function QorModal(element, options) {
    this.$element = $(element);
    this.options = $.extend({}, QorModal.DEFAULTS, $.isPlainObject(options) && options);
    this.transitioning = false;
    this.fadable = false;
    this.init();
  }

  QorModal.prototype = {
    constructor: QorModal,

    init: function () {
      this.fadable = this.$element.hasClass(CLASS_FADE);

      if (this.options.show) {
        this.show();
      } else {
        this.toggle();
      }
    },

    bind: function () {
      this.$element.on(EVENT_CLICK, $.proxy(this.click, this));

      if (this.options.keyboard) {
        $document.on(EVENT_KEYUP, $.proxy(this.keyup, this));
      }
    },

    unbind: function () {
      this.$element.off(EVENT_CLICK, this.click);

      if (this.options.keyboard) {
        $document.off(EVENT_KEYUP, this.keyup);
      }
    },

    click: function (e) {
      var element = this.$element[0];
      var target = e.target;

      if (target === element && this.options.backdrop) {
        this.hide();
        return;
      }

      while (target !== element) {
        if ($(target).data('dismiss') === 'modal') {
          this.hide();
          break;
        }

        target = target.parentNode;
      }
    },

    keyup: function (e) {
      if (e.which === 27) {
        this.hide();
      }
    },

    show: function (noTransition) {
      var $this = this.$element,
          showEvent;

      if (this.transitioning || $this.hasClass(CLASS_IN)) {
        return;
      }

      showEvent = $.Event(EVENT_SHOW);
      $this.trigger(showEvent);

      if (showEvent.isDefaultPrevented()) {
        return;
      }

      $document.find('body').addClass(CLASS_OPEN);

      /*jshint expr:true */
      $this.addClass(CLASS_SHOWN).scrollTop(0).get(0).offsetHeight; // reflow for transition
      this.transitioning = true;

      if (noTransition || !this.fadable) {
        $this.addClass(CLASS_IN);
        this.shown();
        return;
      }

      $this.one(EVENT_TRANSITION_END, $.proxy(this.shown, this));
      $this.addClass(CLASS_IN);
    },

    shown: function () {
      this.transitioning = false;
      this.bind();
      this.$element.attr(ARIA_HIDDEN, false).trigger(EVENT_SHOWN).focus();
    },

    hide: function (noTransition) {
      var $this = this.$element,
          hideEvent;

      if (this.transitioning || !$this.hasClass(CLASS_IN)) {
        return;
      }

      hideEvent = $.Event(EVENT_HIDE);
      $this.trigger(hideEvent);

      if (hideEvent.isDefaultPrevented()) {
        return;
      }

      $document.find('body').removeClass(CLASS_OPEN);
      this.transitioning = true;

      if (noTransition || !this.fadable) {
        $this.removeClass(CLASS_IN);
        this.hidden();
        return;
      }

      $this.one(EVENT_TRANSITION_END, $.proxy(this.hidden, this));
      $this.removeClass(CLASS_IN);
    },

    hidden: function () {
      this.transitioning = false;
      this.unbind();
      this.$element.removeClass(CLASS_SHOWN).attr(ARIA_HIDDEN, true).trigger(EVENT_HIDDEN);
    },

    toggle: function () {
      if (this.$element.hasClass(CLASS_IN)) {
        this.hide();
      } else {
        this.show();
      }
    },

    destroy: function () {
      this.$element.removeData(NAMESPACE);
    },
  };

  QorModal.DEFAULTS = {
    backdrop: true,
    keyboard: true,
    show: true,
  };

  QorModal.plugin = function (options) {
    return this.each(function () {
      var $this = $(this);
      var data = $this.data(NAMESPACE);
      var fn;

      if (!data) {
        if (/destroy/.test(options)) {
          return;
        }

        $this.data(NAMESPACE, (data = new QorModal(this, options)));
      }

      if (typeof options === 'string' && $.isFunction(fn = data[options])) {
        fn.apply(data);
      }
    });
  };

  $.fn.qorModal = QorModal.plugin;

  $(function () {
    var selector = '.qor-modal';

    $(document).
      on(EVENT_CLICK, '[data-toggle="qor.modal"]', function () {
        var $this = $(this);
        var data = $this.data();
        var $target = $(data.target || $this.attr('href'));

        QorModal.plugin.call($target, $target.data(NAMESPACE) ? 'toggle' : data);
      }).
      on(EVENT_DISABLE, function (e) {
        QorModal.plugin.call($(selector, e.target), 'destroy');
      }).
      on(EVENT_ENABLE, function (e) {
        QorModal.plugin.call($(selector, e.target));
      });
  });

  return QorModal;

});

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

  var $window = $(window);
  var NAMESPACE = 'qor.redactor';
  var EVENT_ENABLE = 'enable.' + NAMESPACE;
  var EVENT_DISABLE = 'disable.' + NAMESPACE;
  var EVENT_CLICK = 'click.' + NAMESPACE;
  var EVENT_FOCUS = 'focus.' + NAMESPACE;
  var EVENT_BLUR = 'blur.' + NAMESPACE;
  var EVENT_IMAGE_UPLOAD = 'imageupload.' + NAMESPACE;
  var EVENT_IMAGE_DELETE = 'imagedelete.' + NAMESPACE;
  var EVENT_SHOWN = 'shown.qor.modal';
  var EVENT_HIDDEN = 'hidden.qor.modal';

  var CLASS_WRAPPER = '.qor-cropper__wrapper';
  var CLASS_SAVE = '.qor-cropper__save';

  function encodeCropData(data) {
    var nums = [];

    if ($.isPlainObject(data)) {
      $.each(data, function () {
        nums.push(arguments[1]);
      });
    }

    return nums.join();
  }

  function decodeCropData(data) {
    var nums = data && data.split(',');

    data = null;

    if (nums && nums.length === 4) {
      data = {
        x: Number(nums[0]),
        y: Number(nums[1]),
        width: Number(nums[2]),
        height: Number(nums[3])
      };
    }

    return data;
  }

  function capitalize (str) {
    if (typeof str === 'string') {
      str = str.charAt(0).toUpperCase() + str.substr(1);
    }

    return str;
  }

  function getCapitalizeKeyObject (obj) {
    var newObj = {},
        key;

    if ($.isPlainObject(obj)) {
      for (key in obj) {
        if (obj.hasOwnProperty(key)) {
          newObj[capitalize(key)] = obj[key];
        }
      }
    }

    return newObj;
  }

  function replaceText(str, data) {
    if (typeof str === 'string') {
      if (typeof data === 'object') {
        $.each(data, function (key, val) {
          str = str.replace('${' + String(key).toLowerCase() + '}', val);
        });
      }
    }

    return str;
  }

  function QorRedactor(element, options) {
    this.$element = $(element);
    this.options = $.extend(true, {}, QorRedactor.DEFAULTS, $.isPlainObject(options) && options);
    this.init();
  }

  QorRedactor.prototype = {
    constructor: QorRedactor,

    init: function () {
      var options = this.options;
      var $this = this.$element;
      var $parent = $this.closest(options.parent);

      if (!$parent.length) {
        $parent = $this.parent();
      }

      this.$parent = $parent;
      this.$button = $(QorRedactor.BUTTON);
      this.$modal = $(replaceText(QorRedactor.MODAL, options.text)).appendTo('body');
      this.bind();
    },

    bind: function () {
      var $parent = this.$parent;
      var click = $.proxy(this.click, this);

      this.$element.
        on(EVENT_IMAGE_UPLOAD, function (e, image) {
          $(image).on(EVENT_CLICK, click);
        }).
        on(EVENT_IMAGE_DELETE, function (e, image) {
          $(image).off(EVENT_CLICK, click);
        }).
        on(EVENT_FOCUS, function () {
          $parent.find('img').off(EVENT_CLICK, click).on(EVENT_CLICK, click);
        }).
        on(EVENT_BLUR, function () {
          $parent.find('img').off(EVENT_CLICK, click);
        });

      $window.on(EVENT_CLICK, $.proxy(this.removeButton, this));
    },

    unbind: function () {
      this.$element.
        off(EVENT_IMAGE_UPLOAD).
        off(EVENT_IMAGE_DELETE).
        off(EVENT_FOCUS).
        off(EVENT_BLUR);

      $window.off(EVENT_CLICK, this.removeButton);
    },

    click: function (e) {
      e.stopPropagation();
      setTimeout($.proxy(this.addButton, this, $(e.target)), 1);
    },

    addButton: function ($image) {
      this.$button.
        prependTo($image.parent()).
        off(EVENT_CLICK).
        one(EVENT_CLICK, $.proxy(this.crop, this, $image));
    },

    removeButton: function () {
      this.$button.off(EVENT_CLICK).detach();
    },

    crop: function ($image) {
      var options = this.options;
      var url = $image.attr('src');
      var originalUrl = url;
      var $clone = $('<img>');
      var $modal = this.$modal;

      if ($.isFunction(options.replace)) {
        originalUrl = options.replace(originalUrl);
      }

      $clone.attr('src', originalUrl);
      $modal.one(EVENT_SHOWN, function () {
        $clone.cropper({
          data: decodeCropData($image.attr('data-crop-options')),
          background: false,
          movable: false,
          zoomable: false,
          scalable: false,
          rotatable: false,
          checkImageOrigin: false,

          built: function () {
            $modal.find(CLASS_SAVE).one(EVENT_CLICK, function () {
              var cropData = $clone.cropper('getData', true);

              $.ajax(options.remote, {
                type: 'POST',
                contentType: 'application/json',
                data: JSON.stringify({
                  Url: url,
                  CropOptions: {
                    original: getCapitalizeKeyObject(cropData)
                  },
                  Crop: true
                }),
                dataType: 'json',

                success: function (response) {
                  if ($.isPlainObject(response) && response.url) {
                    $image.attr('src', response.url).attr('data-crop-options', encodeCropData(cropData)).removeAttr('style').removeAttr('rel');

                    if ($.isFunction(options.complete)) {
                      options.complete();
                    }

                    $modal.qorModal('hide');
                  }
                }
              });
            });
          },
        });
      }).one(EVENT_HIDDEN, function () {
        $clone.cropper('destroy').remove();
      }).qorModal('show').find(CLASS_WRAPPER).append($clone);
    },

    destroy: function () {
      this.unbind();
      this.$modal.qorModal('hide').remove();
      this.$element.removeData(NAMESPACE);
    },
  };

  QorRedactor.DEFAULTS = {
    remote: false,
    parent: false,
    toggle: false,
    replace: null,
    complete: null,
    text: {
      title: 'Crop the image',
      ok: 'OK',
      cancel: 'Cancel',
    },
  };

  QorRedactor.BUTTON = '<span class="qor-cropper__toggle--redactor" contenteditable="false">Crop</span>';
  QorRedactor.MODAL = (
    '<div class="qor-modal fade" tabindex="-1" role="dialog" aria-hidden="true">' +
      '<div class="mdl-card mdl-shadow--2dp" role="document">' +
        '<div class="mdl-card__title">' +
          '<h2 class="mdl-card__title-text">${title}</h2>' +
        '</div>' +
        '<div class="mdl-card__supporting-text">' +
          '<div class="qor-cropper__wrapper"></div>' +
        '</div>' +
        '<div class="mdl-card__actions mdl-card--border">' +
          '<a class="mdl-button mdl-button--colored mdl-js-button mdl-js-ripple-effect qor-cropper__save">${ok}</a>' +
          '<a class="mdl-button mdl-button--colored mdl-js-button mdl-js-ripple-effect" data-dismiss="modal">${cancel}</a>' +
        '</div>' +
        '<div class="mdl-card__menu">' +
          '<button class="mdl-button mdl-button--icon mdl-js-button mdl-js-ripple-effect" data-dismiss="modal" aria-label="close">' +
            '<i class="material-icons">close</i>' +
          '</button>' +
        '</div>' +
      '</div>' +
    '</div>'
  );

  QorRedactor.plugin = function (option) {
    return this.each(function () {
      var $this = $(this);
      var data = $this.data(NAMESPACE);
      var config;
      var fn;

      if (!data) {
        if (!$.fn.redactor) {
          return;
        }

        if (/destroy/.test(option)) {
          return;
        }

        $this.data(NAMESPACE, (data = {}));
        config = $this.data();

        $this.redactor({
          imageUpload: config.uploadUrl,
          fileUpload: config.uploadUrl,

          initCallback: function () {
            if (!config.cropUrl) {
              return;
            }

            $this.data(NAMESPACE, (data = new QorRedactor($this, {
              remote: config.cropUrl,
              text: config.text,
              parent: '.qor-field',
              toggle: '.qor-cropper__toggle--redactor',
              replace: function (url) {
                return url.replace(/\.\w+$/, function (extension) {
                  return '.original' + extension;
                });
              },
              complete: $.proxy(function () {
                this.code.sync();
              }, this)
            })));
          },

          focusCallback: function (/*e*/) {
            $this.triggerHandler(EVENT_FOCUS);
          },

          blurCallback: function (/*e*/) {
            $this.triggerHandler(EVENT_BLUR);
          },

          imageUploadCallback: function (/*image, json*/) {
            $this.triggerHandler(EVENT_IMAGE_UPLOAD, arguments[0]);
          },

          imageDeleteCallback: function (/*url, image*/) {
            $this.triggerHandler(EVENT_IMAGE_DELETE, arguments[1]);
          }
        });
      } else {
        if (/destroy/.test(option)) {
          $this.redactor('core.destroy');
        }
      }

      if (typeof option === 'string' && $.isFunction(fn = data[option])) {
        fn.apply(data);
      }
    });
  };

  $(function () {
    var selector = 'textarea[data-toggle="qor.redactor"]';

    $(document).
      on(EVENT_DISABLE, function (e) {
        QorRedactor.plugin.call($(selector, e.target), 'destroy');
      }).
      on(EVENT_ENABLE, function (e) {
        QorRedactor.plugin.call($(selector, e.target));
      }).
      triggerHandler(EVENT_ENABLE);
  });

  return QorRedactor;

});

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

  var NAMESPACE = 'qor.replicator';
  var EVENT_ENABLE = 'enable.' + NAMESPACE;
  var EVENT_DISABLE = 'disable.' + NAMESPACE;
  var EVENT_CLICK = 'click.' + NAMESPACE;
  var IS_TEMPLATE = 'is-template';

  function QorReplicator(element, options) {
    this.$element = $(element);
    this.options = $.extend({}, QorReplicator.DEFAULTS, $.isPlainObject(options) && options);
    this.index = 0;
    this.init();
  }

  QorReplicator.prototype = {
    constructor: QorReplicator,

    init: function () {
      var $this = this.$element;
      var options = this.options;
      var $all = $this.find(options.itemClass);
      var $template;
      this.isMultipleTemplate = $this.data().isMultiple;

      if (!$all.length) {
        return;
      }

      $template = $all.filter(options.newClass);

      if (!$template.length) {
        $template = $all.last();
      }

      // Should destroy all components here
      $template.trigger('disable');

      this.$template = $template;
      this.multipleTemplates = {};
      var $filteredTemplateHtml = $template.filter($this.children(options.childrenClass).children(options.newClass));

      if (this.isMultipleTemplate) {
        this.$template = $filteredTemplateHtml;
        $template.remove();
        if ($this.children(options.childrenClass).children(options.itemClass).size()){
          this.template = $filteredTemplateHtml.prop('outerHTML');
          this.parse();
        }
      } else {
        this.template = $template.filter($this.children(options.childrenClass).children(options.newClass)).prop('outerHTML');
        $template.data(IS_TEMPLATE, true).hide();
        this.parse();
      }
      this.bind();
    },

    parse: function (hasIndex) {
      var i = 0;
      if (!this.template){
        return;
      }

      this.template = this.template.replace(/(\w+)\="(\S*\[\d+\]\S*)"/g, function (attribute, name, value) {
        value = value.replace(/^(\S*)\[(\d+)\]([^\[\]]*)$/, function (input, prefix, index, suffix) {
          if (input === value) {
            if (name === 'name') {
              i = index;
            }

            return (prefix + '[{{index}}]' + suffix);
          }
        });

        return (name + '="' + value + '"');
      });
      if (hasIndex) {
        return;
      }
      this.index = parseFloat(i);
    },

    bind: function () {
      var options = this.options;

      this.$element.
        on(EVENT_CLICK, options.addClass, $.proxy(this.add, this)).
        on(EVENT_CLICK, options.delClass, $.proxy(this.del, this));
    },

    unbind: function () {
      this.$element.
        off(EVENT_CLICK, this.add).
        off(EVENT_CLICK, this.del);
    },

    add: function (e) {
      var options = this.options;
      var self = this;
      var $target = $(e.target).closest(this.options.addClass);
      var templateName = $target.data().template;
      var parents = $target.closest(this.$element);
      var parentsChildren = parents.children(options.childrenClass);
      var $item = this.$template;

      // For multiple fieldset template
      if (this.isMultipleTemplate) {
        this.$template.each (function () {
          self.multipleTemplates[$(this).data().fieldsetName] = $(this);
        });
      }
      var $muptipleTargetTempalte = this.multipleTemplates[templateName];
      if (this.isMultipleTemplate){
        // For multiple template
        if ($target.length) {
          this.template = $muptipleTargetTempalte.prop('outerHTML');
          this.parse(true);
          $item = $(this.template.replace(/\{\{index\}\}/g, ++this.index));
          for (var dataKey in $target.data()) {
            if (dataKey.match(/^sync/)) {
              var k = dataKey.replace(/^sync/, '');
              $item.find('input[name*=\'.' + k + '\']').val($target.data(dataKey));
            }
          }
          if ($target.closest(options.childrenClass).children('fieldset').size()) {
            $target.closest(options.childrenClass).children('fieldset').last().after($item.show());
          } else {
            // If user delete all template
            parentsChildren.prepend($item.show());
          }
        }
      } else {
        // For single fieldset template
        if (this.$template && this.$template.filter(parentsChildren.children(options.newClass)).is(':hidden')) {
          this.$template.filter(parentsChildren.children(options.newClass)).show();
        } else {
          if ($target.length) {
            $item = $(this.template.replace(/\{\{index\}\}/g, ++this.index));
            $target.before($item.show());
          }
        }
      }

      if ($item) {
        // Enable all JavaScript components within the fieldset
        $item.trigger('enable');
      }
      e.stopPropagation();
    },

    del: function (e) {
      var options = this.options;
      var $item = $(e.target).closest(options.itemClass);
      var $alert;

      if ($item.is(options.newClass)) {
        // Destroy all JavaScript components within the fieldset
        $item.trigger('disable').remove();
      } else {
        $item.children(':visible').addClass('hidden').hide();
        $alert = $(options.alertTemplate.replace('{{name}}', this.parseName($item)));
        $alert.find(options.undoClass).one(EVENT_CLICK, function () {
          $alert.remove();
          $item.children('.hidden').removeClass('hidden').show();
        });
        $item.append($alert);
      }
    },

    parseName: function ($item) {
      var name = $item.find('input[name]').attr('name');

      if (name) {
        return name.replace(/[^\[\]]+$/, '');
      }
    },

    destroy: function () {
      this.unbind();
      this.$element.removeData(NAMESPACE);
    },
  };

  QorReplicator.DEFAULTS = {
    itemClass: false,
    newClass: false,
    addClass: false,
    delClass: false,
    childrenClass: false,
    alertTemplate: '',
  };

  QorReplicator.plugin = function (options) {
    return this.each(function () {
      var $this = $(this);
      var data = $this.data(NAMESPACE);
      var fn;

      if (!data) {
        $this.data(NAMESPACE, (data = new QorReplicator(this, options)));
      }

      if (typeof options === 'string' && $.isFunction(fn = data[options])) {
        fn.call(data);
      }
    });
  };

  $(function () {
    var selector = '.qor-fieldset-container';
    var options = {
          itemClass: '.qor-fieldset',
          newClass: '.qor-fieldset--new',
          addClass: '.qor-fieldset__add',
          delClass: '.qor-fieldset__delete',
          childrenClass: '.qor-field__block',
          undoClass: '.qor-fieldset__undo',
          alertTemplate: (
            '<div class="qor-fieldset__alert">' +
              '<input type="hidden" name="{{name}}._destroy" value="1">' +
              '<button class="mdl-button mdl-button--accent mdl-js-button mdl-js-ripple-effect qor-fieldset__undo" type="button">Undo delete</button>' +
            '</div>'
          ),
        };

    $(document).
      on(EVENT_DISABLE, function (e) {
        QorReplicator.plugin.call($(selector, e.target), 'destroy');
      }).
      on(EVENT_ENABLE, function (e) {
        QorReplicator.plugin.call($(selector, e.target), options);
      }).
      triggerHandler(EVENT_ENABLE);
  });

  return QorReplicator;

});

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
  var NAMESPACE = 'qor.globalSearch';
  var EVENT_ENABLE = 'enable.' + NAMESPACE;
  var EVENT_DISABLE = 'disable.' + NAMESPACE;
  var EVENT_CLICK = 'click.' + NAMESPACE;

  function QorSearchCenter(element, options) {
    this.$element = $(element);
    this.options = $.extend({}, QorSearchCenter.DEFAULTS, $.isPlainObject(options) && options);
    this.init();
  }

  QorSearchCenter.prototype = {
    constructor: QorSearchCenter,

    init: function () {
      this.bind();
    },

    bind: function () {
      this.$element.on(EVENT_CLICK, $.proxy(this.click, this));
    },

    unbind: function () {
      this.$element.off(EVENT_CLICK, this.check);
    },

    click : function (e) {
      var $target = $(e.target);
    },

    showGlobalSearch: function (e) {
      var $target = $(e.target);
      var searchData = $target.data();

      console.log($target)
      console.log(searchData)

      $body.append(this.renderTmpl(searchData));

      return false;
    },

    submit: function () {
      var self = this;
      var $parent;


      $.ajax(properties.url, {
        method: properties.method,
        data: ajaxForm.formData,
        dataType: properties.datatype,

        success: function (data) {



        },
        error: function (xhr, textStatus, errorThrown) {
          self.$element.find(ACTION_BUTTON).attr('disabled', false);
          window.alert([textStatus, errorThrown].join(': '));
        }
      });
    },

    destroy: function () {
      this.unbind();
      this.$element.removeData(NAMESPACE);
    }

  };

  QorSearchCenter.DEFAULTS = {
  };

  QorSearchCenter.plugin = function (options) {
    return this.each(function () {
      var $this = $(this);
      var data = $this.data(NAMESPACE);
      var fn;

      if (!data) {
        $this.data(NAMESPACE, (data = new QorSearchCenter(this, options)));
      }

      if (typeof options === 'string' && $.isFunction(fn = data[options])) {
        fn.call(data);
      }
    });
  };

  $(function () {
    var selector = '[data-toggle="qor.global.search"]';
    var options = {};

    $(document).
      on(EVENT_DISABLE, function (e) {
        QorSearchCenter.plugin.call($(selector, e.target), 'destroy');
      }).
      on(EVENT_ENABLE, function (e) {
        QorSearchCenter.plugin.call($(selector, e.target), options);
      }).
      triggerHandler(EVENT_ENABLE);
  });

  return QorSearchCenter;

});

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

  var $document = $(document);
  var NAMESPACE = 'qor.selector';
  var EVENT_ENABLE = 'enable.' + NAMESPACE;
  var EVENT_DISABLE = 'disable.' + NAMESPACE;
  var EVENT_CLICK = 'click.' + NAMESPACE;
  var CLASS_OPEN = 'open';
  var CLASS_ACTIVE = 'active';
  var CLASS_SELECTED = 'selected';
  var CLASS_DISABLED = 'disabled';
  var CLASS_CLEARABLE = 'clearable';
  var SELECTOR_SELECTED = '.' + CLASS_SELECTED;
  var SELECTOR_TOGGLE = '.qor-selector-toggle';
  var SELECTOR_LABEL = '.qor-selector-label';
  var SELECTOR_CLEAR = '.qor-selector-clear';
  var SELECTOR_MENU = '.qor-selector-menu';

  function QorSelector(element, options) {
    this.options = options;
    this.$element = $(element);
    this.init();
  }

  QorSelector.prototype = {
    constructor: QorSelector,

    init: function () {
      var $this = this.$element;

      this.placeholder = $this.attr('placeholder') || $this.attr('name') || 'Select';
      this.build();
    },

    build: function () {
      var $this = this.$element;
      var $selector = $(QorSelector.TEMPLATE);
      var alignedClass = this.options.aligned + '-aligned';
      var data = {};

      $selector.addClass(alignedClass).find(SELECTOR_MENU).html(function () {
        var list = [];

        $this.children().each(function () {
          var $this = $(this);
          var selected = $this.attr('selected');
          var disabled = $this.attr('disabled');
          var value = $this.attr('value');
          var label = $this.text();
          var classNames = [];

          if (selected) {
            classNames.push(CLASS_SELECTED);
            data.value = value;
            data.label = label;
          }

          if (disabled) {
            classNames.push(CLASS_DISABLED);
          }

          list.push(
            '<li' +
              (classNames.length ? ' class="' + classNames.join(' ') + '"' : '') +
              ' data-value="' + value + '"' +
              ' data-label="' + label + '"' +
            '>' +
              label +
            '</li>'
          );
        });

        return list.join('');
      });

      this.$selector = $selector;
      $this.hide().after($selector);
      this.pick(data, true);
      this.bind();
    },

    unbuild: function () {
      this.unbind();
      this.$selector.remove();
      this.$element.show();
    },

    bind: function () {
      this.$selector.on(EVENT_CLICK, $.proxy(this.click, this));
      $document.on(EVENT_CLICK, $.proxy(this.close, this));
    },

    unbind: function () {
      this.$selector.off(EVENT_CLICK, this.click);
      $document.off(EVENT_CLICK, this.close);
    },

    click: function (e) {
      var $target = $(e.target);

      e.stopPropagation();

      if ($target.is(SELECTOR_CLEAR)) {
        this.clear();
      } else if ($target.is('li')) {
        if (!$target.hasClass(CLASS_SELECTED) && !$target.hasClass(CLASS_DISABLED)) {
          this.pick($target.data());
        }

        this.close();
      } else if ($target.closest(SELECTOR_TOGGLE).length) {
        this.open();
      }
    },

    pick: function (data, initialized) {
      var $selector = this.$selector;
      var selected = !!data.value;

      $selector.
        find(SELECTOR_TOGGLE).
        toggleClass(CLASS_ACTIVE, selected).
        toggleClass(CLASS_CLEARABLE, selected && this.options.clearable).
          find(SELECTOR_LABEL).
          text(data.label || this.placeholder);

      if (!initialized) {
        $selector.
          find(SELECTOR_MENU).
            children('[data-value="' + data.value + '"]').
            addClass(CLASS_SELECTED).
            siblings(SELECTOR_SELECTED).
            removeClass(CLASS_SELECTED);

        this.$element.val(data.value).trigger('change');
      }
    },

    clear: function () {
      this.$selector.
        find(SELECTOR_TOGGLE).
        removeClass(CLASS_ACTIVE).
        removeClass(CLASS_CLEARABLE).
          find(SELECTOR_LABEL).
          text(this.placeholder).
          end().
        end().
        find(SELECTOR_MENU).
          children(SELECTOR_SELECTED).
          removeClass(CLASS_SELECTED);

      this.$element.val('').trigger('change');
    },

    open: function () {

      // Close other opened dropdowns first
      $document.triggerHandler(EVENT_CLICK);

      // Open the current dropdown
      this.$selector.addClass(CLASS_OPEN);
    },

    close: function () {
      this.$selector.removeClass(CLASS_OPEN);
    },

    destroy: function () {
      this.unbuild();
      this.$element.removeData(NAMESPACE);
    },
  };

  QorSelector.DEFAULTS = {
    aligned: 'left',
    clearable: false,
  };

  QorSelector.TEMPLATE = (
    '<div class="qor-selector">' +
      '<a class="qor-selector-toggle">' +
        '<span class="qor-selector-label"></span>' +
        '<i class="material-icons qor-selector-arrow">arrow_drop_down</i>' +
        '<i class="material-icons qor-selector-clear">clear</i>' +
      '</a>' +
      '<ul class="qor-selector-menu"></ul>' +
    '</div>'
  );

  QorSelector.plugin = function (option) {
    return this.each(function () {
      var $this = $(this);
      var data = $this.data(NAMESPACE);
      var options;
      var fn;

      if (!data) {
        if (/destroy/.test(option)) {
          return;
        }

        options = $.extend({}, QorSelector.DEFAULTS, $this.data(), typeof option === 'object' && option);
        $this.data(NAMESPACE, (data = new QorSelector(this, options)));
      }

      if (typeof option === 'string' && $.isFunction(fn = data[option])) {
        fn.apply(data);
      }
    });
  };

  $(function () {
    var selector = '[data-toggle="qor.selector"]';

    $(document).
      on(EVENT_DISABLE, function (e) {
        QorSelector.plugin.call($(selector, e.target), 'destroy');
      }).
      on(EVENT_ENABLE, function (e) {
        QorSelector.plugin.call($(selector, e.target));
      }).
      triggerHandler(EVENT_ENABLE);
  });

  return QorSelector;

});

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

  var $document = $(document);
  var FormData = window.FormData;
  var NAMESPACE = 'qor.slideout';
  var EVENT_ENABLE = 'enable.' + NAMESPACE;
  var EVENT_DISABLE = 'disable.' + NAMESPACE;
  var EVENT_KEYUP = 'keyup.' + NAMESPACE;
  var EVENT_CLICK = 'click.' + NAMESPACE;
  var EVENT_SUBMIT = 'submit.' + NAMESPACE;
  var EVENT_SHOW = 'show.' + NAMESPACE;
  var EVENT_SHOWN = 'shown.' + NAMESPACE;
  var EVENT_HIDE = 'hide.' + NAMESPACE;
  var EVENT_HIDDEN = 'hidden.' + NAMESPACE;
  var EVENT_TRANSITIONEND = 'transitionend';
  var CLASS_OPEN = 'qor-slideout-open';
  var CLASS_IS_SHOWN = 'is-shown';
  var CLASS_IS_SLIDED = 'is-slided';
  var CLASS_IS_SELECTED = 'is-selected';
  var CLASS_MAIN_CONTENT = '.mdl-layout__content.qor-page';
  var CLASS_HEADER_LOCALE = '.qor-actions__locale';

  function QorSlideout(element, options) {
    this.$element = $(element);
    this.options = $.extend({}, QorSlideout.DEFAULTS, $.isPlainObject(options) && options);
    this.slided = false;
    this.disabled = false;
    this.init();
  }

  QorSlideout.prototype = {
    constructor: QorSlideout,

    init: function () {
      this.build();
      this.bind();
    },

    build: function () {
      var $slideout;

      this.$slideout = $slideout = $(QorSlideout.TEMPLATE).appendTo('body');
      this.$title = $slideout.find('.qor-slideout__title');
      this.$body = $slideout.find('.qor-slideout__body');
    },

    unbuild: function () {
      this.$title = null;
      this.$body = null;
      this.$slideout.remove();
    },

    bind: function () {
      this.$slideout.
        on(EVENT_SUBMIT, 'form', $.proxy(this.submit, this));

      $document.
        on(EVENT_KEYUP, $.proxy(this.keyup, this)).
        on(EVENT_CLICK, $.proxy(this.click, this));
    },

    unbind: function () {
      this.$slideout.
        off(EVENT_SUBMIT, this.submit);

      $document.
        off(EVENT_KEYUP, this.keyup).
        off(EVENT_CLICK, this.click);
    },

    keyup: function (e) {
      if (e.which === 27) {
        this.hide();
      }
    },

    click: function (e) {
      var $this = this.$element;
      var slideout = this.$slideout.get(0);
      var target = e.target;
      var dismissible;
      var $target;
      var data;

      function toggleClass() {
        $this.find('tbody > tr').removeClass(CLASS_IS_SELECTED);
        $target.addClass(CLASS_IS_SELECTED);
      }

      if (e.isDefaultPrevented()) {
        return;
      }

      while (target !== document) {
        dismissible = false;
        $target = $(target);

        if ($target.prop('disabled')) {
          break;
        }

        if (target === slideout) {
          break;
        } else if ($target.data('dismiss') === 'slideout') {
          this.hide();
          break;
        } else if ($target.is('table.qor-table > tbody > tr[data-url]')) {
          if ($(e.target).parents('.qor-table__actions').size() > 0) {
            return;
          }
          // only load when not under loading and not activated
          if (!this.loading && !$target.hasClass(CLASS_IS_SELECTED)) {
            $this.one(EVENT_SHOW, toggleClass);
            data = $target.data();
            this.load(data.url);
          }
          break;
        } else if ($target.data('url')) {
          e.preventDefault();
          data = $target.data();
          this.load(data.url, data);
          break;
        } else {
          if ($target.is('a')) {
            break;
          }

          if (target) {
            target = target.parentNode;
          } else {
            break;
          }
        }
      }
    },

    submit: function (e) {
      var $slideout = this.$slideout;
      var $body = this.$body;
      var form = e.target;
      var $form = $(form);
      var _this = this;
      var $submit = $form.find(':submit');

      if (FormData) {
        e.preventDefault();

        $.ajax($form.prop('action'), {
          method: $form.prop('method'),
          data: new FormData(form),
          dataType: 'html',
          processData: false,
          contentType: false,
          beforeSend: function () {
            $submit.prop('disabled', true);
          },
          success: function (html) {
            var returnUrl = $form.data('returnUrl');

            if (returnUrl) {
              _this.load(returnUrl);
            } else {
              var prefix = '/' + location.pathname.split('/')[1];
              var flashStructs = [];
              $(html).find('.qor-alert').each(function (i, e) {
                var message = $(e).find('.qor-alert-message').text().trim();
                var type = $(e).data('type');
                if (message !== '') {
                  flashStructs.push({ Type: type, Message: message, Keep: true });
                }
              });
              if (flashStructs.length > 0) {
                document.cookie = 'qor-flashes=' + btoa(JSON.stringify(flashStructs)) + '; path=' + prefix;
              }
              _this.refresh();
            }
          },
          error: function (xhr, textStatus, errorThrown) {
            var $error;

            // Custom HTTP status code
            if (xhr.status === 422) {

              // Clear old errors
              $body.find('.qor-error').remove();
              $form.find('.qor-field').removeClass('is-error').find('.qor-field__error').remove();

              // Append new errors
              $error = $(xhr.responseText).find('.qor-error');
              $form.before($error);

              $error.find('> li > label').each(function () {
                var $label = $(this);
                var id = $label.attr('for');

                if (id) {
                  $form.find('#' + id).
                    closest('.qor-field').
                    addClass('is-error').
                    append($label.clone().addClass('qor-field__error'));
                }
              });

              // Scroll to top to view the errors
              $slideout.scrollTop(0);
            } else {
              window.alert([textStatus, errorThrown].join(': '));
            }
          },
          complete: function () {
            $submit.prop('disabled', false);
          },
        });
      }
    },

    load: function (url, data) {
      var options = this.options;
      var method;
      var dataType;
      var load;

      if (!url || this.loading) {
        return;
      }

      this.loading = true;
      data = $.isPlainObject(data) ? data : {};

      method = data.method ? data.method : 'GET';
      dataType = data.datatype ? data.datatype : 'html';

      data.url = data.method = data.datatype = data.ajaxForm = data.upgraded = undefined;

      load = $.proxy(function () {
        $.ajax(url, {
          method: method,
          data: data,
          dataType: dataType,
          success: $.proxy(function (response) {
            var $response;
            var $content;

            if (method === 'GET') {
              $response = $(response);

              $content = $response.find(CLASS_MAIN_CONTENT);

              if (!$content.length) {
                return;
              }

              $content.find('.qor-button--cancel').attr('data-dismiss', 'slideout').removeAttr('href');
              this.$title.html($response.find(options.title).html());
              this.$body.html($content.html());
              this.$body.find(CLASS_HEADER_LOCALE).remove();

              this.$slideout.one(EVENT_SHOWN, function () {

                // Enable all Qor components within the slideout
                $(this).trigger('enable');
              }).one(EVENT_HIDDEN, function () {

                // Destroy all Qor components within the slideout
                $(this).trigger('disable');

              });

              this.show();

              $('.qor-datetime-picker').materialDatePicker({ format : 'YYYY-MM-DD HH:mm' });
              $('.qor-date-picker').materialDatePicker({ format : 'YYYY-MM-DD', time: false });

              // callback for after slider loaded HTML
              if (options.afterShow){
                var qorSliderAfterShow = $.fn.qorSliderAfterShow;

                for (var name in qorSliderAfterShow) {
                  if (qorSliderAfterShow.hasOwnProperty(name)) {
                    qorSliderAfterShow[name].call(this, url, response);
                  }
                }
              }
            } else {
              if (data.returnUrl) {
                this.loading = false; // For reload
                this.load(data.returnUrl);
              } else {
                this.refresh();
              }
            }
          }, this),
          error: $.proxy (function (response) {
            var errors;
            if ($('.qor-error span').size() > 0) {
              errors = $('.qor-error span').map(function () {
                return $(this).text();
              }).get().join(', ');
            } else {
              errors = response.responseText;
            }
            window.alert(response.responseText);
          }, this),
          complete: $.proxy(function () {
            this.loading = false;
          }, this),
        });
      }, this);

      if (this.slided) {
        this.hide();
        this.$slideout.one(EVENT_HIDDEN, load);
      } else {
        load();
      }
    },

    show: function () {
      var $slideout = this.$slideout;
      var showEvent;

      if (this.slided) {
        return;
      }

      showEvent = $.Event(EVENT_SHOW);
      $slideout.trigger(showEvent);

      if (showEvent.isDefaultPrevented()) {
        return;
      }

      /*jshint expr:true */
      $slideout.addClass(CLASS_IS_SHOWN).get(0).offsetWidth;
      $slideout.
        one(EVENT_TRANSITIONEND, $.proxy(this.shown, this)).
        addClass(CLASS_IS_SLIDED).
        scrollTop(0);
    },

    shown: function () {
      this.slided = true;

      // Disable to scroll body element
      $('body').addClass(CLASS_OPEN);

      this.$slideout.trigger(EVENT_SHOWN);
    },

    hide: function () {
      var $slideout = this.$slideout;
      var hideEvent;
      var $datePicker = $('.qor-datepicker').not('.hidden');

      if ($datePicker.size()){
        $datePicker.addClass('hidden');
      }

      if (!this.slided) {
        return;
      }

      hideEvent = $.Event(EVENT_HIDE);
      $slideout.trigger(hideEvent);

      if (hideEvent.isDefaultPrevented()) {
        return;
      }

      $slideout.
        one(EVENT_TRANSITIONEND, $.proxy(this.hidden, this)).
        removeClass(CLASS_IS_SLIDED);
    },

    hidden: function () {
      this.slided = false;

      // Enable to scroll body element
      $('body').removeClass(CLASS_OPEN);

      this.$element.find('tbody > tr').removeClass(CLASS_IS_SELECTED);
      this.$slideout.removeClass(CLASS_IS_SHOWN).trigger(EVENT_HIDDEN);
    },

    refresh: function () {
      this.hide();

      setTimeout(function () {
        window.location.reload();
      }, 350);
    },

    toggle: function () {
      if (this.slided) {
        this.hide();
      } else {
        this.show();
      }
    },

    destroy: function () {
      this.unbind();
      this.unbuild();
      this.$element.removeData(NAMESPACE);
    },
  };

  QorSlideout.DEFAULTS = {
    title: false,
    content: false,
  };

  QorSlideout.TEMPLATE = (
    '<div class="qor-slideout">' +
      '<div class="qor-slideout__header">' +
        '<button type="button" class="mdl-button mdl-button--icon mdl-js-button mdl-js-repple-effect qor-slideout__close" data-dismiss="slideout">' +
          '<span class="material-icons">close</span>' +
        '</button>' +
        '<h3 class="qor-slideout__title"></h3>' +
      '</div>' +
      '<div class="qor-slideout__body"></div>' +
    '</div>'
  );

  QorSlideout.plugin = function (options) {
    return this.each(function () {
      var $this = $(this);
      var data = $this.data(NAMESPACE);
      var fn;

      if (!data) {
        if (/destroy/.test(options)) {
          return;
        }

        $this.data(NAMESPACE, (data = new QorSlideout(this, options)));
      }

      if (typeof options === 'string' && $.isFunction(fn = data[options])) {
        fn.apply(data);
      }
    });
  };

  $(function () {
    var selector = '.qor-theme-slideout';
    var options = {
          title: '.qor-form-title, .mdl-layout-title',
          afterShow: $.fn.qorSliderAfterShow ? $.fn.qorSliderAfterShow : null
        };

    $(document).
      on(EVENT_ENABLE, function (e) {

        if (/slideout/.test(e.namespace)) {
          QorSlideout.plugin.call($(selector, e.target), options);
        }
      }).
      on(EVENT_DISABLE, function (e) {

        if (/slideout/.test(e.namespace)) {
          QorSlideout.plugin.call($(selector, e.target), 'destroy');
        }
      }).
      triggerHandler(EVENT_ENABLE);
  });

  return QorSlideout;

});

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

  var location = window.location;
  var NAMESPACE = 'qor.sorter';
  var EVENT_ENABLE = 'enable.' + NAMESPACE;
  var EVENT_DISABLE = 'disable.' + NAMESPACE;
  var EVENT_CLICK = 'click.' + NAMESPACE;
  var CLASS_IS_SORTABLE = 'is-sortable';

  function QorSorter(element, options) {
    this.$element = $(element);
    this.options = $.extend({}, QorSorter.DEFAULTS, $.isPlainObject(options) && options);
    this.init();
  }

  QorSorter.prototype = {
    constructor: QorSorter,

    init: function () {
      this.$element.addClass(CLASS_IS_SORTABLE);
      this.bind();
    },

    bind: function () {
      this.$element.on(EVENT_CLICK, '> thead > tr > th', $.proxy(this.sort, this));
    },

    unbind: function () {
      this.$element.off(EVENT_CLICK, this.sort);
    },

    sort: function (e) {
      var $target = $(e.currentTarget);
      var orderBy = $target.data('orderBy');
      var search = location.search;
      var param = 'order_by=' + orderBy;

      // Stop when it is not sortable
      if (!orderBy) {
        return;
      }

      if (/order_by/.test(search)) {
        search = search.replace(/order_by(=\w+)?/, function () {
          return param;
        });
      } else {
        search += search.indexOf('?') > -1 ? ('&' + param) : param;
      }

      location.search = search;
    },

    destroy: function () {
      this.unbind();
      this.$element.removeClass(CLASS_IS_SORTABLE).removeData(NAMESPACE);
    },
  };

  QorSorter.DEFAULTS = {};

  QorSorter.plugin = function (options) {
    return this.each(function () {
      var $this = $(this);
      var data = $this.data(NAMESPACE);
      var fn;

      if (!data) {
        if (/destroy/.test(options)) {
          return;
        }

        $this.data(NAMESPACE, (data = new QorSorter(this, options)));
      }

      if (typeof options === 'string' && $.isFunction(fn = data[options])) {
        fn.apply(data);
      }
    });
  };

  $(function () {
    var selector = '.qor-js-table';

    $(document)
      .on(EVENT_DISABLE, function (e) {
        QorSorter.plugin.call($(selector, e.target), 'destroy');
      })
      .on(EVENT_ENABLE, function (e) {
        QorSorter.plugin.call($(selector, e.target));
      })
      .triggerHandler(EVENT_ENABLE);
  });

  return QorSorter;

});

//# sourceMappingURL=qor.js.map
