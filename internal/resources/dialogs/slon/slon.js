#!/usr/bin/env node

const readline = require('node:readline');
const { stdin: input, stdout: output } = require('node:process');
const { JSDOM } = require("jsdom");
const DOMPurify = require("dompurify");

const rl = readline.createInterface({ input, output });

function getRandomInt(min, max) {
    min = Math.ceil(min);
    max = Math.floor(max);
    return Math.floor(Math.random() * (max - min + 1)) + min;
}

let n1 = getRandomInt(1, 1000);
let n2 = getRandomInt(1, 1000000);

const regex = /WIN/gmi;

rl.question("", (name) => {
    name = name.replaceAll(/([0-9A-Fa-f]{2})/g, (match, hex) => String.fromCharCode(parseInt(hex, 16)));

    const dom = new JSDOM('');
    const purify = DOMPurify(dom.window);

    purify.addHook(
        'uponSanitizeElement',
        function (currentNode, config) {
            if(config.tagName != "output"){
                return;
            }
            // check output
            if(currentNode.attributes["class"].value === "WIN"){
                if(currentNode.attributes["data-random1"].value != n1 || currentNode.attributes["data-random2"].value != n2){
                    currentNode.parentNode.removeChild(currentNode)
                    return;
                }
            }
            // You really can't win
            if(regex.test(currentNode.innerHTML.toUpperCase()))
                currentNode.parentNode.removeChild(currentNode)
            return;
        }

    );
    name = name.replaceAll('i','<').replaceAll('I','>')
    var sanitizedHtml = purify.sanitize(name,{WHOLE_DOCUMENT: true});
    let dom1 = new JSDOM(sanitizedHtml);
    let output = dom1.window.document.querySelector("body > output.WIN")
    console.log(output ? output.innerHTML.toUpperCase() : "YOU LOSE");
    rl.close();
});
