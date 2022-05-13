const BINGO_SIZE = 5;

/**
 * A bingo game.
 * @param {number[]} nums The called numbers
 * @param {number} boardCount The board count
 * @param {number[][][]} boards A boardCount number of matrix arrays, representing bingo boards
 * @returns {number} The number of the winning board
 */
async function bingo(nums, boardCount, boards) {
	console.log("Welcome to Bingo!");

	if (!arguments.length) {
		var { nums, boardCount, boards } = await handleInput();
	}

	for (let boardIndex = 0; boardIndex < boardCount; boardIndex++) {
		if (determineWin(nums, boards[boardIndex])) {
			console.log(`Board ${boardIndex + 1} wins!`);
			return boardIndex + 1;
		}
	}

	console.log("No board wins.");
	return 0;
}

/**
 * Handles CLI input.
 * @returns {object} The nums array, board count and boards
 */
async function handleInput() {
	const rl = require("readline").createInterface({
		input: process.stdin,
		output: process.stdout
	});

	const question = (query) => new Promise(resolve => rl.question(query, resolve));

	const numStr = await question("Please type in your numbers, comma-separated: ")
	const nums = numStr.split(",").map((n) => parseInt(n));

	const boardCount = parseInt(await question("Type the count of bingo boards: "));

	const boards = [];

	for (let currentBoard = 0; currentBoard < boardCount; currentBoard++) {
		console.log(`Now enter bingo board ${currentBoard + 1}, each row with ${BINGO_SIZE} space-separated numbers.`);

		const board = [];

		for (let row = 0; row < BINGO_SIZE; row++) {
			let rowArr;

			while (true) {
				let rowStr = await question(`Ln${row + 1}: `);
				rowArr = rowStr.trim().split(/\s+/).map(n => +n);

				if (rowArr.length === BINGO_SIZE) break;
				else console.log(`Row must be ${BINGO_SIZE} numbers, try again.`)
			}

			board[row] = rowArr;
		}

		boards.push(board);
	}

	rl.close();

	return { nums, boardCount, boards };
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

	return rowMatches === 5 || colMatches === 5;
}

module.exports = {
	bingo,
	determineWin,
	handleInput
};

// bingo();

// bingo(
// 	[7, 4, 9, 5, 11, 17, 23, 2, 0, 14, 21, 24, 10, 16, 13, 6, 15, 25, 12, 22, 18, 20, 8, 19, 3, 26, 1],
// 	1,
// 	[
// 		[
// 			[22, 13, 17, 11, 0],
// 			[8, 2, 23, 4, 24],
// 			[21, 9, 14, 16, 7],
// 			[6, 10, 3, 18, 5],
// 			[1, 12, 20, 15, 19]
// 		]
// 	],
// )

