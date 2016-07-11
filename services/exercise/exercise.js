"use strict";

var totalWorkoutTime = 60 *60 / 2;
var cumulativeWorkoutTime = 0;
var minExerciseTime = 60;
var maxExerciseTime = 0.33 * totalWorkoutTime;
var currentExercise = null;

function randomChoice(arr) {
  var rand = Math.random();
  var index = Math.floor(rand * arr.length);
  var elem = arr[index];
  arr.splice(index, 1);
  return elem;
}

function sayExercise() {
  if (currentExercise) {
    var utterance = new SpeechSynthesisUtterance(currentExercise.exercise.name);
    window.speechSynthesis.speak(utterance);
  }
}

function randomExercise() {
  var duration = Math.random() * (
    totalWorkoutTime - cumulativeWorkoutTime
  );
  if (duration < minExerciseTime) {
    duration = minExerciseTime;
  }
  else if (duration > maxExerciseTime) {
    duration = maxExerciseTime;
  }
  return {
    "exercise": randomChoice(exercisesJSON),
    "duration": duration
  }
}

function render(exercise) {
  currentExercise = exercise; // Set currentExercise
                              // for sayExercise in outer scope

  var imgContainer = document.getElementById("img-container");
  var spanCurrentExerciseName = document.getElementById("current-exercise");

  imgContainer.innerHTML = "";
  spanCurrentExerciseName.innerHTML = exercise.exercise.name +
                                      " for " + exercise.duration + " seconds";

  for (var i = 0; i < exercise.exercise.images.length; i++) {
    var nextImg = document.createElement("img");
    nextImg.src = exercise.exercise.images[i];
    imgContainer.appendChild(nextImg);
  }

}

function displayNextExercise(context) {
  var exercise = randomExercise();
  render(exercise);
  setTimeout(function() {
    cumulativeWorkoutTime += exercise.duration;
    if (cumulativeWorkoutTime > totalWorkoutTime) {
      clearInterval(context.sayIntervalId);
      return;
    }
    displayNextExercise()
  }, exercise.duration * 1000);
}

function start() {
  var intervalId = setInterval(sayExercise, 3000);
  var context = {"sayIntervalId": intervalId}
  displayNextExercise(context);
}
