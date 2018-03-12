import React from 'react';
import ReactDOM from 'react-dom';
import TrainMap from './TrainMap';
import ReactMapboxGl, { GeoJSONLayer } from "react-mapbox-gl";

// Unfortunately, Mapbox GL does not play nicely with non-browser
// testing environments, so the test suite fails to launch.
it('renders without crashing', () => {
  const div = document.createElement('div');
  ReactDOM.render(<TrainMap />, div);
  ReactDOM.unmountComponentAtNode(div);
});