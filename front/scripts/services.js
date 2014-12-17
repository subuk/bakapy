'use strict';

/* Services */

var bakapyServices = angular.module('bakapyServices', []);

bakapyServices.factory('Backups', ['$http', function($http) {
  var backups = {};

  $http.get(CONFIG.METADATA_URL).success(function(data) {
    var links = jQuery(data).find('a'),
        href,
        i,
        j;

    for (i = 0, j = links.length; i < j; i++) {
      href = links.eq(i).attr('href');
      $http.get(CONFIG.METADATA_URL + '/' + href, {'responseType': 'json'}).success(function(item) {
        if (item && item.JobName !== 'undefined') {
          if (typeof backups[item.JobName] === 'undefined') {
            backups[item.JobName] = [];
          }

          item._source = href;
          backups[item.JobName].push(item);
        }
      });
    }
  });

  return backups;
}]);
