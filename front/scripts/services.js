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

      (function(_href) {
        $http.get(CONFIG.METADATA_URL + '/' + _href, {'responseType': 'json'}).success(function(item, status, headers, config) {
          if (item && item.JobName !== 'undefined') {
            if (typeof backups[item.JobName] === 'undefined') {
              backups[item.JobName] = [];
            }

            item._source = _href;
            backups[item.JobName].push(item);
          }
        });
      })(href);

    }
  });

  return backups;
}]);
