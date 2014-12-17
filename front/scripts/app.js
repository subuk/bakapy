'use strict';

/* App Module */

var bakapyApp = angular.module('bakapyApp', [
  'ngRoute',
  'ab-base64',
  'bakapyControllers',
  'bakapyFilters',
  'bakapyDirectives',
  'bakapyServices'
]);


/* CONSTANT */
bakapyApp.constant( 'CONFIG', CONFIG);


/* CONFIG */
bakapyApp.config(['$routeProvider',
  function($routeProvider, $locationProvider) {
    $routeProvider
      .when('/', {
        templateUrl: '/partials/backup-list.html',
        controller: 'BackupListCtrl',
        reloadOnSearch: false
      })
      .when('/404', {
        templateUrl: '/404.html'
      })
      .when('/:id', {
        templateUrl: '/partials/backup-details.html',
        controller: 'BackupDetailCtrl'
      })
      .otherwise({
        redirectTo: '/404'
      });
  }
]);


/* FILTERS */
bakapyApp.filter('bytes', function() {
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
