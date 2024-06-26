// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	v1 "github.com/openshift/api/config/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeAuthentications implements AuthenticationInterface
type FakeAuthentications struct {
	Fake *FakeConfigV1
}

var authenticationsResource = v1.SchemeGroupVersion.WithResource("authentications")

var authenticationsKind = v1.SchemeGroupVersion.WithKind("Authentication")

// Get takes name of the authentication, and returns the corresponding authentication object, and an error if there is any.
func (c *FakeAuthentications) Get(ctx context.Context, name string, options metav1.GetOptions) (result *v1.Authentication, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootGetAction(authenticationsResource, name), &v1.Authentication{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1.Authentication), err
}

// List takes label and field selectors, and returns the list of Authentications that match those selectors.
func (c *FakeAuthentications) List(ctx context.Context, opts metav1.ListOptions) (result *v1.AuthenticationList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootListAction(authenticationsResource, authenticationsKind, opts), &v1.AuthenticationList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1.AuthenticationList{ListMeta: obj.(*v1.AuthenticationList).ListMeta}
	for _, item := range obj.(*v1.AuthenticationList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested authentications.
func (c *FakeAuthentications) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchAction(authenticationsResource, opts))
}

// Create takes the representation of a authentication and creates it.  Returns the server's representation of the authentication, and an error, if there is any.
func (c *FakeAuthentications) Create(ctx context.Context, authentication *v1.Authentication, opts metav1.CreateOptions) (result *v1.Authentication, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateAction(authenticationsResource, authentication), &v1.Authentication{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1.Authentication), err
}

// Update takes the representation of a authentication and updates it. Returns the server's representation of the authentication, and an error, if there is any.
func (c *FakeAuthentications) Update(ctx context.Context, authentication *v1.Authentication, opts metav1.UpdateOptions) (result *v1.Authentication, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateAction(authenticationsResource, authentication), &v1.Authentication{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1.Authentication), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeAuthentications) UpdateStatus(ctx context.Context, authentication *v1.Authentication, opts metav1.UpdateOptions) (*v1.Authentication, error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateSubresourceAction(authenticationsResource, "status", authentication), &v1.Authentication{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1.Authentication), err
}

// Delete takes name of the authentication and deletes it. Returns an error if one occurs.
func (c *FakeAuthentications) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteActionWithOptions(authenticationsResource, name, opts), &v1.Authentication{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeAuthentications) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	action := testing.NewRootDeleteCollectionAction(authenticationsResource, listOpts)

	_, err := c.Fake.Invokes(action, &v1.AuthenticationList{})
	return err
}

// Patch applies the patch and returns the patched authentication.
func (c *FakeAuthentications) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.Authentication, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(authenticationsResource, name, pt, data, subresources...), &v1.Authentication{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1.Authentication), err
}
