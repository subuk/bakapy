!function(namespace) {
  'use strict';

  function Modal(elem, params) {
    this.$element = jQuery(elem);
    this.params = params || {};
    this.cssReadyElement = this.params.cssReadyElement || 'JS-Modal-ready';
    this.cssActiveElement = this.params.cssActiveElement || 'JS-Modal-active';

    this.__construct();
  }

  Modal.prototype.__construct = function __construct() {
    this.$box = this.$element.find('.JS-Modal-Box');
    this.$close = this.$element.find('.JS-Modal-Close');
    this.$title = this.$element.find('.JS-Modal-Title');
    this.$container = this.$element.find('.JS-Modal-Container');

    this._init();
  };

  Modal.prototype._init = function _init() {
    var _this = this;

    this.$close.on('click.JS-Modal', function() { _this._close.apply(_this, []); });

    $('body').on("keyup", function(e) {
      if ((e.keyCode == 27)) {
        _this._close.apply(_this, []);
      }
    });

    $('.JS-Gannt-Modal').click(function() {
      if (_this.$element.hasClass('JS-Modal-active'))
        _this._close.apply(_this, []);
    });

    $('.JS-Modal-Box').click(function(event){
      event.stopPropagation();
    });

    /* API. Events */
    this.$element.on('modal:setContent', function(e, data) { _this.setContent.apply(_this, [data]); });
    this.$element.on('modal:open', function() { _this.open.apply(_this, []); });
    this.$element.on('modal:close', function() { _this.close.apply(_this, []); });
    this.$element.on('modal:clear', function() { _this.clear.apply(_this, []); });

    this._ready();
  } ;

  Modal.prototype._ready = function _ready() {
    this.$element
      .addClass(this.cssReadyElement)
      .addClass('JS-Modal-ready');
  };

  Modal.prototype._setContent = function _setContent(content) {
    this.$container.html(content);
  };

  Modal.prototype._open = function _open() {
    if (!this.$element.hasClass('JS-Modal-active')) {
      this.$element
        .addClass(this.cssActiveElement)
        .addClass('JS-Modal-active')
    }
  };

  Modal.prototype._close = function _close() {
    if (this.$element.hasClass('JS-Modal-active')) {
      this.$element
        .removeClass(this.cssActiveElement)
        .removeClass('JS-Modal-active');
    }
  };

  Modal.prototype._clear = function _clear() {
  };

  /* API. Methods */
  Modal.prototype.setContent = function setContent(content) {
    if (!arguments.length) {
      return false;
    }

    this._setContent(content);
  };

  Modal.prototype.open = function open() {
    this._open();
  };

  Modal.prototype.close = function close() {
    this._close();
  };

  Modal.prototype.clear = function clear() {
    this._clear();
  };

  namespace.Modal = Modal;
}(this);
