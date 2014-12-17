'use strict';

/* Controllers */

var bakapyControllers = angular.module('bakapyControllers', []);

bakapyControllers.controller('BackupListCtrl', ['$scope', '$http', '$location', '$q', 'Backups',
  function($scope, $http, $location, $q, Backups) {
    $scope.sortBySuccess = '';
    $scope.query = typeof $location.search().q !== 'undefined' ? $location.search().q : '';
    $scope.backups = Backups;

    $scope.changeUriQuery = function changeUriQuery(value) {
      if (value === '') {
        return $location.url($location.path());
      }

      return $location.search('q', value);
    }
  }]);

bakapyControllers.controller('BackupDetailCtrl', ['$scope', '$http', '$routeParams', 'base64', 'CONFIG', '$location',
  function($scope, $http, $routeParams, base64, CONFIG, $location) {
    $http.get(CONFIG.METADATA_URL + '/' + $routeParams.id, {'responseType': 'json'}).success(function(data) {
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
          fileList.push({
            'source': (encodeURI(CONFIG.STORAGE_URL + '/' + data.Namespace + '/' + data.Files[i].Name)),
            'size': data.Files[i].Size
          });
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

      $scope.backup = data;
    })
    .error(function(data) {
      // $location.path('/404');
    });
  }]);
