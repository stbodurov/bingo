const { readFileSync } = require("fs");
const { bingo, parseNums, parseBoard, determineWin, parseBoards } = require("./main");

test("bingo() should return number of winning board", () => {
	expect(bingo()).toEqual(1);
})

test("parseNums() return value should have length of 5", () => {
	expect(parseNums("1,2,3,4,5").length).toEqual(5);
})

test("parseBoards() should return 3 matrices", () => {
	const boardStr = readFileSync("./board.txt", "utf-8").split("\n").slice(2);
	const boards = parseBoards(boardStr);
	const expectedBoard = [
		[3, 15, 0, 2, 22],
		[9, 18, 13, 17, 5],
		[19, 8, 7, 25, 23],
		[20, 11, 10, 24, 4],
		[14, 21, 16, 12, 6],
	]

	expect(boards.length).toEqual(3);
	expect(boards[1]).toEqual(expectedBoard);
})

test("determineWin()", () => {
	const nums = [7, 4, 9, 5, 11, 17, 23, 2, 0, 14, 21, 24, 10, 16, 13, 6, 15, 25, 12, 22, 18, 20, 8, 19, 3, 26, 1];
	const board = [
		[7, 13, 17, 11, 0],
		[4, 2, 23, 4, 24],
		[9, 9, 14, 16, 7],
		[5, 10, 3, 18, 5],
		[11, 12, 20, 15, 19]
	];

	const result = determineWin(nums, board);

	expect(result).toBe(true);
});
