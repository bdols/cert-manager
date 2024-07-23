/*
Copyright 2020 The cert-manager Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package metrics

import (
	"time"

	cmacme "github.com/cert-manager/cert-manager/pkg/apis/acme/v1"
)

func (m *Metrics) ObserveACMEChallengeStateChange(ch *cmacme.Challenge) {
	m.ObserveACMEChallengeStateChangeWithTime(ch, time.Now())
}

func (m *Metrics) ObserveACMEChallengeStateChangeWithTime(ch *cmacme.Challenge, t time.Time) {
	labels := []string{
		ch.GetObjectMeta().GetNamespace(),
		ch.GetObjectMeta().GetName(),
		ch.Spec.IssuerRef.Name,
		ch.Spec.IssuerRef.Kind,
		ch.Spec.IssuerRef.Group,
		string(ch.Status.State),
	}
	m.certificateAcmeOrderStatus.WithLabelValues(labels...).Set(float64(t.Unix()))
}

func (m *Metrics) ObserveACMEOrderStateChange(o *cmacme.Order) {
	m.observeACMEOrderStateChangeWithTime(o, time.Now())
}

func (m *Metrics) observeACMEOrderStateChangeWithTime(o *cmacme.Order, t time.Time) {
	labels := []string{
		o.GetObjectMeta().GetNamespace(),
		o.GetObjectMeta().GetName(),
		o.Spec.IssuerRef.Name,
		o.Spec.IssuerRef.Kind,
		o.Spec.IssuerRef.Group,
		string(o.Status.State),
	}
	m.certificateAcmeOrderStatus.WithLabelValues(labels...).Set(float64(t.Unix()))
}

func (m *Metrics) ObserveACMEScheduled(count int, labels ...string) {
	m.acmeScheduled.WithLabelValues(labels...).Add(float64(count))
}

// ObserveACMERequestDuration increases bucket counters for that ACME client duration.
func (m *Metrics) ObserveACMERequestDuration(duration time.Duration, labels ...string) {
	m.acmeClientRequestDurationSeconds.WithLabelValues(labels...).Observe(duration.Seconds())
}

// IncrementACMERequestCount increases the acme client request counter.
func (m *Metrics) IncrementACMERequestCount(labels ...string) {
	m.acmeClientRequestCount.WithLabelValues(labels...).Inc()
}
