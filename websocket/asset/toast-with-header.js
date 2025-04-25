'use strict';
var Toast = function (element, config) {
  var
    _this = this,
    _element = element,
    _config = {
      autohide: true,
      delay: 5000
    };
  for (var prop in config) {
    _config[prop] = config[prop];
  }
  Object.defineProperty(this, 'element', {
    get: function () {
      return _element;
    }
  });
  Object.defineProperty(this, 'config', {
    get: function () {
      return _config;
    }
  });
  // setListeners
  $(_element).on('click', '.toast__close', function () {
    _this.hide();
  })
};

Toast.prototype = {
  show: function () {
    var _this = this;
    $(this.element).addClass('toast_show');
    if (this.config.autohide) {
      setTimeout(function () {
        _this.hide();
      }, this.config.delay)
    }
  },
  hide: function () {
    $(this.element).removeClass('toast_show');
    $(this.element).trigger('hidden.toast');
  }
};

Toast.create = function (header, body, color) {
  return $('<div>', { 'class': 'toast', style: 'background-color: rgba(' + parseInt(color.substr(1, 2), 16) + ',' + parseInt(color.substr(3, 2), 16) + ',' + parseInt(color.substr(5, 2), 16) + ',0.5)' })
    .append(
      ($('<div>', { 'class': 'toast__header' }).text(header))
        .append(($('<button>', { type: 'button', 'class': 'toast__close' })).html('&times;'))
    )
    .append($('<div>', { 'class': 'toast__body' }).text(body));
};

Toast.add = function (params) {
  var config = {
    header: 'Название заголовка',
    body: 'Текст сообщения...',
    color: '#ffffff',
    autohide: true,
    delay: 5000
  };
  if (params !== undefined) {
    for (var item in params) {
      config[item] = params[item];
    }
  }
  if (!$('.toasts').length) {
    $('body').append($('<div>', { 'class': 'toasts', style: 'position: fixed; top: 15px; right: 15px; width: 250px;' }));
  }
  $('.toasts').append(Toast.create(config.header, config.body, config.color));
  var toasts = $('.toast');
  var toast = new Toast(toasts[toasts.length - 1], { autohide: config.autohide, delay: config.delay });
  toast.show();
  return toast;
};