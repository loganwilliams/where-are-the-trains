import React from 'react';
import ReactDOM from 'react-dom';
import './index.css';
import TrainMap from './TrainMap';
import registerServiceWorker from './registerServiceWorker';

ReactDOM.render(<TrainMap />, document.getElementById('root'));
registerServiceWorker();
