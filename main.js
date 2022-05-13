const fs = require("fs");

const BINGO_SIZE = 5;

/**
 * A bingo game.
 * @param {number[]} nums The called numbers
 * @param {number} boardCount The board count
 * @param {number[][][]} boards A boardCount number of matrix arrays, representing bingo boards
 * @returns {number} The number of the winning board
 */
function bingo() {
	const lines = readInput("./board.txt");
	const nums = parseNums(lines[0]);
	const boards = parseBoards(lines.slice(2));

	for (let i = 0; i < boards.length; i++) {
		if (determineWin(nums, boards[i])) {
			console.log(`Board ${i + 1} wins!`);
			return i + 1;
		}
	}

	console.log("No board wins.");
	return 0;
}

/**
 * Determines a winner by checking if nums contains all
 * numbers of a board's row or column.
 * @param {nums} nums The called numbers
 * @param {number[][]} board A single bingo board
 * @returns {boolean} Whether the board wins
 */
function determineWin(nums, board) {
	let rowMatches;
	let colMatches;

	for (let row = 0; row < BINGO_SIZE; row++) {
		rowMatches = 0;
		for (let col = 0; col < BINGO_SIZE; col++) {
			nums.includes(board[row][col]) && rowMatches++;
		}
	}

	for (let col = 0; col < BINGO_SIZE; col++) {
		colMatches = 0;
		for (let row = 0; row < BINGO_SIZE; row++) {
			nums.includes(board[row][col]) && colMatches++;
		}
	}

	return colMatches === BINGO_SIZE || rowMatches === BINGO_SIZE;
}

/**
 * Reads input file (UTF-8 encoded) from the `input` path.
 * @param {string} input The filepath
 * @returns {string[]} An array of strings, representing each line
 */
function readInput(input) {
	const file = fs.readFileSync(input, "utf-8");
	return file.split(/\n/);;
}

/**
 * Parses the nums, separated by commas.
 * @param {string} numStr The numbers string
 * @returns {number[]} An array of the called numbers
 */
function parseNums(numStr) {
	return numStr.split(",").map(n => parseInt(n));
}

/**
 * Parses each board. Boards have an empty line as a separator.
 * @param {string} boardArr An array of the non-parsed boards
 * @returns {number[][][]} The resulting array of matrices
 */
function parseBoards(boardArr) {
	let boards = [];
	let board = [];

	boardArr.forEach((line) => {
		if (line !== "") {
			board.push(line.split(/\s+/).map(n => parseInt(n)));
		} else {
			boards.push(board);
			board = [];
		}
	});

	boards.push(board);
	return boards;
}

bingo();

module.exports = {
	bingo,
	parseNums,
	parseBoards,
	determineWin
};