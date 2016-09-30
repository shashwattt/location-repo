var data = {};

if (navigator.geolocation) {

      navigator.geolocation.getCurrentPosition(displayPosition, displayError,{enableHighAccuracy: true, timeout: 1000, maximumAge: 0 });

}else{
    console.log("Your browser does not support HTML5, shame on you !!");
    getFromIpify();
}

function doLater(response){
    alert(response);
}
function callLocationService(){
    $.post(/api/,JSON.stringify(data),doLater)
}


function displayPosition(position) {
  console.log("Latitude: " + position.coords.latitude + ", Longitude: " + position.coords.longitude);
  data.Longitude = position.coords.longitude;
  data.Latitude = position.coords.latitude;
  callLocationService();
  //$.get(/api/,doLater);

}



function displayError(error) {
  var errors = { 
    1: 'Permission denied',
    2: 'Position unavailable',
    3: 'Request timeout'
  };
  console.log("Error: " + errors[error.code]);
  getFromIpify();

  //$.get(/api/,doLater);
}

function getFromIpify(){
  $.get('http://api.ipify.org/?format=json&callback', function(response){
    console.log(response);
    data.Ip = response.ip;
    callLocationService();
  })
}
