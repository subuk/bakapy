'use strict';

/* Directives */

var bakapyDirectives = angular.module('bakapyDirectives', []);

bakapyDirectives.directive('loading', ['$http', function($http) {
  return {
    restrict: 'A',
    link: function(scope, elem, attrs)
    {
      scope.isLoading = function() {
        return $http.pendingRequests.length > 0;
      };

      scope.$watch(scope.isLoading, function (value)
      {
        if (value) {
          elem.show();
        } else {
          elem.hide();
        }
      });
    }
  };

}]);
