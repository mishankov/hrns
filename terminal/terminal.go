package terminal

import "github.com/fatih/color"

var reasoningColor = color.RGB(150, 150, 150)
var responseColor = color.New()
var errorColor = color.New(color.FgRed)

var toolNameColor = color.New(color.FgRed)
var toolArgsColor = color.New(color.FgHiYellow)

func PrintUserInputPrompt() {
	PrintNewLine()
	color.RGB(100, 100, 100).Print("> ")
}

func PrintReasoning(chunc string) {
	reasoningColor.Print(chunc)
}

func PrintResponse(chunc string) {
	responseColor.Print(chunc)
}

func PrintToolCall(name string, args any) {
	PrintNewLine()
	toolNameColor.Print(name)
	toolArgsColor.Print(" ", args)
	PrintNewLine()
}

func PrintError(error string) {
	errorColor.Print(error)
}

func PrintNewLine() {
	print("\n")
}
