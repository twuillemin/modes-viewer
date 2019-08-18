var viewermap = L.map('viewermap').setView([34.670619, 33.029099], 13);

L.tileLayer('https://api.tiles.mapbox.com/v4/{id}/{z}/{x}/{y}.png?access_token=pk.eyJ1IjoibWFwYm94IiwiYSI6ImNpejY4NXVycTA2emYycXBndHRqcmZ3N3gifQ.rJcFIG214AriISLbB6B5aw', {
    maxZoom: 18,
    attribution: 'Map data &copy; <a href="https://www.openstreetmap.org/">OpenStreetMap</a> contributors, ' +
        '<a href="https://creativecommons.org/licenses/by-sa/2.0/">CC-BY-SA</a>, ' +
        'Imagery © <a href="https://www.mapbox.com/">Mapbox</a>',
    id: 'mapbox.streets'
}).addTo(viewermap);

L.marker([34.670619, 33.029099]).addTo(viewermap);

console.log("View started");

let planes = {};

let planeIcon = L.icon({
    iconUrl: '/img/plane.png',
    shadowUrl: '/img/shadow.png',
    iconSize: [32, 32], // size of the icon
    shadowSize: [32, 32], // size of the shadow
    iconAnchor: [0, 0], // point of the icon which will correspond to marker's location
    shadowAnchor: [-2, -2],  // the same for the shadow
    popupAnchor: [0, -64] // point from which the popup should open relative to the iconAnchor
});

function updatePlane(messagePlane) {
    const planeAddress = messagePlane["address"];

    let planeObj = planes[planeAddress];

    if (planes[planeAddress] === undefined) {
        planeObj = {
            marker: null,
            data:messagePlane
        };
        planes[planeAddress] = planeObj
    }
    else{
        planeObj.data = messagePlane
    }

    if (messagePlane.latitude != null && messagePlane.longitude != null){
        if (planeObj.marker == null){
            const marker = L.marker([messagePlane.latitude, messagePlane.longitude], {icon: planeIcon});
            marker.addTo(viewermap).bindPopup("addr: " + planeAddress);
            planeObj.marker = marker
        }
        else {
            const newLatLng = new L.LatLng(messagePlane.latitude, messagePlane.longitude);
            planeObj.marker.setLatLng(newLatLng);
        }
    }
}

const somePackage = {};
somePackage.connect = function () {
    const ws = new WebSocket('ws://127.0.0.1:8081/events');
    ws.onopen = function () {
        console.log('ws connected');
        somePackage.ws = ws;
    };
    ws.onerror = function () {
        console.log('ws error');
    };
    ws.onclose = function () {
        console.log('ws closed');
    };
    ws.onmessage = function (msg) {
        console.log('rawmessage :', msg.data);
        var plane = JSON.parse(msg.data);
        console.log('in :', plane);
        // message received, do something
        updatePlane(plane)
    };
};

somePackage.connect();