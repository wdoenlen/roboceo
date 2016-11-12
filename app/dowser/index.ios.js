/**
 * Sample React Native App
 * https://github.com/facebook/react-native
 * @flow
 */

import React, { Component } from 'react';
import {
  AppRegistry,
  StyleSheet,
  Text,
  View,
  DeviceEventEmitter
} from 'react-native';
import Button from 'react-native-button';

const ReactNativeHeading = require('react-native-heading');

export default class dowser extends Component {

  constructor(props) {
    super(props);
    this.state = {
      currentHeading: -1,
      nextHeading: Math.random() * 360,
      firstHeading: 0,
      secondHeading: 360
    };
  }

  componentDidMount() {

    ReactNativeHeading.start(1).then(didStart => {
      this.setState({
        headingIsSupported: didStart
      });
    });

    DeviceEventEmitter.addListener('headingUpdated', data => {
      this.setState({currentHeading: data.heading});
    });

  }

  componentWillUnmount() {
    ReactNativeHeading.stop();
    DeviceEventEmitter.removeAllListeners('headingUpdated');
  }

  _chooseNumInInterval(a, b, r) {
    if (typeof r === "undefined") {
      r = Math.random()
    }
    var lower = Math.min(a, b),
        upper = Math.max(a, b);
    return lower + r * (upper - lower);
  }

  _handleNewHeadingPress() {
    var nextHeading;
    if (this.state.firstHeading < this.state.secondHeading) {
      nextHeading = this._chooseNumInInterval(
        this.state.firstHeading, this.state.secondHeading
      );
    }
    // If the first heading is greater than the second heading then
    // we passed through zero when choosing hte second heading. We need to
    // choose a random number, then, between (firstHeading, 360) and
    // (0, secondHeading)
    else {
      var rand = Math.random();
          lenRange = (360 - this.state.firstHeading) + this.state.secondHeading;
      if (rand < this.secondHeading / lenRange) {
        nextHeading = this._chooseNumInInterval(0, this.state.secondHeading);
      }
      else {
        nextHeading = this._chooseNumInInterval(this.state.firstHeading, 360);
      }
    }
    this.setState({nextHeading: nextHeading});
  }

  _handleSetFirstHeadingPress() {
    this.setState({firstHeading: this.state.currentHeading});
  }

  _handleSetSecondHeadingPress() {
    this.setState({secondHeading: this.state.currentHeading});
  }

  _handleClearHeadingPress() {
    this.setState({
      firstHeading: 0,
      secondHeading: 360
    });
  }

  render() {
    return (
      <View style={styles.container}>
        <Text style={styles.welcome}>
          Please head towards {Math.round(this.state.nextHeading * 100) / 100}
        </Text>
        <Text style={styles.welcome}>
          Current heading is {Math.round(this.state.currentHeading * 100) / 100}
        </Text>
        <Button
          style={styles.button}
          onPress={() => this._handleNewHeadingPress()}>
          New Direction
        </Button>
        <Button
          style={styles.button}
          onPress={() => this._handleSetFirstHeadingPress()}>
          Set first heading
        </Button>
        <Button
          style={styles.button}
          onPress={() => this._handleSetSecondHeadingPress()}>
          Set second heading
        </Button>
        <Button
          style={styles.button}
          onPress={() => this._handleClearHeadingPress()}>
          Clear
        </Button>
      </View>
    );
  }
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: '#F5FCFF',
  },
  welcome: {
    fontSize: 20,
    textAlign: 'center',
    margin: 10,
  },
  button: {
    margin: 10
  },
  instructions: {
    textAlign: 'center',
    color: '#333333',
    marginBottom: 5,
  },
});

AppRegistry.registerComponent('dowser', () => dowser);
