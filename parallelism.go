package main

var (
	safeParallellGlChannelInput  = make(chan func() []any, 50)
	safeParallellGlChannelOutput = make(chan []any, 50)
)

func performOperations() {
	select {
	case op := <-safeParallellGlChannelInput:
		ret := op()

		if len(ret) > 0 {
			safeParallellGlChannelOutput <- ret
		}
	default:

	}
}
