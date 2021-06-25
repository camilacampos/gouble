package gouble

import (
	"fmt"
	"reflect"
)

type MockResponse struct {
	method reflect.Method

	ErrorResponse    error         // TODO: Remove
	successResponse  reflect.Value // TODO: Remove
	successResponses []reflect.Value
}

// TODO: Remove
func (resp *MockResponse) SuccessValue() interface{} {
	return resp.successResponse.Interface()
}

func (resp *MockResponse) ReturnValues() []interface{} {
	resps := make([]interface{}, 0)
	for _, resp := range resp.successResponses {
		resps = append(resps, resp.Interface())
	}
	return resps
}

type iMock interface {
	SetDouble(*Double)
}

type Double struct {
	mock           iMock
	mockResponses  map[string]*MockResponse
	lastMethodName string
}

func Mock() *Double {
	return &Double{mockResponses: make(map[string]*MockResponse)}
}

func (d *Double) Allow(mock iMock) *Double {
	d.mock = mock
	mock.SetDouble(d)
	return d
}

func (d *Double) ToReceive(methodName string) *Double {
	if d.mock == nil {
		panic("No mock was allowed!")
	}

	mockType := typeOf(d.mock)
	m, ok := mockType.MethodByName(methodName)
	if !ok {
		panic(fmt.Sprintf("%s does not implement %s", mockType.String(), methodName))
	}

	// TODO: Have default values for `successResponses`
	defaultResponse := reflect.Indirect(reflect.New(m.Type.Out(0)))
	d.mockResponses[methodName] = &MockResponse{successResponse: defaultResponse, method: m}
	d.lastMethodName = methodName

	return d
}

func (d *Double) lastMethod() reflect.Method {
	lastResponse := d.mockResponses[d.lastMethodName]
	if lastResponse == nil {
		panic("No method was allowed!")
	}
	return lastResponse.method
}

func (d *Double) resetLastMethod() {
	d.lastMethodName = ""
}

func (d *Double) AndReturn(responses ...interface{}) {
	if len(responses) != d.lastMethod().Type.NumOut() {
		msg := fmt.Sprintf(
			"wrong number of return values for %s. Expected: %v | Got: %v",
			d.lastMethod().Name,
			d.lastMethod().Type.NumOut(),
			len(responses),
		)
		panic(msg)
	}

	resps := make([]reflect.Value, 0)
	var errResponse error
	for i, r := range responses {
		expectedType := d.lastMethod().Type.Out(i)
		if expectedType.String() == "error" {
			err, ok := r.(error)
			if !ok {
				panic(mismatchTypeErrorMessage(d, expectedType, r, i))
			}
			errResponse = err
		} else if expectedType.String() != typeToString(r) {
			panic(mismatchTypeErrorMessage(d, expectedType, r, i))
		}
		resps = append(resps, reflect.ValueOf(r))
	}

	d.mockResponses[d.lastMethod().Name] = &MockResponse{
		successResponse:  resps[0],
		successResponses: resps,
		ErrorResponse:    errResponse,
	}
	d.resetLastMethod()
}

func mismatchTypeErrorMessage(d *Double, expected reflect.Type, got interface{}, index int) string {
	return fmt.Sprintf(
		"%s of %s [%s] does not return given types. Index: %v - Expected: %s | Got: %s",
		d.lastMethod().Name,
		typeToString(d.mock),
		d.lastMethod().Type.String(),
		index,
		expected.String(),
		typeToString(got),
	)
}

// TODO: Remove (use one `AndReturn`)
func (d *Double) AndReturnWithError(successResponse interface{}, err error) {
	expectedType := d.lastMethod().Type.Out(0)
	if expectedType != typeOf(successResponse) {
		msg := fmt.Sprintf(
			"%s of %s does not return value of type: %s",
			d.lastMethod().Type.String(),
			typeToString(d.mock),
			typeToString(successResponse),
		)
		panic(msg)
	}

	lastType := d.lastMethod().Type.Out(d.lastMethod().Type.NumOut() - 1)
	if lastType.String() != "error" {
		msg := fmt.Sprintf(
			"%s of %s does not return error",
			d.lastMethod().Type.String(),
			typeToString(d.mock),
		)
		panic(msg)
	}

	d.mockResponses[d.lastMethod().Name] = &MockResponse{
		successResponse: reflect.ValueOf(successResponse),
		ErrorResponse:   err,
	}
	d.resetLastMethod()
}

// TODO: Remove (use one `AndReturn`)
func (d *Double) AndReturnWithoutError(successResponse interface{}) {
	expectedType := d.lastMethod().Type.Out(0)
	if expectedType != typeOf(successResponse) {
		msg := fmt.Sprintf(
			"%s of %s does not return value of type: %s",
			d.lastMethod().Type.String(),
			typeToString(d.mock),
			typeToString(successResponse),
		)
		panic(msg)
	}

	d.mockResponses[d.lastMethod().Name] = &MockResponse{successResponse: reflect.ValueOf(successResponse)}
	d.resetLastMethod()
}

// TODO: Remove (use one `AndReturn`)
func (d *Double) AndThrowError(err error) {
	lastType := d.lastMethod().Type.Out(d.lastMethod().Type.NumOut() - 1)

	if lastType.String() != "error" {
		msg := fmt.Sprintf(
			"%s of %s does not return error",
			d.lastMethod().Type.String(),
			typeToString(d.mock),
		)
		panic(msg)
	}

	d.MockFor(d.lastMethod().Name).ErrorResponse = err
	d.resetLastMethod()
}

func (d *Double) MockFor(method string) *MockResponse {
	mock := d.mockResponses[method]
	if mock == nil {
		panic(fmt.Sprintf("mock not defined for method %s", method))
	}

	return mock
}

func typeOf(obj interface{}) reflect.Type {
	return reflect.TypeOf(obj)
}

func typeToString(obj interface{}) string {
	return typeOf(obj).String()
}
