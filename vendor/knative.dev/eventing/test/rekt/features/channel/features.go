/*
Copyright 2021 The Knative Authors

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

package channel

import (
	"context"
	"fmt"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/test"
	"github.com/google/uuid"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/reconciler-test/pkg/eventshub"
	"knative.dev/reconciler-test/pkg/eventshub/assert"
	"knative.dev/reconciler-test/pkg/feature"
	"knative.dev/reconciler-test/pkg/manifest"
	"knative.dev/reconciler-test/resources/svc"

	eventasssert "knative.dev/reconciler-test/pkg/eventshub/assert"

	"knative.dev/eventing/test/rekt/resources/channel"
	"knative.dev/eventing/test/rekt/resources/channel_impl"
	"knative.dev/eventing/test/rekt/resources/containersource"
	"knative.dev/eventing/test/rekt/resources/delivery"
	"knative.dev/eventing/test/rekt/resources/source"
	"knative.dev/eventing/test/rekt/resources/subscription"
)

func ChannelChain(length int, createSubscriberFn func(ref *duckv1.KReference, uri string) manifest.CfgFn) *feature.Feature {
	f := feature.NewFeature()
	sink := feature.MakeRandomK8sName("sink")
	cs := feature.MakeRandomK8sName("containersource")

	var channels []string
	for i := 0; i < length; i++ {
		name := feature.MakeRandomK8sName(fmt.Sprintf("channel-%04d", i))
		channels = append(channels, name)
		f.Setup("install channel", channel_impl.Install(name))
		f.Setup("channel is ready", channel_impl.IsReady(name))
	}

	f.Setup("install sink", eventshub.Install(sink, eventshub.StartReceiver))

	// use the rest for the chain
	for i := 0; i < length; i++ {
		sub := feature.MakeRandomK8sName(fmt.Sprintf("subscription-%04d", i))
		if i == length-1 {
			// install the final connection to the sink
			f.Setup("install sink subscription", subscription.Install(sub,
				subscription.WithChannel(channel_impl.AsRef(channels[i])),
				createSubscriberFn(svc.AsKReference(sink), ""),
			))
		} else {
			f.Setup("install subscription", subscription.Install(sub,
				subscription.WithChannel(channel_impl.AsRef(channels[i])),
				createSubscriberFn(channel_impl.AsRef(channels[i+1]), ""),
			))
		}

		f.Setup("subscription is ready", subscription.IsReady(sub))
	}

	// attach the first channel to the source
	f.Requirement("install containersource", containersource.Install(cs, containersource.WithSink(channel_impl.AsRef(channels[0]), "")))
	f.Requirement("containersource goes ready", containersource.IsReady(cs))

	f.Assert("chained channels relay events", assert.OnStore(sink).MatchEvent(test.HasType("dev.knative.eventing.samples.heartbeat")).AtLeast(1))

	return f
}

func DeadLetterSink(createSubscriberFn func(ref *duckv1.KReference, uri string) manifest.CfgFn) *feature.Feature {
	f := feature.NewFeature()
	sink := feature.MakeRandomK8sName("sink")
	failer := feature.MakeK8sNamePrefix("failer")
	cs := feature.MakeRandomK8sName("containersource")
	name := feature.MakeRandomK8sName("channel")

	f.Setup("install sink", eventshub.Install(sink, eventshub.StartReceiver))
	f.Setup("install failing receiver", eventshub.Install(failer, eventshub.StartReceiver, eventshub.DropFirstN(1)))
	f.Setup("install channel", channel_impl.Install(name, delivery.WithDeadLetterSink(svc.AsKReference(sink), "")))
	f.Setup("install containersource", containersource.Install(cs, source.WithSink(channel_impl.AsRef(name), "")))
	f.Setup("install subscription", subscription.Install(feature.MakeRandomK8sName("subscription"),
		subscription.WithChannel(channel_impl.AsRef(name)),
		createSubscriberFn(svc.AsKReference(failer), ""),
	))

	f.Setup("channel is ready", channel_impl.IsReady(name))
	f.Setup("containersource is ready", containersource.IsReady(cs))

	f.Requirement("Channel has dead letter sink uri", channel_impl.HasDeadLetterSinkURI(name, channel_impl.GVR()))

	f.Assert("dls receives events", assert.OnStore(sink).
		MatchEvent(test.HasType("dev.knative.eventing.samples.heartbeat")).
		AtLeast(1),
	)

	return f
}

func DeadLetterSinkGenericChannel(createSubscriberFn func(ref *duckv1.KReference, uri string) manifest.CfgFn) *feature.Feature {
	f := feature.NewFeature()
	sink := feature.MakeRandomK8sName("sink")
	failer := feature.MakeK8sNamePrefix("failer")
	cs := feature.MakeRandomK8sName("containersource")
	name := feature.MakeRandomK8sName("channel")

	f.Setup("install sink", eventshub.Install(sink, eventshub.StartReceiver))
	f.Setup("install failing receiver", eventshub.Install(failer, eventshub.StartReceiver, eventshub.DropFirstN(1)))
	f.Setup("install channel", channel.Install(name,
		channel.WithTemplate(),
		delivery.WithDeadLetterSink(svc.AsKReference(sink), "")),
	)
	f.Setup("install containersource", containersource.Install(cs, source.WithSink(channel.AsRef(name), "")))
	f.Setup("install subscription", subscription.Install(feature.MakeRandomK8sName("subscription"),
		subscription.WithChannel(channel.AsRef(name)),
		createSubscriberFn(svc.AsKReference(failer), ""),
	))

	f.Setup("channel is ready", channel.IsReady(name))
	f.Setup("containersource is ready", containersource.IsReady(cs))

	f.Requirement("Channel has dead letter sink uri", channel_impl.HasDeadLetterSinkURI(name, channel.GVR()))

	f.Assert("dls receives events", assert.OnStore(sink).
		MatchEvent(test.HasType("dev.knative.eventing.samples.heartbeat")).
		AtLeast(1),
	)

	return f
}

func AsDeadLetterSink(createSubscriberFn func(ref *duckv1.KReference, uri string) manifest.CfgFn) *feature.Feature {
	f := feature.NewFeatureNamed("As dead letter sink")

	cs := feature.MakeRandomK8sName("containersource")

	name := feature.MakeRandomK8sName("channel")
	dls := feature.MakeRandomK8sName("dls-channel")

	failer := feature.MakeRandomK8sName("failer")
	sink := feature.MakeRandomK8sName("sink")

	f.Setup("install containersource", containersource.Install(cs, source.WithSink(channel.AsRef(name), "")))

	f.Setup("install channel", channel.Install(name,
		channel.WithTemplate(),
		delivery.WithDeadLetterSink(channel.AsRef(dls), "")),
	)
	f.Setup("install subscription", subscription.Install(feature.MakeRandomK8sName("subscription"),
		subscription.WithChannel(channel.AsRef(name)),
		createSubscriberFn(svc.AsKReference(failer), ""),
	))

	f.Setup("install DLS channel", channel.Install(dls,
		channel.WithTemplate(),
	))
	f.Setup("install DLS subscription", subscription.Install(feature.MakeRandomK8sName("dls-subscription"),
		subscription.WithChannel(channel.AsRef(dls)),
		createSubscriberFn(svc.AsKReference(sink), ""),
	))

	f.Setup("install sink", eventshub.Install(sink, eventshub.StartReceiver))
	f.Setup("install failing receiver", eventshub.Install(failer, eventshub.StartReceiver, eventshub.DropFirstN(10)))

	f.Setup("channel is ready", channel.IsReady(name))
	f.Setup("channel is ready", channel.IsReady(dls))
	f.Setup("containersource is ready", containersource.IsReady(cs))

	f.Requirement("Channel has dead letter sink uri", channel_impl.HasDeadLetterSinkURI(name, channel.GVR()))

	f.Assert("dls receives events", assert.OnStore(sink).
		MatchEvent(test.HasType("dev.knative.eventing.samples.heartbeat")).
		AtLeast(1),
	)

	return f
}

func EventTransformation() *feature.Feature {
	f := feature.NewFeature()
	channel1 := feature.MakeRandomK8sName("channel 1")
	channel2 := feature.MakeRandomK8sName("channel 2")
	subscription1 := feature.MakeRandomK8sName("subscription 1")
	subscription2 := feature.MakeRandomK8sName("subscription 2")
	prober := eventshub.NewProber()
	prober.SetTargetResource(channel_impl.GVR(), channel1)
	prober.SenderFullEvents(3)

	f.Setup("install sink", prober.ReceiverInstall("sink"))
	f.Setup("install transform service", prober.ReceiverInstall("transform", eventshub.ReplyWithTransformedEvent("transformed", "transformer", "")))
	f.Setup("install channel 1", channel_impl.Install(channel1))
	f.Setup("install channel 2", channel_impl.Install(channel2))
	f.Setup("install subscription 1", subscription.Install(subscription1,
		subscription.WithChannel(channel_impl.AsRef(channel1)),
		subscription.WithSubscriber(prober.AsKReference("transform"), ""),
		subscription.WithReply(channel_impl.AsRef(channel2), ""),
	))
	f.Setup("install subscription 2", subscription.Install(subscription2,
		subscription.WithChannel(channel_impl.AsRef(channel2)),
		subscription.WithSubscriber(prober.AsKReference("sink"), ""),
	))
	f.Setup("subscription 1 is ready", subscription.IsReady(subscription1))
	f.Setup("subscription 2 is ready", subscription.IsReady(subscription2))
	f.Setup("channel 1 is ready", channel_impl.IsReady(channel1))
	f.Setup("channel 2 is ready", channel_impl.IsReady(channel2))
	f.Setup("install source", prober.SenderInstall("source"))

	f.Requirement("sender is finished", prober.SenderDone("source"))
	f.Requirement("receiver is finished", prober.ReceiverDone("source", "sink"))

	f.Assert("sink receives events", prober.AssertReceivedAll("source", "sink"))
	f.Assert("events have passed through transform service", func(ctx context.Context, t feature.T) {
		events := prober.ReceivedBy(ctx, "sink")
		if len(events) != 3 {
			t.Errorf("expected 3 events, got %d", len(events))
		}
		for _, e := range events {
			if e.Event.Type() != "transformed" {
				t.Errorf(`expected event type to be "transformed", got %q`, e.Event.Type())
			}
		}
	})
	return f
}

func SingleEventWithEncoding(encoding binding.Encoding) *feature.Feature {
	f := feature.NewFeature()
	channel := feature.MakeRandomK8sName("channel")
	sub := feature.MakeRandomK8sName("subscription")
	prober := eventshub.NewProber()
	prober.SetTargetResource(channel_impl.GVR(), channel)

	event := cloudevents.NewEvent()
	event.SetID(feature.MakeRandomK8sName("test"))
	event.SetType("myevent")
	event.SetSource("http://sender.svc/")
	prober.ExpectEvents([]string{event.ID()})

	f.Setup("install sink", prober.ReceiverInstall("sink"))
	f.Setup("install channel", channel_impl.Install(channel))
	f.Setup("install subscription", subscription.Install(sub,
		subscription.WithChannel(channel_impl.AsRef(channel)),
		subscription.WithSubscriber(prober.AsKReference("sink"), ""),
	))

	f.Setup("subscription is ready", subscription.IsReady(sub))
	f.Setup("channel is ready", channel_impl.IsReady(channel))
	f.Setup("install source", prober.SenderInstall("source", eventshub.InputEventWithEncoding(event, encoding)))

	f.Requirement("sender is finished", prober.SenderDone("source"))
	f.Requirement("receiver is finished", prober.ReceiverDone("source", "sink"))

	f.Assert("sink receives events", prober.AssertReceivedAll("source", "sink"))

	return f
}

func ChannelPreferHeaderCheck(createSubscriberFn func(ref *duckv1.KReference, uri string) manifest.CfgFn) *feature.Feature {
	f := feature.NewFeatureNamed("Channel PreferHeader Check")

	channelName := feature.MakeRandomK8sName("channel")
	sub := feature.MakeRandomK8sName("subscription")
	source := feature.MakeRandomK8sName("source")
	sink := feature.MakeRandomK8sName("sink")

	eventSource := "source1"
	eventType := "type1"
	eventBody := `{"msg":"test msg"}`
	event := cloudevents.NewEvent()
	event.SetID(uuid.New().String())
	event.SetType(eventType)
	event.SetSource(eventSource)
	event.SetData(cloudevents.ApplicationJSON, []byte(eventBody))

	f.Setup("install sink", eventshub.Install(sink, eventshub.StartReceiver))
	f.Setup("install channel", channel.Install(channelName,
		channel.WithTemplate(),
	))
	f.Setup("install subscription", subscription.Install(sub,
		subscription.WithChannel(channel.AsRef(channelName)),
		createSubscriberFn(svc.AsKReference(sink), ""),
	))

	f.Setup("subscription is ready", subscription.IsReady(sub))
	f.Setup("channel is ready", channel.IsReady(channelName))

	f.Requirement("install source", eventshub.Install(
		source,
		eventshub.StartSenderToResource(channel.GVR(), channelName),
		eventshub.InputEvent(event),
	))

	f.Stable("test message without explicit prefer header should have the header").
		Must("delivers events",
			eventasssert.OnStore(sink).Match(
				eventasssert.HasAdditionalHeader("Prefer", "reply"),
			).AtLeast(1))

	return f
}