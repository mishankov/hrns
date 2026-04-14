package terminal

import "github.com/fatih/color"

var reasoningColor = color.RGB(150, 150, 150)
var responseColor = color.New()
var errorColor = color.New(color.FgRed)
var harnessMessageColor = color.New(color.FgBlue)

var toolNameColor = color.New(color.FgRed)
var toolArgsColor = color.New(color.FgHiYellow)

func PrintUserInputPrompt() {
	PrintNewLine()
	color.RGB(100, 100, 100).Print("> ")
}

func PrintReasoningChunc(chunc string) {
	reasoningColor.Print(chunc)
}

func PrintResponseChunc(chunc string) {
	responseColor.Print(chunc)
}

func PrintToolCall(name string, args any) {
	toolNameColor.Print(name)
	toolArgsColor.Print(" ", args)
	PrintNewLine()
}

func PrintError(error string) {
	errorColor.Print(error)
	PrintNewLine()
}

func PrintHarnessMessage(message string) {
	harnessMessageColor.Print(message)
	PrintNewLine()
}

func PrintNewLine() {
	print("\n")
}
