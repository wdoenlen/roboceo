import React, { Component } from 'react';
import {
  AppRegistry,
  Linking,
  Image,
  StyleSheet,
  Text,
  View,
  MapView,
  AsyncStorage,
  ScrollView,
  Navigator,
  DatePickerIOS,
  ActivityIndicator,
  TouchableHighlight,
  TouchableOpacity,
} from 'react-native';

import moment from 'moment';

class Row extends Component {
  render() {
    var event = this.props.event;
    return (<View style={ { flexDirection: 'row', padding: 10, paddingTop: 20, paddingBottom: 20, flex: 1, borderColor: '#DDD', borderBottomWidth: 1 } }>
              <View style={ { flex: 1 } }>
                <View>
                  <Text
                    numberOfLines={ 1 }
                    style={ { fontWeight: '500' } }>
                    { event.name }
                  </Text>
                </View>
                <View>
                  <Text>
                    { moment(event.start_time).local().format('MMM DD hh:mm') } -
                    { event.end_time ? moment(event.end_time).local().format('MMM DD hh:mm') : null }
                  </Text>
                </View>
                <View style={ { marginTop: 10 } }>
                  <Text numberOfLines={ 1 }>
                    { event.description }
                  </Text>
                </View>
              </View>
              <View>
                <TouchableOpacity
                  onPress={ () => {
                              Linking.openURL(`fb://event?id=${event.id}`)
                                .catch(() => Linking.openURL(`https://facebook.com/${event.id}`))
                            } }
                  style={ { paddingLeft: 20, paddingRight: 20, alignItems: 'center' } }>
                  <Text style={ { fontSize: 40 } }>
                    >
                  </Text>
                </TouchableOpacity>
              </View>
            </View>);
  }
}

class BrowsePage extends Component {

  state = {
    loading: false,
    events: [],
    region: {
      latitude: 36.6,
      longitude: 139.7,
      latitudeDelta: 1,
      longitudeDelta: 1.4,
    },
    setRegion: {
      latitude: 36.6,
      longitude: 139.7,
      latitudeDelta: 1,
      longitudeDelta: 1.4,
    },
    start: new Date(Date.now() - 1000 * 60 * 60 * 1),
    end: new Date(Date.now() + 1000 * 60 * 60 * 5),
  }

  async _loadInitialState() {
    var data = await AsyncStorage.getItem('region');
    var region = JSON.parse(data);
    if (region) {
      this._setRegion(region, {
        overwrite: true
      });
    }
  }

  componentWillMount() {
    this._loadInitialState();
  }

  _fetchEvents(region) {
    var currentRequest = this.state.currentRequest;
    if (currentRequest) {
      currentRequest.onload = null;
      currentRequest.abort();
      this.setState({
        currentRequest: null
      });
    }

    var bb = [
      region.latitude - region.latitudeDelta / 2,
      region.longitude - region.longitudeDelta / 2,
      region.latitude + region.latitudeDelta / 2,
      region.longitude + region.longitudeDelta / 2,
    ].join(',');

    var url = 'http://backend.machineexecutive.com:8000/events'
      + '?start=' + this.state.start.toISOString()
      + '&end=' + this.state.end.toISOString()
      + '&bb=' + bb;

    var xhr = new XMLHttpRequest();
    xhr.open('GET', url, true);
    // xhr.responseType = 'arraybuffer';

    var that = this;
    xhr.onload = function(e) {
      if (xhr.status != 200) {
        throw new Error('bad response: ' + xhr.status);
      }

      var events = JSON.parse(xhr.responseText);

      events = events || [];

      // AsyncStorage.setItem('lastEvents', JSON.stringify(events));
      that.setState({
        events: events,
        currentRequest: null,
      });
    }
    xhr.onerror = (e) => {
      console.warn('load failed: ', xhr.statusText);
      this.setState({
        currentRequest: null,
      });
    }

    this.setState({
      currentRequest: xhr
    });

    xhr.send(null);
  }

  componentWillUpdate(nextProps, nextState) {
    if (this.state.region !== nextState.region) {
      this._fetchEvents(nextState.region);
      AsyncStorage.setItem('region', JSON.stringify(nextState.region));
    }
  }

  _setRegion(region, options) {
    var updates = {
      region: region,
    };

    if (options && options.overwrite) {
      updates.setRegion = region;
    }

    this.setState(updates);
  }

  _onLocate() {
    navigator.geolocation.getCurrentPosition((position) => {
      var region = {
        latitude: position.coords.latitude,
        longitude: position.coords.longitude,
        latitudeDelta: 0.1,
        longitudeDelta: 0.1,
      };
      this._setRegion(region, {
        overwrite: true
      });
    });
  }

  render() {
    var events = this.state.events.sort(function(a, b) {
      return moment(a.start_time).toDate() - moment(b.start_time).toDate();
    });

    var rows = events.map(function(event, i) {
      return (<Row
                event={ event }
                key={ i } />);
    });

    var overlays = events
      .filter(function(event) {
        return event.latitude && event.longitude;
      })
      .map(function(event) {
        return {
          coordinates: [{
            latitude: event.latitude,
            longitude: event.longitude,
          }],
          strokeColor: '#f00',
          lineWidth: 3,
        };
      });

    return (
      <View style={ styles.container }>
        <View style={ { backgroundColor: '#DDD', flexDirection: 'row', padding: 10 } }>
          <Text style={ { flex: 1 } }>
            Start:
            { ' ' + moment(this.state.start).local().format('MMM DD hh:mm') }
          </Text>
          <Text style={ { flex: 1 } }>
            End:
            { ' ' + moment(this.state.end).local().format('MMM DD hh:mm') }
          </Text>
        </View>
        <View style={ { flex: 1, position: 'relative' } }>
          <MapView
            ref="map"
            style={ styles.map }
            overlays={ overlays }
            showsUserLocation={ true }
            showsCompass={ true }
            onRegionChangeComplete={ (region) => this._setRegion(region) }
            region={ this.state.setRegion } />
          <TouchableHighlight
            underlayColor="#DDD"
            onPress={ () => this._onLocate() }
            style={ { position: 'absolute', backgroundColor: 'white', bottom: 10, right: 10, padding: 10, borderColor: '#DDD', borderWidth: 1 } }>
            <Image
              style={ { width: 20, height: 20 } }
              source={ require('./locate.png') } />
          </TouchableHighlight>
        </View>
        <View style={ { backgroundColor: '#FF5722', padding: 5, paddingTop: 10, paddingBottom: 10, flexDirection: 'row' } }>
          <Text style={ { flex: 1, color: 'white' } }>
            { rows.length } results
          </Text>
          <ActivityIndicator
            color="white"
            style={ { opacity: this.state.currentRequest ? 1 : 0 } } />
        </View>
        <ScrollView style={ { flex: 1 } }>
          { rows }
        </ScrollView>
      </View>
      );
  }
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    paddingTop: 20,
  },
  map: {
    flex: 1,
  }
});

class EventBrowser extends Component {
  _renderScene(route, navigator) {
    switch (route.view) {
      case 'browse':
        return (<BrowsePage />);
        break;
      default:
        throw new Error('unknown view ' + route.view);
    }
  }

  render() {
    return (<Navigator
              style={ { flex: 1, backgroundColor: 'white' } }
              initialRoute={ { view: 'browse' } }
              renderScene={ this._renderScene.bind(this) } />);
  }
}

AppRegistry.registerComponent('EventBrowser', () => EventBrowser);
