const readline = require('readline');

// Create interface for reading input from terminal
const rl = readline.createInterface({
  input: process.stdin,
  output: process.stdout
});

// Game options
const options = ['rock', 'paper', 'scissors'];

// Track scores
let playerScore = 0;
let computerScore = 0;
let gameCount = 0;

// Welcome message
console.log('\n===================================');
console.log('Welcome to Rock, Paper, Scissors!');
console.log('===================================\n');
console.log('How to play:');
console.log('- Type "rock", "paper", or "scissors" to make your move');
console.log('- Type "score" to see the current score');
console.log('- Type "exit" or "quit" to end the game\n');

// Get computer's choice
function getComputerChoice() {
  const randomIndex = Math.floor(Math.random() * 3);
  return options[randomIndex];
}

// Determine winner
function determineWinner(playerChoice, computerChoice) {
  if (playerChoice === computerChoice) {
    return 'tie';
  }
  
  if (
    (playerChoice === 'rock' && computerChoice === 'scissors') ||
    (playerChoice === 'paper' && computerChoice === 'rock') ||
    (playerChoice === 'scissors' && computerChoice === 'paper')
  ) {
    return 'player';
  }
  
  return 'computer';
}

// Display current score
function displayScore() {
  console.log('\n----- CURRENT SCORE -----');
  console.log(`Games played: ${gameCount}`);
  console.log(`You: ${playerScore}`);
  console.log(`Computer: ${computerScore}`);
  console.log(`Ties: ${gameCount - playerScore - computerScore}`);
  console.log('------------------------\n');
}

// Play a round
function playRound(playerChoice) {
  const computerChoice = getComputerChoice();
  
  console.log(`You chose: ${playerChoice}`);
  console.log(`Computer chose: ${computerChoice}`);
  
  const winner = determineWinner(playerChoice, computerChoice);
  
  gameCount++;
  
  if (winner === 'tie') {
    console.log('It\'s a tie!');
  } else if (winner === 'player') {
    playerScore++;
    console.log('You win this round!');
  } else {
    computerScore++;
    console.log('Computer wins this round!');
  }
  
  displayScore();
}

// Game loop
function gameLoop() {
  rl.question('Enter your move (rock/paper/scissors): ', (input) => {
    const playerChoice = input.toLowerCase().trim();
    
    if (playerChoice === 'exit' || playerChoice === 'quit') {
      console.log('\nThanks for playing!');
      console.log('Final score:');
      displayScore();
      rl.close();
      return;
    }
    
    if (playerChoice === 'score') {
      displayScore();
      gameLoop();
      return;
    }
    
    if (!options.includes(playerChoice)) {
      console.log('Invalid choice! Please enter rock, paper, or scissors.');
      gameLoop();
      return;
    }
    
    playRound(playerChoice);
    gameLoop();
  });
}

// Start the game
gameLoop();