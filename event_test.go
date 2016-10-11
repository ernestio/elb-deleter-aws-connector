/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"testing"
	"time"

	ecc "github.com/ernestio/ernest-config-client"
	"github.com/nats-io/nats"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	testEvent = Event{
		UUID:             "test",
		BatchID:          "test",
		ProviderType:     "aws",
		VPCID:            "vpc-0000000",
		DatacenterRegion: "eu-west-1",
		DatacenterSecret: "key",
		DatacenterToken:  "token",
		ELBName:          "test-elb",
		ELBIsPrivate:     false,
		ELBPorts: []Port{
			Port{
				FromPort: 80,
				ToPort:   80,
				Protocol: "HTTP",
			},
		},
		NetworkAWSIDs:       []string{"subnet-0000000"},
		InstanceAWSIDs:      []string{"i-0000000"},
		SecurityGroupAWSIDs: []string{"sg-0000000"},
	}
)

func waitMsg(ch chan *nats.Msg) (*nats.Msg, error) {
	select {
	case msg := <-ch:
		return msg, nil
	case <-time.After(time.Millisecond * 100):
	}
	return nil, errors.New("timeout")
}

func testSetup() (chan *nats.Msg, chan *nats.Msg) {
	doneChan := make(chan *nats.Msg, 10)
	errChan := make(chan *nats.Msg, 10)

	nc = ecc.NewConfig(os.Getenv("NATS_URI")).Nats()

	nc.ChanSubscribe("elb.delete.aws.done", doneChan)
	nc.ChanSubscribe("elb.delete.aws.error", errChan)

	return doneChan, errChan
}

func TestEvent(t *testing.T) {
	completed, errored := testSetup()

	Convey("Given I an event", t, func() {
		Convey("With valid fields", func() {
			valid, _ := json.Marshal(testEvent)
			Convey("When processing the event", func() {
				var e Event
				err := e.Process(valid)

				Convey("It should not error", func() {
					So(err, ShouldBeNil)
					msg, timeout := waitMsg(errored)
					So(msg, ShouldBeNil)
					So(timeout, ShouldNotBeNil)
				})

				Convey("It should load the correct values", func() {
					So(e.UUID, ShouldEqual, "test")
					So(e.BatchID, ShouldEqual, "test")
					So(e.ProviderType, ShouldEqual, "aws")
					So(e.VPCID, ShouldEqual, "vpc-0000000")
					So(e.DatacenterRegion, ShouldEqual, "eu-west-1")
					So(e.DatacenterSecret, ShouldEqual, "key")
					So(e.DatacenterToken, ShouldEqual, "token")
					So(e.ELBName, ShouldEqual, "test-elb")
					So(e.ELBIsPrivate, ShouldEqual, false)
					So(len(e.ELBPorts), ShouldEqual, 1)
					So(e.ELBPorts[0].FromPort, ShouldEqual, 80)
					So(e.ELBPorts[0].ToPort, ShouldEqual, 80)
					So(e.ELBPorts[0].Protocol, ShouldEqual, "HTTP")
					So(len(e.NetworkAWSIDs), ShouldEqual, 1)
					So(e.NetworkAWSIDs[0], ShouldEqual, "subnet-0000000")
					So(len(e.InstanceAWSIDs), ShouldEqual, 1)
					So(e.InstanceAWSIDs[0], ShouldEqual, "i-0000000")
					So(len(e.NetworkAWSIDs), ShouldEqual, 1)
					So(e.SecurityGroupAWSIDs[0], ShouldEqual, "sg-0000000")
				})
			})

			Convey("When validating the event", func() {
				var e Event
				e.Process(valid)
				err := e.Validate()

				Convey("It should not error", func() {
					So(err, ShouldBeNil)
					msg, timeout := waitMsg(errored)
					So(msg, ShouldBeNil)
					So(timeout, ShouldNotBeNil)
				})
			})

			Convey("When completing the event", func() {
				var e Event
				e.Process(valid)
				e.Complete()
				Convey("It should produce a elb.delete.aws.done event", func() {
					msg, timeout := waitMsg(completed)
					So(msg, ShouldNotBeNil)
					So(string(msg.Data), ShouldEqual, string(valid))
					So(timeout, ShouldBeNil)
					msg, timeout = waitMsg(errored)
					So(msg, ShouldBeNil)
					So(timeout, ShouldNotBeNil)
				})
			})

			Convey("When erroring the event", func() {
				log.SetOutput(ioutil.Discard)
				var e Event
				e.Process(valid)
				e.Error(errors.New("error"))
				Convey("It should produce a elb.delete.aws.error event", func() {
					msg, timeout := waitMsg(errored)
					So(msg, ShouldNotBeNil)
					So(string(msg.Data), ShouldContainSubstring, `"error":"error"`)
					So(timeout, ShouldBeNil)
					msg, timeout = waitMsg(completed)
					So(msg, ShouldBeNil)
					So(timeout, ShouldNotBeNil)
				})
				log.SetOutput(os.Stdout)
			})
		})

		Convey("With no datacenter vpc id", func() {
			testEventInvalid := testEvent
			testEventInvalid.VPCID = ""
			invalid, _ := json.Marshal(testEventInvalid)

			Convey("When validating the event", func() {
				var e Event
				e.Process(invalid)
				err := e.Validate()
				Convey("It should error", func() {
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldEqual, "Datacenter VPC ID invalid")
				})
			})
		})

		Convey("With no datacenter region", func() {
			testEventInvalid := testEvent
			testEventInvalid.DatacenterRegion = ""
			invalid, _ := json.Marshal(testEventInvalid)

			Convey("When validating the event", func() {
				var e Event
				e.Process(invalid)
				err := e.Validate()
				Convey("It should error", func() {
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldEqual, "Datacenter Region invalid")
				})
			})
		})

		Convey("With no datacenter access key", func() {
			testEventInvalid := testEvent
			testEventInvalid.DatacenterSecret = ""
			invalid, _ := json.Marshal(testEventInvalid)

			Convey("When validating the event", func() {
				var e Event
				e.Process(invalid)
				err := e.Validate()
				Convey("It should error", func() {
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldEqual, "Datacenter credentials invalid")
				})
			})
		})

		Convey("With no datacenter access token", func() {
			testEventInvalid := testEvent
			testEventInvalid.DatacenterToken = ""
			invalid, _ := json.Marshal(testEventInvalid)

			Convey("When validating the event", func() {
				var e Event
				e.Process(invalid)
				err := e.Validate()
				Convey("It should error", func() {
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldEqual, "Datacenter credentials invalid")
				})
			})
		})

		Convey("With no elb name", func() {
			testEventInvalid := testEvent
			testEventInvalid.ELBName = ""
			invalid, _ := json.Marshal(testEventInvalid)

			Convey("When validating the event", func() {
				var e Event
				e.Process(invalid)
				err := e.Validate()
				Convey("It should error", func() {
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldEqual, "ELB name is invalid")
				})
			})
		})
	})
}
