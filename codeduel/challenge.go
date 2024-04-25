package codeduel

func RandomChallenge() Challenge {
	return Challenge{
		Id:    1,
		Title: "Add Two Numbers",
		Description: `## Goal
A complex palindrome is a string that is a palindrome when only its alphanumeric characters are considered and the case of the characters is ignored. The task is to determine whether a given string is a complex palindrome.

## Input
A string text made of ASCII characters.
## Output
if the string is a palindrome the program will return complex palindrome and the filtered text separated by a comma and a space. The filtered text will be all lowercase and will have a space between the words.

if the string is not a palindrome the program will return not a complex palindrome.`,
		TestCases: []TestCase{
			{Input: "1 2", Output: "3"},
			{Input: "3 4", Output: "7"},
			{Input: "5 6", Output: "11"},
			{Input: "7 8", Output: "15"},
			{Input: "9 10", Output: "19"},
			{Input: "11 12", Output: "23"},
			{Input: "13 14", Output: "27"},
			{Input: "15 16", Output: "31"},
			{Input: "17 18", Output: "35"},
			{Input: "19 20", Output: "39"},
		},
		HiddenTestCases: []TestCase{
			{Input: "100 1", Output: "101"},
		},
	}
}
