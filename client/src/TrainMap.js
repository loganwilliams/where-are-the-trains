import React, { Component } from 'react';
import './TrainMap.css';
import ReactMapboxGl, { GeoJSONLayer } from "react-mapbox-gl";

const Map = ReactMapboxGl({
  accessToken: "pk.eyJ1IjoibG9nYW53IiwiYSI6IlQzWHJqc3cifQ.KY3j-syHXeYmI69JmLqGqQ"
});

class TrainMap extends Component {
  constructor() {
    super();

    this.state = {
      drawnTrains: false,
      currentTrains: false,
      oldPositionByTrainId: false,
      center: [-74.006, 40.7128]
    };
  }

  componentWillMount() {
    this.updateJson();
    this.interpolate();
  }

  onStyleLoad(map, e) {
    this.setState( {map: map} );
  }

  updateJson() {
    fetch('http://localhost:8080/live')
    .then(result => result.json())
    .then((resultJson) => {
      if (this.state.drawnTrains) {
        var oldPositionByTrainId = this.state.newPositionByTrainId;
        var newPositionByTrainId = {};

        for (var i = 0; i < resultJson.features.length; i++) {
            newPositionByTrainId[resultJson.features[i].properties.id] = resultJson.features[i].geometry.coordinates.slice();
        }

        this.setState({
          drawnTrains: this.state.currentTrains,
          oldPositionByTrainId: oldPositionByTrainId,
          newPositionByTrainId: newPositionByTrainId,
          currentTrains: resultJson
        });

        this.interpolant = 0.005;
      } else {
        // this is the first time that we have recieved any train data
        var newPositionByTrainId = {};

        for (var i = 0; i < resultJson.features.length; i++) {
            newPositionByTrainId[resultJson.features[i].properties.id] = resultJson.features[i].geometry.coordinates.slice();
        }

        this.setState({
          drawnTrains: resultJson,
          currentTrains: resultJson,
          newPositionByTrainId: newPositionByTrainId
        });
      }
    });

    // update the train positions every 10s
    window.setTimeout(this.updateJson.bind(this), 10000);
  }

  componentWillUpdate(nextProps, nextState) {
    const map = nextState.map;
    const trains = nextState.drawnTrains;

    // this is necessary in order for the trains to update position.
    if (map) {
      map.getSource('trains').setData(trains);
    }
  }

  // interpolate each trains position from its old coordinates to its new ones, so that they move smoothly.
  interpolate() {
    if (this.state.oldPositionByTrainId) {

      var interpolatedTrains = this.state.currentTrains;

      for (var i = 0; i < this.state.currentTrains.features.length; i++) {
        if (this.state.currentTrains.features[i].properties.id in this.state.oldPositionByTrainId) {
          if (interpolatedTrains.features[i].geometry.coordinates[0] != this.state.oldPositionByTrainId[this.state.currentTrains.features[i].properties.id][0]) {
            interpolatedTrains.features[i].geometry.coordinates[0] = this.state.newPositionByTrainId[this.state.currentTrains.features[i].properties.id][0] * this.interpolant + this.state.oldPositionByTrainId[this.state.currentTrains.features[i].properties.id][0] * (1.0 - this.interpolant)
            interpolatedTrains.features[i].geometry.coordinates[1] = this.state.newPositionByTrainId[this.state.currentTrains.features[i].properties.id][1] * this.interpolant + this.state.oldPositionByTrainId[this.state.currentTrains.features[i].properties.id][1] * (1.0 - this.interpolant)
          }
        }
      }

      this.setState({
        drawnTrains: interpolatedTrains
      });

      this.interpolant += 0.01;
      if (this.interpolant > 1.0) {
        this.interpolant = 1.0;
      }
    }

    requestAnimationFrame(this.interpolate.bind(this));
  }

  render() {
    var geojson = [];
    
    var lineColors = {
      'circle-stroke-width': [
        'match', ['get', 'direction'],
        'N', 2.0,
        'S', 1.5,
        2.0],
      'circle-stroke-color': [
        'match', ['get', 'direction'],
        'N', 'black',
        'S', 'white',
        'gray'],
      'circle-color': [
        'match', ['get', 'line'],
        'A', '#2850ad',
        'C', '#2850ad',
        'E', '#2850ad',
        'B', '#ff6319',
        'D', '#ff6319',
        'F', '#ff6319',
        'M', '#ff6319',
        'G', '#6cbe45',
        'L', '#a7a9ac',
        'J', '#996633',
        'Z', '#2850ad',
        'N', '#fccc0a',
        'Q', '#fccc0a',
        'R', '#fccc0a',
        'W', '#fccc0a',
        '1', '#ee352e',
        '2', '#ee352e',
        '3', '#ee352e',
        '4', '#00933c',
        '5', '#00933c',
        '6', '#00933c',
        '6X', '#00933c',
        '7', '#b933ad',
        '7X', '#b933ad',
         '#808183' ]};

    if (this.state.drawnTrains) {
      geojson = <GeoJSONLayer
          id="trains"
          data={this.state.drawnTrains}
          circleLayout={{ visibility: 'visible'}}
          circlePaint={lineColors}
        />
    }

    return (
      <Map
        style="mapbox://styles/loganw/cje8694kmg25o2sqsy9ji49cw"
        containerStyle={{
          height: "100vh",
          width: "100vw"
        }}
        center={this.state.center}
        onStyleLoad={this.onStyleLoad.bind(this)}>
        {geojson}
      </Map>
    );
  }
}

export default TrainMap;
