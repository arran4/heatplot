package heatPlot

import "testing"

func TestSimplify(t *testing.T) {
	for _, eachTest := range []struct {
		InputFormula    string
		ExpectedFormula string
	}{
		{
			InputFormula:    "-(42 + 55.75) = X / 16.25",
			ExpectedFormula: "-(42 + 55.75) = X / 16.25",
		},
		{
			InputFormula:    "-(-(42 + 55.75)) = X",
			ExpectedFormula: "42 + 55.75 = X",
		},
		{
			InputFormula:    "1 - -(-(42 + 55.75)) = X",
			ExpectedFormula: "1 - (42 + 55.75) = X",
		},
		{
			InputFormula:    "42 Expm1 55.75 = X",
			ExpectedFormula: "42 Expm1 55.75 = X",
		},
		{
			InputFormula:    "42 Expm1 T = X",
			ExpectedFormula: "42 Expm1 T = X",
		},
		{
			InputFormula:    "42 % T = X",
			ExpectedFormula: "42 % T = X",
		},
		{
			InputFormula:    "-(-(-(42 + 55.75) - -(-(T + Y - X ^ T)))) = X / 16.25",
			ExpectedFormula: "-(42 + 55.75) - (T + Y - X ^ T) = X / 16.25",
		},
	} {
		f := ParseFunction(eachTest.InputFormula)
		controlOutputFormula := f.String()
		outputFormula := f.Simplify().String()
		if controlOutputFormula != eachTest.InputFormula {
			t.Logf("Controll error %#v doesn't match %#v", controlOutputFormula, eachTest.InputFormula)
			t.Fail()
		}
		if outputFormula != eachTest.ExpectedFormula {
			t.Logf("Formula  %#v", eachTest.InputFormula)
			t.Logf("Control  %#v", controlOutputFormula)
			t.Logf("Became   %#v", outputFormula)
			t.Logf("Expected %#v", eachTest.ExpectedFormula)

			t.Fail()
		}
	}
}
