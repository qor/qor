$.Redactor.fn.modal = function () {
  return {
    callbacks: {},
    loadTemplates: function()
    {
      this.opts.modal = {
        imageEdit: String()
        + '<section id="redactor-modal-image-edit">'
          + '<label class="redactor-image-link-option"><input type="checkbox" id="redactor-image-link-blank"> ' + this.lang.get('link_new_tab') + '</label>'
        + '</section>',

        image: String()
        + '<section id="redactor-modal-image-insert">'
          + '<div id="redactor-modal-image-droparea"></div>'
        + '</section>',

        file: String()
        + '<section id="redactor-modal-file-insert">'
          + '<div id="redactor-modal-file-upload-box">'
            + '<label>' + this.lang.get('filename') + '</label>'
            + '<input type="text" id="redactor-filename" /><br><br>'
            + '<div id="redactor-modal-file-upload"></div>'
          + '</div>'
        + '</section>',

        link: String()
        + '<section id="redactor-modal-link-insert">'
          + '<label>URL</label>'
          + '<input type="url" id="redactor-link-url" />'
          + '<label>' + this.lang.get('text') + '</label>'
          + '<input type="text" id="redactor-link-url-text" />'
          + '<label><input type="checkbox" id="redactor-link-blank"> ' + this.lang.get('link_new_tab') + '</label>'
        + '</section>'
      };


      $.extend(this.opts, this.opts.modal);

    },
    addCallback: function(name, callback)
    {
      this.modal.callbacks[name] = callback;
    },
    createTabber: function($modal)
    {
      this.modal.$tabber = $('<div>').attr('id', 'redactor-modal-tabber');

      $modal.prepend(this.modal.$tabber);
    },
    addTab: function(id, name, active)
    {
      var $tab = $('<a href="#" rel="tab' + id + '">').text(name);
      if (active)
      {
        $tab.addClass('active');
      }

      var self = this;
      $tab.on('click', function(e)
      {
        e.preventDefault();
        $('.redactor-tab').hide();
        $('.redactor-' + $(this).attr('rel')).show();

        self.modal.$tabber.find('a').removeClass('active');
        $(this).addClass('active');

      });

      this.modal.$tabber.append($tab);
    },
    addTemplate: function(name, template)
    {
      this.opts.modal[name] = template;
    },
    getTemplate: function(name)
    {
      return this.opts.modal[name];
    },
    getModal: function()
    {
      return this.$modalBody.find('section');
    },
    load: function(templateName, title, width)
    {
      this.modal.templateName = templateName;
      this.modal.width = width;

      this.modal.build();
      this.modal.enableEvents();
      this.modal.setTitle(title);
      this.modal.setDraggable();
      this.modal.setContent();

      // callbacks
      if (typeof this.modal.callbacks[templateName] != 'undefined')
      {
        this.modal.callbacks[templateName].call(this);
      }

    },
    show: function()
    {
      // ios keyboard hide
      if (this.utils.isMobile() && !this.utils.browser('msie'))
      {
        document.activeElement.blur();
      }

      $(document.body).removeClass('body-redactor-hidden');
      this.modal.bodyOveflow = $(document.body).css('overflow');
      $(document.body).css('overflow', 'hidden');

      if (this.utils.isMobile())
      {
        this.modal.showOnMobile();
      }
      else
      {
        this.modal.showOnDesktop();
      }

      this.$modalOverlay.show();
      this.$modalBox.show();

      this.modal.setButtonsWidth();

      this.utils.saveScroll();

      // resize
      if (!this.utils.isMobile())
      {
        setTimeout($.proxy(this.modal.showOnDesktop, this), 0);
        $(window).on('resize.redactor-modal', $.proxy(this.modal.resize, this));
      }

      // modal shown callback
      this.core.setCallback('modalOpened', this.modal.templateName, this.$modal);

      // fix bootstrap modal focus
      $(document).off('focusin.modal');

      // enter
      this.$modal.find('input[type=text],input[type=url],input[type=email]').on('keydown.redactor-modal', $.proxy(this.modal.setEnter, this));
    },
    showOnDesktop: function()
    {
      var height = this.$modal.outerHeight();
      var windowHeight = $(window).height();
      var windowWidth = $(window).width();

      if (this.modal.width > windowWidth)
      {
        this.$modal.css({
          width: '96%',
          marginTop: (windowHeight/2 - height/2) + 'px'
        });
        return;
      }

      if (height > windowHeight)
      {
        this.$modal.css({
          width: this.modal.width + 'px',
          marginTop: '20px'
        });
      }
      else
      {
        this.$modal.css({
          width: this.modal.width + 'px',
          marginTop: (windowHeight/2 - height/2) + 'px'
        });
      }
    },
    showOnMobile: function()
    {
      this.$modal.css({
        width: '96%',
        marginTop: '2%'
      });

    },
    resize: function()
    {
      if (this.utils.isMobile())
      {
        this.modal.showOnMobile();
      }
      else
      {
        this.modal.showOnDesktop();
      }
    },
    setTitle: function(title)
    {
      this.$modalHeader.html(title);
    },
    setContent: function()
    {
      this.$modalBody.html(this.modal.getTemplate(this.modal.templateName));
    },
    setDraggable: function()
    {
      if (typeof $.fn.draggable === 'undefined') return;

      this.$modal.draggable({ handle: this.$modalHeader });
      this.$modalHeader.css('cursor', 'move');
    },
    setEnter: function(e)
    {
      if (e.which != 13) return;

      e.preventDefault();
      this.$modal.find('button.redactor-modal-action-btn').click();
    },
    createCancelButton: function()
    {
      var button = $('<button>').addClass('redactor-modal-btn redactor-modal-close-btn').html(this.lang.get('cancel'));
      button.on('click', $.proxy(this.modal.close, this));

      this.$modalFooter.append(button);
    },
    createDeleteButton: function(label)
    {
      return this.modal.createButton(label, 'delete');
    },
    createActionButton: function(label)
    {
      return this.modal.createButton(label, 'action');
    },
    createButton: function(label, className)
    {
      var button = $('<button>').addClass('redactor-modal-btn').addClass('redactor-modal-' + className + '-btn').html(label);
      this.$modalFooter.append(button);

      return button;
    },
    setButtonsWidth: function()
    {
      var buttons = this.$modalFooter.find('button');
      var buttonsSize = buttons.length;
      if (buttonsSize === 0) return;

      buttons.css('width', (100/buttonsSize) + '%');
    },
    build: function()
    {
      this.modal.buildOverlay();

      this.$modalBox = $('<div id="redactor-modal-box" />').hide();
      this.$modal = $('<div id="redactor-modal" />');
      this.$modalHeader = $('<header />');
      this.$modalClose = $('<span id="redactor-modal-close" />').html('&times;');
      this.$modalBody = $('<div id="redactor-modal-body" />');
      this.$modalFooter = $('<footer />');

      this.$modal.append(this.$modalHeader);
      this.$modal.append(this.$modalClose);
      this.$modal.append(this.$modalBody);
      this.$modal.append(this.$modalFooter);
      this.$modalBox.append(this.$modal);
      this.$modalBox.appendTo(document.body);
    },
    buildOverlay: function()
    {
      this.$modalOverlay = $('<div id="redactor-modal-overlay">').hide();
      $('body').prepend(this.$modalOverlay);
    },
    enableEvents: function()
    {
      this.$modalClose.on('click.redactor-modal', $.proxy(this.modal.close, this));
      $(document).on('keyup.redactor-modal', $.proxy(this.modal.closeHandler, this));
      this.$editor.on('keyup.redactor-modal', $.proxy(this.modal.closeHandler, this));
      this.$modalBox.on('click.redactor-modal', $.proxy(this.modal.close, this));
    },
    disableEvents: function()
    {
      this.$modalClose.off('click.redactor-modal');
      $(document).off('keyup.redactor-modal');
      this.$editor.off('keyup.redactor-modal');
      this.$modalBox.off('click.redactor-modal');
      $(window).off('resize.redactor-modal');
    },
    closeHandler: function(e)
    {
      if (e.which != this.keyCode.ESC) return;

      this.modal.close(false);
    },
    close: function(e)
    {
      if (e)
      {
        if (!$(e.target).hasClass('redactor-modal-close-btn') && e.target != this.$modalClose[0] && e.target != this.$modalBox[0])
        {
          return;
        }

        e.preventDefault();
      }

      if (!this.$modalBox) return;

      this.modal.disableEvents();

      this.$modalOverlay.remove();

      this.$modalBox.fadeOut('fast', $.proxy(function()
      {
        this.$modalBox.remove();

        setTimeout($.proxy(this.utils.restoreScroll, this), 0);

        if (e !== undefined) this.selection.restore();

        $(document.body).css('overflow', this.modal.bodyOveflow);
        this.core.setCallback('modalClosed', this.modal.templateName);

      }, this));

    }
  };
}



