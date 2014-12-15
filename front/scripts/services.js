'use strict';

/* Services */

var bakapyServices = angular.module('bakapyServices', []);

bakapyServices.factory('Backups', ['$http', function($http) {
  var backups = {};

  $http.get(CONFIG.METADATA_URL).success(function(data) {
    var links = jQuery(data).find('a'),
        expr = /^\w{8}-\w{4}-\w{4}-\w{4}-\w{12}$/,
        href,
        i,
        j;

    for (i = 0, j = links.length; i < j; i++) {
      href = links.eq(i).attr('href');
      if (expr.test(href)) {
        $http.get(CONFIG.METADATA_URL + '/' + href, {'responseType': 'json'}).success(function(item) {
          if (typeof backups[item.JobName] === 'undefined') {
            backups[item.JobName] = [];
          }

          backups[item.JobName].push(item);
        });
      }
    }
  });

  return backups;
}]);
