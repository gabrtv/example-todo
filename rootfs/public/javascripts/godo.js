(function() {
  var app = angular.module("GoDo", ["ngRoute", "ngResource"]);

  app.factory("Item", ["$resource", function($resource) {
    return $resource("/items/:id", {id: "@id"});
  }]);

  app.config(['$routeProvider', function($routeProvider) {
    $routeProvider.when('/', {
      templateUrl: '/html/index.html',
      controller: 'ItemsCtrl',
      resolve: {
        items: ["Item", function(Item) {
          return Item.query().$promise;
        }]
      }
    });
  }]);

  app.controller('ItemsCtrl', ['$scope', 'items', 'Item', function($scope, items, Item) {
    $scope.items = items;

    $scope.addItem = function() {
      Item.save($scope.newItem).$promise.then(function(item) {
        $scope.items.push(item)
        $scope.newItem = {}
      });
    };

    $scope.removeItem = function(index) {
      $scope.items[index].$delete().then(function() {
        $scope.items.splice(index, 1);
      });
    };
  }]);

})();
