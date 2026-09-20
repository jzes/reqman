package request

import (
	"errors"
	"testing"

	domainrequest "github.com/jzes/reqman/internal/domain/request"
)

func TestRequestResultMessageUpdatesSuccessfulTarget(t *testing.T) {
	target := &fakeTarget{}
	response := domainrequest.Response{Status: "200 OK"}

	RequestResultMessage{Response: response}.UpdateTarget(target)

	if target.inFlight {
		t.Fatal("inFlight = true, want false")
	}
	if target.responseError != "" {
		t.Fatalf("responseError = %q, want empty", target.responseError)
	}
	if target.response.Status != "200 OK" {
		t.Fatalf("response status = %q, want 200 OK", target.response.Status)
	}
	if !target.hasResponse {
		t.Fatal("hasResponse = false, want true")
	}
}

func TestRequestResultMessageUpdatesFailedTarget(t *testing.T) {
	target := &fakeTarget{response: domainrequest.Response{Status: "200 OK"}, hasResponse: true}

	RequestResultMessage{Err: errors.New("boom")}.UpdateTarget(target)

	if target.inFlight {
		t.Fatal("inFlight = true, want false")
	}
	if target.responseError != "boom" {
		t.Fatalf("responseError = %q, want boom", target.responseError)
	}
	if target.response.Status != "" || target.response.StatusCode != 0 || target.response.Body != "" {
		t.Fatalf("response = %+v, want zero", target.response)
	}
	if target.hasResponse {
		t.Fatal("hasResponse = true, want false")
	}
}

type fakeTarget struct {
	inFlight      bool
	responseError string
	response      domainrequest.Response
	hasResponse   bool
}

func (t *fakeTarget) SetRequestInFlight(inFlight bool) { t.inFlight = inFlight }
func (t *fakeTarget) SetResponseError(err string)      { t.responseError = err }
func (t *fakeTarget) SetResponse(res domainrequest.Response) {
	t.response = res
}
func (t *fakeTarget) SetHasResponse(hasResponse bool) { t.hasResponse = hasResponse }
