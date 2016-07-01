import React, { Component } from 'react';
import {
  AppRegistry,
  Clipboard,
  Linking,
  Image,
  StyleSheet,
  Text,
  View,
  MapView,
  AlertIOS,
  AsyncStorage,
  ScrollView,
  Navigator,
  DatePickerIOS,
  ActivityIndicator,
  TouchableHighlight,
  TouchableOpacity,
} from 'react-native';

class PickOriginPage extends Component {

  state = {
    region: {
      latitude: 36.6,
      longitude: 139.7,
      latitudeDelta: 0.005,
      longitudeDelta: 0.005,
    },
    initialRegion: {
      latitude: 36.6,
      longitude: 139.7,
      latitudeDelta: 0.005,
      longitudeDelta: 0.005,
    },
  }

  componentWillUpdate(nextProps, nextState) {
    AsyncStorage.setItem('PickOriginState', JSON.stringify(nextState));
  }

  async componentWillMount() {
    var stateData = await AsyncStorage.getItem('PickOriginState');
    if (stateData) {
      this.setState(JSON.parse(stateData));
    }
  }

  _onLocate() {
    navigator.geolocation.getCurrentPosition((position) => {
      var region = {
        latitude: position.coords.latitude,
        longitude: position.coords.longitude,
        latitudeDelta: this.state.region.latitudeDelta || 0.005,
        longitudeDelta: this.state.region.longitudeDelta || 0.005,
      };
      this.setState({
        initialRegion: region,
        region: region,
      });
    }, null, {
      enableHighAccuracy: true,
      timeout: 20000,
      maximumAge: 1000
    });
  }

  render() {
    return (
      <View style={ styles.container }>
        <View style={ { flex: 1, position: 'relative' } }>
          <MapView
            ref="map"
            style={ styles.map }
            showsUserLocation={ true }
            showsCompass={ true }
            onRegionChangeComplete={ (region) => this.setState({
                                       region: region
                                     }) }
            region={ this.state.initialRegion } />
          <View
            pointerEvents="none"
            style={ { backgroundColor: 'transparent', alignItems: 'center', justifyContent: 'center', position: 'absolute', top: 0, right: 0, bottom: 0, left: 0 } }>
            <Text style={ { fontSize: 30, color: '#D0021B' } }>
              ✛
            </Text>
          </View>
          <TouchableHighlight
            underlayColor="#DDD"
            onPress={ () => this._onLocate() }
            style={ { position: 'absolute', backgroundColor: 'white', bottom: 10, right: 10, padding: 10, borderColor: '#DDD', borderWidth: 1 } }>
            <Image
              style={ { width: 20, height: 20 } }
              source={ require('./locate.png') } />
          </TouchableHighlight>
        </View>
        <TouchableHighlight
          onPress={ () => this.props.onNext(this.state.region) }
          underlayColor="#3EB399"
          style={ styles.button }>
          <Text style={ { fontSize: 20 } }>
            Set Origin
          </Text>
        </TouchableHighlight>
      </View>
      );
  }
}

const styles = StyleSheet.create({
  button: {
    alignItems: 'center',
    padding: 20,
    backgroundColor: '#50E3C2'
  },
  container: {
    flex: 1,
    backgroundColor: 'white',
  },
  map: {
    flex: 1,
  }
});

const PlaceTypes = ['accounting', 'airport', 'amusement_park', 'aquarium', 'art_gallery', 'atm', 'bakery', 'bank', 'bar', 'beauty_salon', 'bicycle_store', 'book_store', 'bowling_alley', 'bus_station', 'cafe', 'campground', 'car_dealer', 'car_rental', 'car_repair', 'car_wash', 'casino', 'cemetery', 'church', 'city_hall', 'clothing_store', 'convenience_store', 'courthouse', 'dentist', 'department_store', 'doctor', 'electrician', 'electronics_store', 'embassy', 'establishment', 'finance', 'fire_station', 'florist', 'food', 'funeral_home', 'furniture_store', 'gas_station', 'general_contractor', 'grocery_or_supermarket', 'gym', 'hair_care', 'hardware_store', 'health', 'hindu_temple', 'home_goods_store', 'hospital', 'insurance_agency', 'jewelry_store', 'laundry', 'lawyer', 'library', 'liquor_store', 'local_government_office', 'locksmith', 'lodging', 'meal_delivery', 'meal_takeaway', 'mosque', 'movie_rental', 'movie_theater', 'moving_company', 'museum', 'night_club', 'painter', 'park', 'parking', 'pet_store', 'pharmacy', 'physiotherapist', 'place_of_worship', 'plumber', 'police', 'post_office', 'real_estate_agency', 'restaurant', 'roofing_contractor', 'rv_park', 'school', 'shoe_store', 'shopping_mall', 'spa', 'stadium', 'storage', 'store', 'subway_station', 'synagogue', 'taxi_stand', 'train_station', 'travel_agency', 'university', 'veterinary_care', 'zoo'];

