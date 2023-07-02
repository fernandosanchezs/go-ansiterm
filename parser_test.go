package ansiterm

import (
	"fmt"
	"testing"
)

func TestStateTransitions(t *testing.T) {
	stateTransitionHelper(t, "CsiEntry", "Ground", alphabetics)
	stateTransitionHelper(t, "CsiEntry", "CsiParam", csiCollectables)
	stateTransitionHelper(t, "Escape", "CsiEntry", []rune{ANSI_ESCAPE_SECONDARY})
	stateTransitionHelper(t, "Escape", "OscString", []rune{0x5D})
	stateTransitionHelper(t, "Escape", "Ground", escapeToGroundBytes)
	stateTransitionHelper(t, "Escape", "EscapeIntermediate", intermeds)
	stateTransitionHelper(t, "EscapeIntermediate", "EscapeIntermediate", intermeds)
	stateTransitionHelper(t, "EscapeIntermediate", "EscapeIntermediate", executors)
	stateTransitionHelper(t, "EscapeIntermediate", "Ground", escapeIntermediateToGroundBytes)
	stateTransitionHelper(t, "OscString", "Ground", []rune{ANSI_BEL, 0x9C})
	stateTransitionHelper(t, "OscString", "OscString", printables)
	stateTransitionHelper(t, "OscString", "Escape", []rune{ANSI_ESCAPE_PRIMARY})
	stateTransitionHelper(t, "Ground", "Ground", executors)
}

func TestAnyToX(t *testing.T) {
	anyToXHelper(t, []rune{ANSI_ESCAPE_PRIMARY}, "Escape")
	anyToXHelper(t, []rune{DCS_ENTRY}, "DcsEntry")
	anyToXHelper(t, []rune{OSC_STRING}, "OscString")
	anyToXHelper(t, []rune{CSI_ENTRY}, "CsiEntry")
	anyToXHelper(t, toGroundBytes, "Ground")
}

func TestCollectCsiParams(t *testing.T) {
	parser, _ := createTestParser("CsiEntry")
	parser.Parse(csiCollectables)

	buffer := parser.context.paramBuffer
	bufferCount := len(buffer)

	if bufferCount != len(csiCollectables) {
		t.Errorf("Buffer:    %v", buffer)
		t.Errorf("CsiParams: %v", csiCollectables)
		t.Errorf("Buffer count failure: %d != %d", bufferCount, len(csiParams))
		return
	}

	for i, v := range csiCollectables {
		if v != buffer[i] {
			t.Errorf("Buffer:    %v", buffer)
			t.Errorf("CsiParams: %v", csiParams)
			t.Errorf("Mismatch at buffer[%d] = %d", i, buffer[i])
		}
	}
}

func TestParseParams(t *testing.T) {
	parseParamsHelper(t, []rune{}, []string{})
	parseParamsHelper(t, []rune{';'}, []string{})
	parseParamsHelper(t, []rune{';', ';'}, []string{})
	parseParamsHelper(t, []rune{'7'}, []string{"7"})
	parseParamsHelper(t, []rune{'7', ';'}, []string{"7"})
	parseParamsHelper(t, []rune{'7', ';', ';'}, []string{"7"})
	parseParamsHelper(t, []rune{'7', ';', ';', '8'}, []string{"7", "8"})
	parseParamsHelper(t, []rune{'7', ';', '8', ';'}, []string{"7", "8"})
	parseParamsHelper(t, []rune{'7', ';', ';', '8', ';', ';'}, []string{"7", "8"})
	parseParamsHelper(t, []rune{'7', '8'}, []string{"78"})
	parseParamsHelper(t, []rune{'7', '8', ';'}, []string{"78"})
	parseParamsHelper(t, []rune{'7', '8', ';', '9', '0'}, []string{"78", "90"})
	parseParamsHelper(t, []rune{'7', '8', ';', ';', '9', '0'}, []string{"78", "90"})
	parseParamsHelper(t, []rune{'7', '8', ';', '9', '0', ';'}, []string{"78", "90"})
	parseParamsHelper(t, []rune{'7', '8', ';', '9', '0', ';', ';'}, []string{"78", "90"})
}

func TestCursor(t *testing.T) {
	cursorSingleParamHelper(t, 'A', "CUU")
	cursorSingleParamHelper(t, 'B', "CUD")
	cursorSingleParamHelper(t, 'C', "CUF")
	cursorSingleParamHelper(t, 'D', "CUB")
	cursorSingleParamHelper(t, 'E', "CNL")
	cursorSingleParamHelper(t, 'F', "CPL")
	cursorSingleParamHelper(t, 'G', "CHA")
	cursorTwoParamHelper(t, 'H', "CUP")
	cursorTwoParamHelper(t, 'f', "HVP")
	funcCallParamHelper(t, []rune{'?', '2', '5', 'h'}, "CsiEntry", "Ground", []string{"DECTCEM([true])"})
	funcCallParamHelper(t, []rune{'?', '2', '5', 'l'}, "CsiEntry", "Ground", []string{"DECTCEM([false])"})
}

func TestErase(t *testing.T) {
	// Erase in Display
	eraseHelper(t, 'J', "ED")

	// Erase in Line
	eraseHelper(t, 'K', "EL")
}

func TestSelectGraphicRendition(t *testing.T) {
	funcCallParamHelper(t, []rune{'m'}, "CsiEntry", "Ground", []string{"SGR([0])"})
	funcCallParamHelper(t, []rune{'0', 'm'}, "CsiEntry", "Ground", []string{"SGR([0])"})
	funcCallParamHelper(t, []rune{'0', ';', '1', 'm'}, "CsiEntry", "Ground", []string{"SGR([0 1])"})
	funcCallParamHelper(t, []rune{'0', ';', '1', ';', '2', 'm'}, "CsiEntry", "Ground", []string{"SGR([0 1 2])"})
}

func TestScroll(t *testing.T) {
	scrollHelper(t, 'S', "SU")
	scrollHelper(t, 'T', "SD")
}

func TestPrint(t *testing.T) {
	parser, evtHandler := createTestParser("Ground")
	parser.Parse(printables)
	validateState(t, parser.currState, "Ground")

	for i, v := range printables {
		expectedCall := fmt.Sprintf("Print([%s])", string(v))
		actualCall := evtHandler.FunctionCalls[i]
		if actualCall != expectedCall {
			t.Errorf("Actual != Expected: %v != %v at %d", actualCall, expectedCall, i)
		}
	}
}

func TestClear(t *testing.T) {
	p, _ := createTestParser("Ground")
	fillContext(p.context)
	p.clear()
	validateEmptyContext(t, p.context)
}

func TestClearOnStateChange(t *testing.T) {
	clearOnStateChangeHelper(t, "Ground", "Escape", []rune{ANSI_ESCAPE_PRIMARY})
	clearOnStateChangeHelper(t, "Ground", "CsiEntry", []rune{CSI_ENTRY})
}

func TestC0(t *testing.T) {
	expectedCall := "Execute([" + string(rune(ANSI_LINE_FEED)) + "])"
	c0Helper(t, []rune{ANSI_LINE_FEED}, "Ground", []string{expectedCall})
	expectedCall = "Execute([" + string(rune(ANSI_CARRIAGE_RETURN)) + "])"
	c0Helper(t, []rune{ANSI_CARRIAGE_RETURN}, "Ground", []string{expectedCall})
}

func TestEscDispatch(t *testing.T) {
	funcCallParamHelper(t, []rune{'M'}, "Escape", "Ground", []string{"RI([])"})
}
