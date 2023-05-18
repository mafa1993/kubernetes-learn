/*
Copyright The Kubernetes Authors.

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

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1alpha1 "app-controller/pkg/apis/appcontroller/v1alpha1"
	"context"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeApps implements AppInterface
type FakeApps struct {
	Fake *FakeAppcontrollerV1alpha1
	ns   string
}

var appsResource = v1alpha1.SchemeGroupVersion.WithResource("apps")

var appsKind = v1alpha1.SchemeGroupVersion.WithKind("App")

// Get takes name of the app, and returns the corresponding app object, and an error if there is any.
func (c *FakeApps) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.App, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(appsResource, c.ns, name), &v1alpha1.App{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.App), err
}

// List takes label and field selectors, and returns the list of Apps that match those selectors.
func (c *FakeApps) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.AppList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(appsResource, appsKind, c.ns, opts), &v1alpha1.AppList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.AppList{ListMeta: obj.(*v1alpha1.AppList).ListMeta}
	for _, item := range obj.(*v1alpha1.AppList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested apps.
func (c *FakeApps) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(appsResource, c.ns, opts))

}

// Create takes the representation of a app and creates it.  Returns the server's representation of the app, and an error, if there is any.
func (c *FakeApps) Create(ctx context.Context, app *v1alpha1.App, opts v1.CreateOptions) (result *v1alpha1.App, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(appsResource, c.ns, app), &v1alpha1.App{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.App), err
}

// Update takes the representation of a app and updates it. Returns the server's representation of the app, and an error, if there is any.
func (c *FakeApps) Update(ctx context.Context, app *v1alpha1.App, opts v1.UpdateOptions) (result *v1alpha1.App, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(appsResource, c.ns, app), &v1alpha1.App{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.App), err
}

// Delete takes name of the app and deletes it. Returns an error if one occurs.
func (c *FakeApps) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(appsResource, c.ns, name, opts), &v1alpha1.App{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeApps) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(appsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.AppList{})
	return err
}

// Patch applies the patch and returns the patched app.
func (c *FakeApps) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.App, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(appsResource, c.ns, name, pt, data, subresources...), &v1alpha1.App{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.App), err
}