class PickTypesPage extends Component {

  state = {
    selections: { },
    randomMode: true,
  }

  componentWillUpdate(nextProps, nextState) {
    AsyncStorage.setItem('PickTypesState', JSON.stringify(nextState));
  }

  async componentWillMount() {
    var stateData = await AsyncStorage.getItem('PickTypesState');
    if (stateData) {
      this.setState(JSON.parse(stateData));
    }
  }

  selectedTypes() {
    var selected = [];

    var picked = this.state.selections;

    if (this.state.randomMode || Object.keys(picked).length === 0) {
      for (var i = 0; i < PlaceTypes.length; i++) {
        if (Math.random() < 0.2) {
          selected.push(PlaceTypes[i]);
        }
      }

      // do at least one
      if (selected.length === 0) {
        var i = Math.floor(Math.random(PlaceTypes.length));
        selected.push(PlaceTypes[i]);
      }

    } else {
      for (var t in picked) {
        if (this.state.selections[t]) {
          selected.push(t);
        }
      }
    }

    return selected;
  }

  _toggleType(type) {
    if (type === 'random') {
      this.setState({
        randomMode: true,
        selections: { },
      });
      return;
    }

    var update = this.state.selections;
    update[type] = !update[type];
    this.setState({
      selections: update,
      randomMode: false,
    });
  }

  render() {

    var items = ['random'].concat(PlaceTypes).map((type) => {
      var boxStyle = { };
      var textStyle = { };

      var isSelected = this.state.selections[type];


      if (type === 'random' && this.state.randomMode) {
        boxStyle.backgroundColor = '#D0021B';
        textStyle.color = 'white';

      } else if (isSelected && !this.state.randomMode) {
        boxStyle.backgroundColor = '#D0021B';
        textStyle.color = 'white';

      } else {
        boxStyle.borderColor = '#AAA';
        boxStyle.borderWidth = 1;
        textStyle.color = '#999';
      }

      return (<TouchableHighlight
                underlayColor="#990213"
                onPress={ () => this._toggleType(type) }
                key={ type }
                style={ [{ alignItems: 'center', justifyContent: 'center', width: 90, height: 90, padding: 10, margin: 10 }, boxStyle] }>
                <View style={ { alignItems: 'center' } }>
                  <Text style={ [{ fontWeight: '500' }, textStyle] }>
                    { type }
                  </Text>
                </View>
              </TouchableHighlight>);
    });

    return (<View style={ styles.container }>
              <ScrollView
                style={ { flex: 1 } }
                contentContainerStyle={ { paddingTop: 20, justifyContent: 'space-around', flexDirection: 'row', flexWrap: 'wrap' } }>
                { items }
              </ScrollView>
              <TouchableHighlight
                onPress={ () => this.props.onNext(this.selectedTypes()) }
                underlayColor="#3EB399"
                style={ styles.button }>
                <Text style={ { fontSize: 20 } }>
                  Set Categories
                </Text>
              </TouchableHighlight>
            </View>);
  }
}

class LoadingPage extends Component {
  render() {
    return (<View style={ [styles.container, { alignItems: 'center', justifyContent: 'center', backgroundColor: '#50E3C2' }] }>
              <ActivityIndicator
                size="large"
                style={ { margin: 20 } } />
              <Text>
                Loading...
              </Text>
            </View>);
  }
}

function getDestination(origin, types) {
  var url = 'http://backend.machineexecutive.com:8001/'
  + '?lat=' + origin.latitude
  + '&lng=' + origin.longitude
  + '&types=' + types.join(',');

  return fetch(url)
    .then(function(resp) {
      if (resp.status !== 200) {
        return Promise.reject(new Error('bad response: ' + resp.status));
      }

      return resp.json();
    })
    .then(function(data) {
      return {
        name: data.name,
        latitude: data.lat,
        longitude: data.lng,
      };
    });
}

