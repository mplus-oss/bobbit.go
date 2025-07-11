package shell

import (
	"fmt"
	"os"
)

func Print(message string) {
	fmt.Fprint(os.Stderr, message)
}

func Println(message string) {
	Print(message + "\n")
}

func Printf(format string, v ...any) {
	fmt.Fprintf(os.Stderr, format, v...)
}

func Printfln(format string, v ...any) {
	Printf(format+"\n", v...)
}

func Fatalln(exitCode int, message string) {
	Println(message)
	os.Exit(exitCode)
}

func Fatalfln(exitCode int, format string, v ...any) {
	Printfln(format, v...)
	os.Exit(exitCode)
}
