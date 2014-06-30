var poddclubApp = angular.module('poddclubApp', []);

poddclubApp.controller('CategoryListCtrl', function ($scope) {

  $scope.currentCategory = "Tech";

  $scope.categories = [
    {'name': 'Tech',
     'snippet': 'Tech talks.',
     'podcastCount': 13,
     'podcasts':[{'name':'TWIT','episodeNumber':432,'author':'Leo Laport','length':'900s','url':'http://www.youtube.com/blahblah'},{'name':'Tech Talk','episodeNumber':432,'author':'Leo Laport','length':'900s','url':'http://www.youtube.com/blahblah'}] },
    {'name': 'TED Talks',
     'snippet': 'Ideas worth spreading.',
     'podcastCount': 45,
     'podcasts':[{'name':'TED Talks 1','episodeNumber':432,'author':'Leo Laport','length':'900s','url':'http://www.youtube.com/blahblah'},{'name':'TED Talks 2','episodeNumber':432,'author':'Leo Laport','length':'900s','url':'http://www.youtube.com/blahblah'}] },
    {'name': 'Astrophysics',
     'snippet': 'Podcasts from the stars.',
     'podcastCount': 6,
     'podcasts':[{'name':'Astro1','episodeNumber':42,'author':'Neil deGrasse Tyson','length':'800s','url':'http://www.youtube.com/blahblah'},{'name':'Astro2','episodeNumber':34,'author':'Neil deGrasse Tyson','length':'700s','url':'http://www.youtube.com/blahblah'}] },
  ];
  
  $scope.addCategory = function(){
	  $scope.categories.push({'name':$scope.newCategory,'snippet':"hi","podcastCount":0})
	  $scope.newCategory = ''
  }

  $scope.addPodcast = function(){
    for (category in $scope.categories){
      if ($scope.categories[category]['name'] === $scope.currentCategory){
        $scope.categories[category]['podcasts'].push({'name':$scope.newPodcastName,'episodeNumber':0,'author':$scope.newPodcastAuthor,'length':'900s','url':$scope.newPodcastURL})
        $scope.newPodcastName = ''
        $scope.newPodcastAuthor = ''
        $scope.newPodcastURL = ''
      }
    }
  }

  // $scope.addPodcast = function(){
  //   for (var category in $scope.categories){
  //     var podcasts = category['podcasts']
  //     podcasts.push({'name':'new','episodeNumber':333,'author':'whoKnows','lenght':'90s','url':$scope.newPodcastURL}
  //     $scope.newPodcastUrl = ''
  //   }
  // }

  $scope.changeCurrentCategory = function(name){
    $scope.currentCategory = name
  }

  $scope.removeCategory = function(){}
  
});
