package gouble_test

import (
	"errors"
	"testing"

	"github.com/camilacampos/gouble"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestMocks(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Mocks definition Suite")
}

type fakeMock struct {
	double *gouble.Double
}

type someType struct {
	bla string
	ble int
}

func newFakeMock() *fakeMock {
	return &fakeMock{}
}

func (mock *fakeMock) SetDouble(double *gouble.Double) {
	mock.double = double
}

// TODO: Use `returnValues`
func (mock *fakeMock) Method1() (bool, error) {
	resp := mock.double.MockFor("Method1")
	return resp.SuccessValue().(bool), resp.ErrorResponse
}

// TODO: Use `returnValues`
func (mock *fakeMock) Method2() (*someType, error) {
	resp := mock.double.MockFor("Method2")
	return resp.SuccessValue().(*someType), resp.ErrorResponse
}

// TODO: Use `returnValues`
func (mock *fakeMock) Method3() string {
	resp := mock.double.MockFor("Method3")
	return resp.SuccessValue().(string)
}

func (mock *fakeMock) Method4() (int, string, error) {
	resp := mock.double.MockFor("Method4")
	responses := resp.ReturnValues()

	return responses[0].(int), responses[1].(string), responses[2].(error)
}

var _ = Describe("Mocks test", func() {
	Context("Handling multiple values responses", func() {
		It("returns given values (2 return values)", func() {
			expectedResponse := &someType{bla: "bla", ble: 1}
			expectedErr := errors.New("some error")

			mock := newFakeMock()
			gouble.Mock().Allow(mock).ToReceive("Method2").AndReturn(expectedResponse, expectedErr)

			resp, err := mock.Method2()
			Expect(resp).To(Equal(expectedResponse))
			Expect(err).To(Equal(expectedErr))
		})

		It("returns given values (3+ return values)", func() {
			expectedI := 1
			expectedS := "bla"
			expectedErr := errors.New("some error")

			mock := newFakeMock()
			gouble.Mock().Allow(mock).ToReceive("Method4").AndReturn(expectedI, expectedS, expectedErr)

			i, s, err := mock.Method4()
			Expect(i).To(Equal(expectedI))
			Expect(s).To(Equal(expectedS))
			Expect(err).To(Equal(expectedErr))
		})

		It("panics when an wrong type is given wrong", func() {
			mock := newFakeMock()

			Expect(func() {
				gouble.Mock().Allow(mock).ToReceive("Method4").AndReturn(1)
			}).To(PanicWith("wrong number of return values for Method4. Expected: 3 | Got: 1"))
		})

		It("panics when an wrong type is given", func() {
			testCases := []struct {
				param1               interface{}
				param2               interface{}
				param3               interface{}
				expectedErrorMessage string
			}{
				{"wrong", "string", errors.New("some error"), ".* does not return given types. Index: 0 - Expected: int | Got: string$"},
				{1, 2, errors.New("some error"), ".* does not return given types. Index: 1 - Expected: string | Got: int$"},
				{1, "string", "should've been an error", ".* does not return given types. Index: 3 - Expected: error | Got: string$"},
			}

			for _, t := range testCases {
				mock := newFakeMock()

				Expect(func() {
					gouble.Mock().Allow(mock).ToReceive("Method4").AndReturn(t.param1, t.param2, t.param3)
				}).To(PanicWith(MatchRegexp(t.expectedErrorMessage)))
			}

		})

		It("does not panic when given correct number of arguments with correct types", func() {
			mock := newFakeMock()

			Expect(func() {
				gouble.Mock().Allow(mock).ToReceive("Method4").AndReturn(1, "string", errors.New("some error"))
			}).NotTo(Panic())
		})
	})

	It("returns given value for primitive types", func() {
		mock := newFakeMock()

		gouble.Mock().Allow(mock).ToReceive("Method1").AndReturnWithoutError(true)

		value, err := mock.Method1()
		Expect(err).NotTo(HaveOccurred())
		Expect(value).To(Equal(true))
	})

	It("returns given value for complex types", func() {
		mock := newFakeMock()

		expectedResponse := &someType{bla: "bla", ble: 1}

		gouble.Mock().Allow(mock).ToReceive("Method2").AndReturnWithoutError(expectedResponse)

		value, err := mock.Method2()
		Expect(err).NotTo(HaveOccurred())
		Expect(value).To(Equal(expectedResponse))
	})

	It("returns given error", func() {
		mock := newFakeMock()

		expectedErr := errors.New("some error")

		gouble.Mock().Allow(mock).ToReceive("Method1").AndThrowError(expectedErr)

		_, err := mock.Method1()
		Expect(err).To(MatchError(expectedErr))
	})

	Context("Handling default responses", func() {
		It("returns zeroed value as default for primitive types", func() {
			mock := newFakeMock()

			gouble.Mock().Allow(mock).ToReceive("Method1")

			value, err := mock.Method1()
			Expect(err).NotTo(HaveOccurred())
			Expect(value).To(Equal(false), "should return zero value of response type")
		})

		It("returns zeroed value as default for complex types", func() {
			mock := newFakeMock()

			gouble.Mock().Allow(mock).ToReceive("Method2")

			value, err := mock.Method2()
			Expect(err).NotTo(HaveOccurred())

			var expectedValue *someType
			Expect(value).To(Equal(expectedValue), "should return zero value of response type")
		})
	})

	Context("Handling unformatted mocks", func() {
		It("panics when Allow is not called", func() {
			Expect(func() {
				gouble.Mock().ToReceive("method")
			}).To(PanicWith("No mock was allowed!"))
		})

		It("panics when AndReturn is called without ToReceive", func() {
			mock := newFakeMock()

			Expect(func() {
				gouble.Mock().Allow(mock).AndReturnWithoutError(1)
			}).To(PanicWith("No method was allowed!"))
		})

		It("panics when AndThrowError is called without ToReceive", func() {
			mock := newFakeMock()

			Expect(func() {
				gouble.Mock().Allow(mock).AndThrowError(errors.New("some error"))
			}).To(PanicWith("No method was allowed!"))
		})

		It("panics when mock does not implement allowed function", func() {
			mock := newFakeMock()

			Expect(func() {
				gouble.Mock().Allow(mock).ToReceive("undefinedMethod")
			}).To(PanicWith(MatchRegexp(".* does not implement undefinedMethod")))
		})

		It("panics when mocked function has wrong definition", func() {
			mock := newFakeMock()

			Expect(func() {
				gouble.Mock().Allow(mock).ToReceive("Method1").AndReturnWithoutError(1)
			}).To(PanicWith(MatchRegexp(".* does not return value of type: int")))
		})

		It("panics when mocking an error response for functions that don't return error", func() {
			mock := newFakeMock()

			Expect(func() {
				gouble.Mock().Allow(mock).ToReceive("Method3").AndThrowError(errors.New("some error"))
			}).To(PanicWith(MatchRegexp(".* does not return error")))
		})

		It("panics when calling a not mocked function", func() {
			mock := newFakeMock()

			gouble.Mock().Allow(mock).ToReceive("Method1").AndReturnWithoutError(true)

			Expect(func() {
				mock.Method3()
			}).To(PanicWith("mock not defined for method Method3"))
		})
	})
})
