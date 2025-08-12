/**
 * FizzBuzz implementation in JavaScript
 * Run with: node fizzbuzz.js
 */

// Get command line arguments (optional)
const args = process.argv.slice(2);
const max = args.length > 0 ? parseInt(args[0]) : 100; // Default to 100 if no argument provided

// FizzBuzz implementation
for (let i = 1; i <= max; i++) {
    if (i % 3 === 0 && i % 5 === 0) {
        console.log('FizzBuzz');
    } else if (i % 3 === 0) {
        console.log('Fizz');
    } else if (i % 5 === 0) {
        console.log('Buzz');
    } else {
        console.log(i);
    }
}