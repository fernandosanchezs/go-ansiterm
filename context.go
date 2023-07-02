package ansiterm

type ansiContext struct {
	currentChar rune
	paramBuffer []rune
	interBuffer []rune
}
