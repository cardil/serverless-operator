package chaos

import (
	"sync"
	"time"

	"github.com/openshift-knative/serverless-operator/test"
	"gotest.tools/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kafkasourcev1beta1 "knative.dev/eventing-kafka/pkg/apis/sources/v1beta1"
)

func NewKafkaSourceDuck() Duck {
	return &donald{}
}

// see: https://upload.wikimedia.org/wikipedia/en/a/a5/Donald_Duck_angry_transparent_background.png
type donald struct {
	*stateCtx
	synchronized
}

func (d *donald) Quak() {
	t := d.tb
	testCtx := test.SetupClusterAdmin(t)
	var wg *sync.WaitGroup
	defer func() {
		testCtx.Cleanup(t)
		d.log.Info("Clean up done.")
		if wg != nil {
			d.log.Info("Done, sir.")
			wg.Done()
		}
	}()

	ns := findNamespacesByPrefix(testCtx, d.stateCtx,
		d.cfg.KafkaSource.NamespacePrefixes...)
	d.ready().Done()
	ch := d.finished()
	for {
		select {
		case wg = <-ch:
			d.log.Info("someone told me to stop")
			return
		default:
			d.quak(testCtx, ns)
		}
	}
}

func (d *donald) takenCareOf(sctx *stateCtx, synch synchronized) Duck {
	return &donald{
		stateCtx:     sctx.named("Donald"),
		synchronized: synch,
	}
}

func (d *donald) quak(testCtx *test.Context, namespaces []string) {
	d.log.Infof("Looking at Kafka sources in namespaces: %v", namespaces)
	victims := make([]kafkasourcev1beta1.KafkaSource, 0)
	for _, ns := range namespaces {
		victims = append(victims, d.lookupKafkaSourcesInNamespace(ns, testCtx)...)
	}
	d.log.Infof("Found %d Kafka source victims to quak at", len(victims))
	for _, kafkaSource := range victims {
		d.quakAt(kafkaSource, testCtx)
	}
	d.log.Info("Done with the quaking. Doing to sleep for now...")
	time.Sleep(d.cfg.Wait)
}

func (d *donald) quakAt(
	source kafkasourcev1beta1.KafkaSource,
	testCtx *test.Context,
) {
	d.log.Infof("Quak at %s in namespace %s",
		source.GetName(), source.GetNamespace())
	podSelector := source.Status.Selector
	pods := testCtx.Clients.Kube.CoreV1().Pods(source.GetNamespace())
	err := pods.DeleteCollection(d.ctx, metav1.DeleteOptions{}, metav1.ListOptions{
		LabelSelector: podSelector,
	})
	assert.NilError(d.tb, err)
}

func (d *donald) lookupKafkaSourcesInNamespace(
	namespace string,
	testCtx *test.Context,
) []kafkasourcev1beta1.KafkaSource {
	kafkasources := testCtx.Clients.Kafka.SourcesV1beta1().KafkaSources(namespace)
	kss, err := kafkasources.List(d.ctx, metav1.ListOptions{})
	assert.NilError(d.tb, err)
	return kss.Items
}
