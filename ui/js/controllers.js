var poddclubApp = angular.module('poddclubApp', []);

poddclubApp.controller('CategoryListCtrl', function ($scope, $http, $sce) {

  $scope.categories = [
    {'name': 'Tech',
     'podcasts':[{'name':'TWIT','episodeNumber':432,'author':'Leo Laport','length':'900s','url':'http://www.youtube.com/blahblah'},{'name':'Tech Talk','episodeNumber':432,'author':'Leo Laport','length':'900s','url':'http://www.youtube.com/blahblah'}] },
    {'name': 'TED Talks',
     'podcasts':[{'name':'TED Talks 1','episodeNumber':432,'author':'Leo Laport','length':'900s','url':'http://www.youtube.com/blahblah'},{'name':'TED Talks 2','episodeNumber':432,'author':'Leo Laport','length':'900s','url':'http://www.youtube.com/blahblah'}] },
    {'name': 'Astrophysics',
     'podcasts':[{'name':'Astro1','episodeNumber':42,'author':'Neil deGrasse Tyson','length':'800s','url':'http://www.youtube.com/blahblah'},{'name':'Astro2','episodeNumber':34,'author':'Neil deGrasse Tyson','length':'700s','url':'http://www.youtube.com/blahblah'}] },
  ];

  $scope.loadCategories = function(userId) {
      $http.get('/api/users/1/categories').
          success(function(data) {
            $scope.categories = data;
            if ($scope.categories.length >= 1) {
                $scope.currentCategory = $scope.categories[0];
                $scope.loadPodcasts(userId, $scope.categories[0].id);
            }
            else {
                $scope.currentCategory = null;
            }
          });
  }

  $scope.loadPodcasts = function(userId, categoryId) {
      $http.get('/api/users/' + userId + '/categories/' + categoryId + '/podcasts').
          success(function(data) {
            $scope.podcasts = data;
          });
  }

  $scope.addCategory = function(){
    var apiUrl = "/api/users/" + $scope.userId + "/categories";
    var newCategory = $scope.newCategory;
    $http({
        method: 'POST',
        url: apiUrl,
        data: $.param({name: newCategory}),
        headers: {'Content-Type': 'application/x-www-form-urlencoded'}
    }).success(function() {
        $scope.newCategory = "";
        $scope.loadCategories($scope.userId);
    });
  }

  $scope.scheduleDownload = function(userId, categoryId, url) {
    var apiUrl = "/api/users/" + userId + "/categories/" + categoryId + "/schedule-download";

    $http({
        method: 'POST',
        url: apiUrl,
        data: $.param({url: url}),
        headers: {'Content-Type': 'application/x-www-form-urlencoded'}
    })

    $scope.newPodcastUrl = "";
  }

  $scope.deletePodcast = function(podcastId) {
    var apiUrl = "/api/podcasts/" + podcastId;
    $http.delete(apiUrl)
        .success(function() {
            $scope.loadPodcasts($scope.userId, $scope.currentCategory.id)
        })
  }

  $scope.setCurrentCategory = function(category){
    $scope.currentCategory = category;
    $scope.loadPodcasts($scope.userId, category.id);
  }

  $scope.rssLink = function(category) {
    return $sce.trustAsHtml('/rss/' + category.id);
  }

  $scope.formatDuration = function(duration) {
    function zeroPad(num, places) {
      var zero = places - num.toFixed(0).toString().length + 1;
      return Array(+(zero > 0 && zero)).join("0") + num;
    }
    var hours = Math.floor(duration / 3600);
    duration -= hours * 3600
    var minutes = Math.floor(duration / 60)
    duration -= minutes * 60
    var seconds = duration
    var s = "";
    if (hours > 0) {
        s += zeroPad(hours, 2) + ":"
    }
    s += zeroPad(minutes, 2) + ":"
    s += zeroPad(seconds, 2);
    return s;
  }

  $scope.formatSize = function(size) {
    return "" + (size / (1024 * 1024)).toFixed(0) + " MB";
  }

  $scope.removeCategory = function(){}

  $scope.userId = 1;
  $scope.loadCategories($scope.userId);
});
