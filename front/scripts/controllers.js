'use strict';

/* Controllers */

var bakapyControllers = angular.module('bakapyControllers', []);

bakapyControllers.controller('BackupListCtrl', ['$scope', '$http',
  function($scope, $http) {
    $http.get('/metadata').success(function(data) {
      $scope.backups = [];

      var links = jQuery(data).find('a'),
          expr = /^\w{8}-\w{4}-\w{4}-\w{4}-\w{12}$/,
          href,
          i,
          j;

      for (i = 0, j = links.length; i < j; i++) {
        href = links.eq(i).attr('href');
        if (expr.test(href)) {
          $http.get('/metadata/'+ href, {'responseType': 'json'}).success(function(item) {
            $scope.backups.push(item);
          });
        }
      }
    });
  }]);

bakapyControllers.controller('BackupDetailCtrl', ['$scope', '$http', '$routeParams', 'base64', 'CONFIG', '$location',
  function($scope, $http, $routeParams, base64, CONFIG, $location) {
    $http.get('/metadata/'+ $routeParams.id, {'responseType': 'json'}).success(function(data) {
      var Duration = 0,
          AvgSpeed = 0,
          fileList = [],
          i,
          j;

      if (data.Output) {
        data.Output = base64.decode(data.Output);
      }

      if (data.Errput) {
        data.Errput = base64.decode(data.Errput);
      }

      if (data.Files) {
        for (i = 0, j = data.Files.length; i < j; i++) {
          fileList.push(encodeURI(CONFIG.STORAGE_URL + '/' + data.Namespace + '/' + data.Files[i].Name));
        }
        data.Files = fileList;
      }

      if (data.StartTime && data.EndTime) {
        Duration = (new Date(data.EndTime).getTime() - new Date(data.StartTime).getTime()) / 1000;
        data.Duration = Duration;
      }

      if (data.Duration && data.TotalSize) {
        AvgSpeed = data.TotalSize / data.Duration;
        data.AvgSpeed = AvgSpeed;
      }

      console.log('data.Duration ', data.Duration)
      console.log('data.AvgSpeed ', data.AvgSpeed);

      $scope.backup = data;
    });
  }]);