class ViewPage extends Component {

  state = {
    nameHidden: true,
  }

  render() {
    var orig = this.props.origin;
    var dest = this.props.destination;

    var overlays = [
      {
        coordinates: [orig, dest],
        strokeColor: '#0002',
        lineWidth: 3,
      },
      {
        coordinates: [orig],
        strokeColor: '#0F0',
        lineWidth: 15,
      },
      {
        coordinates: [dest],
        strokeColor: '#F00',
        lineWidth: 15,
      },
    ];

    var region = {
      latitude: dest.latitude,
      longitude: dest.longitude,
      latitudeDelta: 0.005,
      longitudeDelta: 0.005,
    };

    return (<View style={ [styles.container, { paddingTop: 40 }] }>
              <View style={ { alignItems: 'center' } }>
                <TouchableOpacity onPress={ () => this.setState({
                                              nameHidden: !this.state.nameHidden
                                            }) }>
                  <Text
                    numberOfLines={ 1 }
                    style={ { fontWeight: 'bold', fontSize: 20, paddingLeft: 20, paddingRight: 20, } }>
                    { this.state.nameHidden ? 'Your Destination' : dest.name }
                  </Text>
                </TouchableOpacity>
              </View>
              <View style={ { flexDirection: 'row', padding: 10 } }>
                <TouchableHighlight
                  underlayColor="#CCC"
                  onPress={ this.props.onCopy.bind(this) }
                  style={ { alignItems: 'center', padding: 10, margin: 10, flex: 1, borderColor: '#EEE', borderWidth: 1 } }>
                  <Text>
                    Copy
                  </Text>
                </TouchableHighlight>
                <TouchableHighlight
                  underlayColor="#CCC"
                  onPress={ this.props.onMap.bind(this) }
                  style={ { alignItems: 'center', padding: 10, margin: 10, flex: 1, borderColor: '#EEE', borderWidth: 1 } }>
                  <Text>
                    Map
                  </Text>
                </TouchableHighlight>
              </View>
              <MapView
                ref="map"
                style={ styles.map }
                overlays={ overlays }
                showsUserLocation={ true }
                showsCompass={ true }
                region={ region } />
              <View>
                <TouchableHighlight
                  onPress={ this.props.onAnother.bind(this) }
                  underlayColor="#3EB399"
                  style={ styles.button }>
                  <Text style={ { fontSize: 20 } }>
                    Try Again
                  </Text>
                </TouchableHighlight>
              </View>
            </View>);
  }
}

class App extends Component {

  coordinates() {
    var dest = this.state.destination;
    if (!dest) {
      return null;
    }
    return `${dest.latitude},${dest.longitude}`;
  }

  _renderScene(route, navigator) {
    switch (route.view) {
      case 'pickorigin':
        return (<PickOriginPage onNext={ (origin) => {
                           this.setState({
                             origin: origin
                           });
                           navigator.push({
                             view: 'picktypes'
                           })
                         } } />);

      case 'picktypes':
        return (<PickTypesPage onNext={ (selected) => {
                          this.setState({
                            types: selected
                          });
                          navigator.push({
                            view: 'loading'
                          });
                          getDestination(this.state.origin, selected).then((destination) => {
                            this.setState({
                              destination: destination
                            });
                            navigator.push({
                              view: 'view'
                            });
                          });
                        } } />);

      case 'view':
        return (<ViewPage
                  onCopy={ () => {
                             Clipboard.setString(this.coordinates())
                             AlertIOS.alert('copied to clipboard');
                           } }
                  onMap={ () => {
                            Linking.openURL('comgooglemaps://?daddr=' + this.coordinates())
                              .catch(() => Linking.openURL('maps://?daddr=' + this.coordinates()))
                          } }
                  onAnother={ () => navigator.popToTop() }
                  origin={ this.state.origin }
                  destination={ this.state.destination } />);

      case 'loading':
        return (<LoadingPage />);

      default:
        throw new Error('unknown view ' + route.view);
    }
  }

  render() {
    return (<Navigator
              style={ { flex: 1 } }
              initialRoute={ { view: 'pickorigin' } }
              renderScene={ this._renderScene.bind(this) } />);
  }
}

AppRegistry.registerComponent('PlacePicker', () => App);
