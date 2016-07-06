/**
 * Sample React Native App
 * https://github.com/facebook/react-native
 * @flow
 */

import React, { Component } from 'react';
import { AppRegistry, StyleSheet, Text, View, PushNotificationIOS } from 'react-native';

class Alarm extends Component {
  state = {
    registered: false,
  }
  async _onToken(token) {
    var url = 'http://backend.machineexecutive.com:8004/register?token=' + token;
    var resp = await fetch(url);
    if (resp.status !== 200) {
      throw new Error('bad response ' + resp.status);
    }
    var txt = await resp.text();
    this.setState({
      registered: true
    });
  }

  componentWillMount() {
    PushNotificationIOS.addEventListener('register', this._onToken.bind(this));
  }
  componentDidMount() {
    PushNotificationIOS.requestPermissions();
  }
  render() {
    var status = this.state.registered ? 'Registered.' : 'Registering...'
    return (
      <View style={styles.container}>
        <Text style={styles.welcome}>
          {status}
        </Text>
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
  instructions: {
    textAlign: 'center',
    color: '#333333',
    marginBottom: 5,
  },
});

AppRegistry.registerComponent('Alarm', () => Alarm);
