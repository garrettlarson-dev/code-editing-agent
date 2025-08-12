/**
 * FizzBuzz implementation
 * Prints numbers from 1 to n, but:
 * - For multiples of 3, prints "Fizz" instead
 * - For multiples of 5, prints "Buzz" instead
 * - For multiples of both 3 and 5, prints "FizzBuzz"
 */

function fizzBuzz(n) {
  for (let i = 1; i <= n; i++) {
    if (i % 3 === 0 && i % 5 === 0) {
      console.log("FizzBuzz");
    } else if (i % 3 === 0) {
      console.log("Fizz");
    } else if (i % 5 === 0) {
      console.log("Buzz");
    } else {
      console.log(i);
    }
  }
}

// Run FizzBuzz for numbers 1 to 100
fizzBuzz(100);

// If you want to run with a custom number, you can call the function like:
// fizzBuzz(15);

console.log("FizzBuzz completed!");
