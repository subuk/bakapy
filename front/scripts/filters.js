'use strict';

/* Filters */

var bakapyFilters = angular.module('bakapyFilters', []);

bakapyFilters.filter('bytes', function() {
  return function(bytes, precision) {
    if (isNaN(parseFloat(bytes)) || !isFinite(bytes)) {
      return '-';
    }

    if (bytes === 0) {
      return 0;
    }

    if (typeof precision === 'undefined') {
      precision = 1;
    }

    var units = ['bytes', 'kB', 'MB', 'GB', 'TB', 'PB'],
        number = Math.floor(Math.log(bytes) / Math.log(1024));

    return (bytes / Math.pow(1024, Math.floor(number))).toFixed(precision) + ' ' + units[number];
  };
});

bakapyFilters.filter('keyLength', function() {
  return function(input) {
    if (!angular.isObject(input)) {
      throw Error('Usage of non-objects with keylength filter!');
    }

    return Object.keys(input).length;
  }
});
