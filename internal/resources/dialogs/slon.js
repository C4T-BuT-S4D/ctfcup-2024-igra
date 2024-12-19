#!/usr/bin/env node

const readline = require('node:readline');
const { stdin: input, stdout: output } = require('node:process');

const rl = readline.createInterface({ input, output });

const askQuestions = async (questions) => {
    const answers = [];

    for (const question of questions) {
        // Use a promise to wait for each response
        const answer = await new Promise((resolve) => {
            rl.question(question, (response) => resolve(response));
        });
        answers.push(answer);
    }

    return answers;
};

function getRandomInt(min, max) {
    min = Math.ceil(min);
    max = Math.floor(max);
    return Math.floor(Math.random() * (max - min + 1)) + min;
}

let n1 = getRandomInt(1, 1000);
let n2 = getRandomInt(1, 1000000);

const regex = /WIN/gmi;


// Questions to ask
const questions = [
    "What is your name: ",
    "What is the number am I thinking of: ",
    "What is the second number am I thinking of: ",
];

// Ask the questions
askQuestions(questions).then((answers) => {
    const [name, answer1, answer2] = answers;

    if (regex.test(name)) {
        console.log("Sorry, you can't use that name.");
    }

    if (parseInt(answer1) === n1 && parseInt(answer2) === n2) {
        console.log(`YOU WIN, ${name.toUpperCase()}!`);
    } else {
        console.log(`TRY AGAIN, ${name.toUpperCase()}!`);
    }

    rl.close();
});




