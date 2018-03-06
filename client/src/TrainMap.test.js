import React from 'react';
import ReactDOM from 'react-dom';
import TrainMap from './TrainMap';

it('renders without crashing', () => {
  const div = document.createElement('div');
  ReactDOM.render(<TrainMap />, div);
  ReactDOM.unmountComponentAtNode(div);
});
