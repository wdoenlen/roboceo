var totalWorkoutTime = 60*60 / 2;
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

function renderExercise(exercise, exerciseTime) {
  var nextExercise = randomChoice(exercisesJSON);
  var imgContainer = document.getElementById("img-container");
  var spanCurrentExerciseName = document.getElementById("current-exercise");

  imgContainer.innerHTML = "";
  spanCurrentExerciseName.innerHTML = nextExercise.name +
                                      " for " + exerciseTime + " seconds";

  for (var i = 0; i < nextExercise.images.length; i++) {
    var nextImg = document.createElement("img");
    nextImg.src = nextExercise.images[i];
    imgContainer.appendChild(nextImg);
  }

  console.log(exerciseTime);
  console.log(nextExercise);

}

function sayExercise() {
  if (currentExercise) {
    var utterance = new SpeechSynthesisUtterance(currentExercise.name);
    window.speechSynthesis.speak(utterance);
  }
}

function sleep (time) {
  return new Promise((resolve) => setTimeout(resolve, time));
}

function createExerciseRoutine() {

  var intervalId = setInterval(sayExercise, 3000);

  while (cumulativeWorkoutTime < totalWorkoutTime) {
    var exerciseTime = Math.random() * (
      totalWorkoutTime - cumulativeWorkoutTime
    );
    if (exerciseTime < minExerciseTime) {
      exerciseTime = minExerciseTime;
    }
    else if (exerciseTime > maxExerciseTime) {
      exerciseTime = maxExerciseTime;
    }
    currentExercise = randomChoice(exercisesJSON);
    renderExercise(currentExercise, exerciseTime);

    cumulativeWorkoutTime += exerciseTime;
    console.log("Cum workout time is " + cumulativeWorkoutTime);

  }

  clearInterval(intervalId);

}
