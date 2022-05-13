const { bingo } = require("./main-cli");

describe('bingo', () => {
	let logSpy;

	beforeEach(() => {
		logSpy = jest.spyOn(console, 'log');
	});

	test("returns board 1 as a winner", async () => {
		const nums = [7, 4, 9, 5, 11, 17, 23, 2, 0, 14, 21, 24, 10, 16, 13, 6, 15, 25, 12, 22, 18, 20, 8, 19, 3, 26, 1];
		const boardCount = 1;
		const boards = [
			[
				[22, 13, 17, 11, 0],
				[8, 2, 23, 4, 24],
				[21, 9, 14, 16, 7],
				[6, 10, 3, 18, 5],
				[1, 12, 20, 15, 19]
			]
		];

		return bingo(nums, boardCount, boards).then((winningBoard) => {
			expect(winningBoard).toEqual(1);
			expect(logSpy).toHaveBeenCalledWith('Board 1 wins!');
		});
	});

	test("passing empty args returns 0", async () => {
		const nums = [];
		const boardCount = 0;
		const boards = [];

		return bingo(nums, boardCount, boards).then((winningBoard) => {
			expect(winningBoard).toEqual(0);
			expect(logSpy).toHaveBeenCalledWith('No board wins.');
		});
	});
});