$.Redactor.fn.image = function () {
  return {
    show: function()
    {

      this.modal.load('image', this.lang.get('image'), 700);
      this.selection.save();
      this.upload.init('#redactor-modal-image-droparea', this.opts.imageUpload, this.image.insert);

      var isMobile = this.utils.isMobile();

      if (isMobile) {
        $('#redactor-modal-image-droparea input[type="file"]').trigger('click');
      } else {
        this.modal.show();
      }

    },
    showEdit: function($image)
    {
      var $link = $image.closest('a');

      this.modal.load('imageEdit', this.lang.get('edit'), 705);

      var me = this,
          $modal = this.modal.getModal(),
          src = $image.data('origin') || $image[0].src
          .replace(/(jpg|jpeg|png|gif|bmp)$/, 'original.$1')
          .replace(/https?:\/\/[^\/]+/, '');

      src = decodeURI(src);

      var img = new Image();
          img.src = src;

      img.onload = function() {
        $(this).cropper({
          multiple: true,
          zoomable: false
        });
      }

      $modal.append(img);

      this.modal.createCancelButton();
      this.image.buttonDelete = this.modal.createDeleteButton(this.lang.get('_delete'));
      this.image.buttonSave = this.modal.createActionButton(this.lang.get('save'));

      this.image.buttonDelete.on('click', $.proxy(function()
      {
        this.image.remove($image);

      }, this));

      this.image.buttonSave.on('click', $.proxy(function()
      {
        var URL = this.$element.data('crop-url'),
            imageDataURL = $(img).cropper('getDataURL', true),
            data = $(img).cropper('getData', true),
            data = JSON.stringify({Url: src.replace(/\.original\.(jpg|jpeg|png|gif|bmp)$/, '.$1'), CropOption: data, Crop: true});

        $.ajax({
          type: 'POST',
          contentType: 'application/json; charset=UTF-8',
          dataType: 'json',
          cache: false,
          timeout: 7777,
          url: URL,
          data: data,
          beforeSends: function() {
            // sending cropped image; disable submit.
            $(me.$element[0].form).find('input[type="submit"]').attr('disabled', true);
          }
        }).done(function(data) {
          $(me.$element[0].form).find('input[type="submit"]').removeAttr('disabled');
          $image.data('origin', src)[0].src = data.url;
          me.image.update($image);
        });

        $image.data('origin', src)[0].src = imageDataURL;

        this.image.update($image);

      }, this));

      $('#redactor-image-title').val($image.attr('alt'));

      if (!this.opts.imageLink) $('.redactor-image-link-option').hide();
      else
      {
        var $redactorImageLink = $('#redactor-image-link');

        $redactorImageLink.attr('href', $image.attr('src'));
        if ($link.length !== 0)
        {
          $redactorImageLink.val($link.attr('href'));
          if ($link.attr('target') == '_blank') $('#redactor-image-link-blank').prop('checked', true);
        }
      }

      if (!this.opts.imagePosition) $('.redactor-image-position-option').hide();
      else
      {
        var floatValue = ($image.css('display') == 'block' && $image.css('float') == 'none') ? 'center' : $image.css('float');
        $('#redactor-image-align').val(floatValue);
      }

      this.modal.show();

    },
    setFloating: function($image)
    {
      var floating = $('#redactor-image-align').val();

      var imageFloat = '';
      var imageDisplay = '';
      var imageMargin = '';

      switch (floating)
      {
        case 'left':
          imageFloat = 'left';
          imageMargin = '0 ' + this.opts.imageFloatMargin + ' ' + this.opts.imageFloatMargin + ' 0';
        break;
        case 'right':
          imageFloat = 'right';
          imageMargin = '0 0 ' + this.opts.imageFloatMargin + ' ' + this.opts.imageFloatMargin;
        break;
        case 'center':
          imageDisplay = 'block';
          imageMargin = 'auto';
        break;
      }

      $image.css({ 'float': imageFloat, display: imageDisplay, margin: imageMargin });
      $image.attr('rel', $image.attr('style'));
    },
    update: function($image)
    {
      this.image.hideResize();
      this.buffer.set();

      var $link = $image.closest('a');

      $image.attr('alt', $('#redactor-image-title').val());

      this.image.setFloating($image);

      // as link
      var link = $.trim($('#redactor-image-link').val());
      if (link !== '')
      {
        // test url (add protocol)
        var pattern = '((xn--)?[a-z0-9]+(-[a-z0-9]+)*\\.)+[a-z]{2,}';
        var re = new RegExp('^(http|ftp|https)://' + pattern, 'i');
        var re2 = new RegExp('^' + pattern, 'i');

        if (link.search(re) == -1 && link.search(re2) === 0 && this.opts.linkProtocol)
        {
          link = this.opts.linkProtocol + '://' + link;
        }

        var target = ($('#redactor-image-link-blank').prop('checked')) ? true : false;

        if ($link.length === 0)
        {
          var a = $('<a href="' + link + '">' + this.utils.getOuterHtml($image) + '</a>');
          if (target) a.attr('target', '_blank');

          $image.replaceWith(a);
        }
        else
        {
          $link.attr('href', link);
          if (target)
          {
            $link.attr('target', '_blank');
          }
          else
          {
            $link.removeAttr('target');
          }
        }
      }
      else if ($link.length !== 0)
      {
        $link.replaceWith(this.utils.getOuterHtml($image));

      }

      this.modal.close();
      this.observe.images();
      this.code.sync();

    },
    setEditable: function($image)
    {
      if (this.opts.imageEditable)
      {
        $image.on('dragstart', $.proxy(this.image.onDrag, this));
      }

      $image.on('mousedown', $.proxy(this.image.hideResize, this));
      $image.on('click.redactor touchstart', $.proxy(function(e)
      {
        this.observe.image = $image;

        if (this.$editor.find('#redactor-image-box').length !== 0) return false;

        this.image.resizer = this.image.loadEditableControls($image);

        $(document).on('click.redactor-image-resize-hide.' + this.uuid, $.proxy(this.image.hideResize, this));
        this.$editor.on('click.redactor-image-resize-hide.' + this.uuid, $.proxy(this.image.hideResize, this));

        // resize
        if (!this.opts.imageResizable) return;

        this.image.resizer.on('mousedown.redactor touchstart.redactor', $.proxy(function(e)
        {
          this.image.setResizable(e, $image);
        }, this));


      }, this));
    },
    setResizable: function(e, $image)
    {
      e.preventDefault();

        this.image.resizeHandle = {
            x : e.pageX,
            y : e.pageY,
            el : $image,
            ratio: $image.width() / $image.height(),
            h: $image.height()
        };

        e = e.originalEvent || e;

        if (e.targetTouches)
        {
             this.image.resizeHandle.x = e.targetTouches[0].pageX;
             this.image.resizeHandle.y = e.targetTouches[0].pageY;
        }

      this.image.startResize();


    },
    startResize: function()
    {
      $(document).on('mousemove.redactor-image-resize touchmove.redactor-image-resize', $.proxy(this.image.moveResize, this));
      $(document).on('mouseup.redactor-image-resize touchend.redactor-image-resize', $.proxy(this.image.stopResize, this));
    },
    moveResize: function(e)
    {
      e.preventDefault();

      e = e.originalEvent || e;

      var height = this.image.resizeHandle.h;

            if (e.targetTouches) height += (e.targetTouches[0].pageY -  this.image.resizeHandle.y);
            else height += (e.pageY -  this.image.resizeHandle.y);

      var width = Math.round(height * this.image.resizeHandle.ratio);

      if (height < 50 || width < 100) return;

            this.image.resizeHandle.el.width(width);
            this.image.resizeHandle.el.height(this.image.resizeHandle.el.width()/this.image.resizeHandle.ratio);

            this.code.sync();
    },
    stopResize: function()
    {
      this.handle = false;
      $(document).off('.redactor-image-resize');

      this.image.hideResize();
    },
    onDrag: function(e)
    {
      if (this.$editor.find('#redactor-image-box').length !== 0)
      {
        e.preventDefault();
        return false;
      }

      this.$editor.on('drop.redactor-image-inside-drop', $.proxy(function()
      {
        setTimeout($.proxy(this.image.onDrop, this), 1);

      }, this));
    },
    onDrop: function()
    {
      this.image.fixImageSourceAfterDrop();
      this.observe.images();
      this.$editor.off('drop.redactor-image-inside-drop');
      this.clean.clearUnverified();
      this.code.sync();
    },
    fixImageSourceAfterDrop: function()
    {
      this.$editor.find('img[data-save-url]').each(function()
      {
        var $el = $(this);
        $el.attr('src', $el.attr('data-save-url'));
        $el.removeAttr('data-save-url');
      });
    },
    hideResize: function(e)
    {
      if (e && $(e.target).closest('#redactor-image-box').length !== 0) return;
      if (e && e.target.tagName == 'IMG')
      {
        var $image = $(e.target);
        $image.attr('data-save-url', $image.attr('src'));
      }

      var imageBox = this.$editor.find('#redactor-image-box');
      if (imageBox.length === 0) return;

      if (this.opts.imageEditable)
      {
        this.image.editter.remove();
      }

      $(this.image.resizer).remove();

      imageBox.find('img').css({
        marginTop: imageBox[0].style.marginTop,
        marginBottom: imageBox[0].style.marginBottom,
        marginLeft: imageBox[0].style.marginLeft,
        marginRight: imageBox[0].style.marginRight
      });

      imageBox.css('margin', '');
      imageBox.find('img').css('opacity', '');
      imageBox.replaceWith(function()
      {
        return $(this).contents();
      });

      $(document).off('click.redactor-image-resize-hide.' + this.uuid);
      this.$editor.off('click.redactor-image-resize-hide.' + this.uuid);

      if (typeof this.image.resizeHandle !== 'undefined')
      {
        this.image.resizeHandle.el.attr('rel', this.image.resizeHandle.el.attr('style'));
      }

      this.code.sync();

    },
    loadResizableControls: function($image, imageBox)
    {
      if (this.opts.imageResizable && !this.utils.isMobile())
      {
        var imageResizer = $('<span id="redactor-image-resizer" data-redactor="verified"></span>');

        if (!this.utils.isDesktop())
        {
          imageResizer.css({ width: '15px', height: '15px' });
        }

        imageResizer.attr('contenteditable', false);
        imageBox.append(imageResizer);
        imageBox.append($image);

        return imageResizer;
      }
      else
      {
        imageBox.append($image);
        return false;
      }
    },
    loadEditableControls: function($image)
    {
      var imageBox = $('<span id="redactor-image-box" data-redactor="verified">');
      imageBox.css('float', $image.css('float')).attr('contenteditable', false);

      if ($image[0].style.margin != 'auto')
      {
        imageBox.css({
          marginTop: $image[0].style.marginTop,
          marginBottom: $image[0].style.marginBottom,
          marginLeft: $image[0].style.marginLeft,
          marginRight: $image[0].style.marginRight
        });

        $image.css('margin', '');
      }
      else
      {
        imageBox.css({ 'display': 'block', 'margin': 'auto' });
      }

      $image.css('opacity', '.5').after(imageBox);


      if (this.opts.imageEditable)
      {
        // editter
        this.image.editter = $('<span id="redactor-image-editter" data-redactor="verified">' + this.lang.get('edit') + '</span>');
        this.image.editter.attr('contenteditable', false);
        this.image.editter.on('click', $.proxy(function()
        {
          this.image.showEdit($image);
        }, this));

        imageBox.append(this.image.editter);

        // position correction
        var editerWidth = this.image.editter.innerWidth();
        this.image.editter.css('margin-left', '-' + editerWidth/2 + 'px');
      }

      return this.image.loadResizableControls($image, imageBox);

    },
    remove: function(image)
    {
      var $image = $(image);
      var $link = $image.closest('a');
      var $figure = $image.closest('figure');
      var $parent = $image.parent();
      if ($('#redactor-image-box').length !== 0)
      {
        $parent = $('#redactor-image-box').parent();
      }

      var $next;
      if ($figure.length !== 0)
      {
        $next = $figure.next();
        $figure.remove();
      }
      else if ($link.length !== 0)
      {
        $parent = $link.parent();
        $link.remove();
      }
      else
      {
        $image.remove();
      }

      $('#redactor-image-box').remove();

      if ($figure.length !== 0)
      {
        this.caret.setStart($next);
      }
      else
      {
        this.caret.setStart($parent);
      }

      // delete callback
      this.core.setCallback('imageDelete', $image[0].src, $image);

      this.modal.close();
      this.code.sync();
    },
    insert: function(json, direct, e)
    {
      // error callback
      if (typeof json.error != 'undefined')
      {
        this.modal.close();
        this.selection.restore();
        this.core.setCallback('imageUploadError', json);
        return;
      }

      var $img;
      if (typeof json == 'string')
      {
        $img = $(json).attr('data-redactor-inserted-image', 'true');
      }
      else
      {
        $img = $('<img>');
        $img.attr('src', json.filelink).attr('data-redactor-inserted-image', 'true');
      }


      var node = $img;
      var isP = this.utils.isCurrentOrParent('P');
      if (isP)
      {
        // will replace
        node = $('<blockquote />').append($img);
      }

      if (direct)
      {
        this.selection.removeMarkers();
        var marker = this.selection.getMarker();
        this.insert.nodeToCaretPositionFromPoint(e, marker);
      }
      else
      {
        this.modal.close();
      }

      this.selection.restore();
      this.buffer.set();

      this.insert.html(this.utils.getOuterHtml(node), false);

      var $image = this.$editor.find('img[data-redactor-inserted-image=true]').removeAttr('data-redactor-inserted-image');

      if (isP)
      {
        $image.parent().contents().unwrap().wrap('<p />');
      }
      else if (this.opts.linebreaks)
      {
        $image.before('<br>').after('<br>');
      }

      if (typeof json == 'string') return;

      this.core.setCallback('imageUpload', $image, json);

    }
  };
}
