/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elb"
	ecc "github.com/ernestio/ernest-config-client"
	"github.com/nats-io/nats"
)

var nc *nats.Conn
var natsErr error

func eventHandler(m *nats.Msg) {
	var e Event

	err := e.Process(m.Data)
	if err != nil {
		return
	}

	if err = e.Validate(); err != nil {
		e.Error(err)
		return
	}

	err = deleteELB(&e)
	if err != nil {
		e.Error(err)
		return
	}

	e.Complete()
}

func mapListeners(ev *Event) []*elb.Listener {
	var l []*elb.Listener

	for _, port := range ev.ELBPorts {
		l = append(l, &elb.Listener{
			Protocol:         aws.String(port.Protocol),
			LoadBalancerPort: aws.Int64(port.FromPort),
			InstancePort:     aws.Int64(port.ToPort),
			InstanceProtocol: aws.String(port.Protocol),
			SSLCertificateId: aws.String(port.SSLCertID),
		})
	}

	return l
}

func deleteELB(ev *Event) error {
	creds := credentials.NewStaticCredentials(ev.DatacenterSecret, ev.DatacenterToken, "")
	svc := elb.New(session.New(), &aws.Config{
		Region:      aws.String(ev.DatacenterRegion),
		Credentials: creds,
	})

	// Delete Loadbalancer
	req := elb.DeleteLoadBalancerInput{
		LoadBalancerName: aws.String(ev.ELBName),
	}

	_, err := svc.DeleteLoadBalancer(&req)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	nc = ecc.NewConfig(os.Getenv("NATS_URI")).Nats()

	fmt.Println("listening for elb.create.aws")
	nc.Subscribe("elb.create.aws", eventHandler)

	runtime.Goexit()
}
